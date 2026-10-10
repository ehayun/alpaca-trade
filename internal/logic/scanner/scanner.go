// Package scanner finds index stocks whose EMA 20, 100 and 200 are close together.
package scanner

import (
	"bufio"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"alpaca-app/internal/logic/market"
	"alpaca-app/internal/model"
)

// Index components: S&P 500, Nasdaq-100, Dow 30. Update them by hand when an index changes.
//
//go:embed symbols/*.txt
var lists embed.FS

const (
	historyDays = 1000 // ~690 trading days: EMA 200 settles to within a fraction of a percent
	chunkSize   = 100  // symbols per Alpaca request; the SDK pages through the bars
)

var (
	mu     sync.Mutex
	status = model.ScanStatus{Matches: []model.ScanMatch{}}
)

// Symbols returns the unique symbols of all index lists, sorted.
func Symbols() []string {
	files, _ := lists.ReadDir("symbols")
	var out []string
	for _, f := range files {
		data, err := lists.ReadFile("symbols/" + f.Name())
		if err != nil {
			panic(err) // embedded at build time; cannot fail
		}
		sc := bufio.NewScanner(strings.NewReader(string(data)))
		for sc.Scan() {
			if s := strings.TrimSpace(sc.Text()); s != "" && !strings.HasPrefix(s, "#") {
				out = append(out, s)
			}
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}

// Status returns the current scan state. Matches is replaced, never changed in place, so sharing it is safe.
func Status() model.ScanStatus {
	mu.Lock()
	defer mu.Unlock()
	return status
}

// Stochastic rules a scan can add to the EMA rule (with the slow stochastic 14, 3, 3).
const (
	StochOversold   = "oversold"   // %K <= 20
	StochOverbought = "overbought" // %K >= 80
	StochBullCross  = "bullcross"  // in the last 3 days, %K crossed above %D from 20 or less
)

// Options of one scan.
type Options struct {
	MaxSpreadPct float64
	StochRule    string // "" = do not use the stochastic
	PriceClose   bool   // the close must also be inside the limit (spread of EMAs + close)
	ResultFile   string // a successful result is written here
}

// Start begins a scan in the background and returns at once; if a scan is running, it changes nothing.
func Start(opts Options) model.ScanStatus {
	mu.Lock()
	defer mu.Unlock()
	if status.Running {
		return status
	}
	symbols := Symbols()
	now := time.Now()
	status = model.ScanStatus{
		Running: true, StartedAt: &now, Total: len(symbols),
		MaxSpreadPct: opts.MaxSpreadPct, StochRule: opts.StochRule, PriceClose: opts.PriceClose, Matches: []model.ScanMatch{},
	}
	go run(symbols, opts)
	return status
}

func run(symbols []string, opts Options) {
	ctx := context.Background()
	matches, scanned, failed := []model.ScanMatch{}, 0, false
	defer func() {
		if r := recover(); r != nil {
			g.Log().Errorf(ctx, "scan panic: %v", r)
			failed = true
		}
		now := time.Now()
		mu.Lock()
		defer mu.Unlock()
		status.Running, status.Failed, status.FinishedAt = false, failed, &now
		status.Scanned = scanned
		if !failed {
			status.Matches = matches
			if err := Save(opts.ResultFile, status); err != nil {
				g.Log().Errorf(ctx, "save scan: %+v", err)
			}
		}
	}()

	for i := 0; i < len(symbols); i += chunkSize {
		chunk := symbols[i:min(i+chunkSize, len(symbols))]
		candles, err := market.DailyCandlesMulti(ctx, chunk, historyDays)
		if err != nil {
			g.Log().Errorf(ctx, "scan %s..%s: %+v", chunk[0], chunk[len(chunk)-1], err)
			failed = true
			return
		}
		for _, s := range chunk {
			m, ok, stochOK := Evaluate(s, candles[s], opts.StochRule)
			if !ok {
				continue
			}
			scanned++
			if m.Spread(opts.PriceClose) <= opts.MaxSpreadPct && stochOK {
				matches = append(matches, m)
			}
		}
		mu.Lock()
		status.Done = i + len(chunk)
		mu.Unlock()
	}
	slices.SortFunc(matches, func(a, b model.ScanMatch) int {
		sa, sb := a.Spread(opts.PriceClose), b.Spread(opts.PriceClose)
		if sa < sb {
			return -1
		}
		if sa > sb {
			return 1
		}
		return strings.Compare(a.Symbol, b.Symbol)
	})
}

// Save writes a finished scan to path as JSON. It writes a temp file and renames it,
// so a reader never sees half a file.
func Save(path string, s model.ScanStatus) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Load reads the scan saved at path; it returns nil, nil when there is no saved scan.
func Load(path string) (*model.ScanStatus, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var s model.ScanStatus
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Evaluate computes the last EMAs, their spread and the stochastic; ok is false without enough history
// for EMA 200. stochOK tells whether the symbol passes stochRule ("" always passes).
func Evaluate(symbol string, candles []model.Candle, stochRule string) (m model.ScanMatch, ok, stochOK bool) {
	if len(candles) < 200 {
		return m, false, false
	}
	last := candles[len(candles)-1].Close
	if last <= 0 {
		return m, false, false
	}
	e20, e100, e200 := lastEMA(candles, 20), lastEMA(candles, 100), lastEMA(candles, 200)
	hi, lo := max(e20, e100, e200), min(e20, e100, e200)
	k, d := market.Stochastic(candles, market.StochPeriod, market.StochSmooth, market.StochD)
	return model.ScanMatch{
		Symbol: symbol, Close: last, EMA20: e20, EMA100: e100, EMA200: e200,
		SpreadPct:      (hi - lo) / last * 100,
		PriceSpreadPct: (max(hi, last) - min(lo, last)) / last * 100,
		StochK:         k[len(k)-1].Value, StochD: d[len(d)-1].Value,
	}, true, StochPass(stochRule, k, d)
}

// StochPass checks one stochastic rule on %K and %D (each ends on the same, latest day).
func StochPass(rule string, k, d []model.Point) bool {
	if rule == "" {
		return true
	}
	if len(k) == 0 || len(d) < 2 {
		return false
	}
	switch last := k[len(k)-1].Value; rule {
	case StochOversold:
		return last <= 20
	case StochOverbought:
		return last >= 80
	case StochBullCross:
		k = k[len(k)-len(d):] // align %K with %D, which starts later
		for i := max(1, len(d)-3); i < len(d); i++ {
			if k[i-1].Value <= d[i-1].Value && k[i].Value > d[i].Value && k[i-1].Value <= 20 {
				return true
			}
		}
	}
	return false
}

func lastEMA(candles []model.Candle, period int) float64 {
	pts := market.EMA(candles, period)
	return pts[len(pts)-1].Value
}
