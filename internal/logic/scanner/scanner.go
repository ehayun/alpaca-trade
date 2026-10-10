// Package scanner finds index stocks whose EMA 20, 100 and 200 are close together.
package scanner

import (
	"bufio"
	"context"
	"embed"
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

// Start begins a scan in the background and returns at once; if a scan is running, it changes nothing.
func Start(maxSpreadPct float64) model.ScanStatus {
	mu.Lock()
	defer mu.Unlock()
	if status.Running {
		return status
	}
	symbols := Symbols()
	now := time.Now()
	status = model.ScanStatus{
		Running: true, StartedAt: &now, Total: len(symbols), MaxSpreadPct: maxSpreadPct, Matches: []model.ScanMatch{},
	}
	go run(symbols, maxSpreadPct)
	return status
}

func run(symbols []string, maxSpreadPct float64) {
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
			m, ok := Evaluate(s, candles[s])
			if !ok {
				continue
			}
			scanned++
			if m.SpreadPct <= maxSpreadPct {
				matches = append(matches, m)
			}
		}
		mu.Lock()
		status.Done = i + len(chunk)
		mu.Unlock()
	}
	slices.SortFunc(matches, func(a, b model.ScanMatch) int {
		if a.SpreadPct < b.SpreadPct {
			return -1
		}
		if a.SpreadPct > b.SpreadPct {
			return 1
		}
		return strings.Compare(a.Symbol, b.Symbol)
	})
}

// Evaluate computes the last EMAs and their spread; ok is false without enough history for EMA 200.
func Evaluate(symbol string, candles []model.Candle) (m model.ScanMatch, ok bool) {
	if len(candles) < 200 {
		return m, false
	}
	last := candles[len(candles)-1].Close
	if last <= 0 {
		return m, false
	}
	e20, e100, e200 := lastEMA(candles, 20), lastEMA(candles, 100), lastEMA(candles, 200)
	hi, lo := max(e20, e100, e200), min(e20, e100, e200)
	return model.ScanMatch{
		Symbol: symbol, Close: last, EMA20: e20, EMA100: e100, EMA200: e200, SpreadPct: (hi - lo) / last * 100,
	}, true
}

func lastEMA(candles []model.Candle, period int) float64 {
	pts := market.EMA(candles, period)
	return pts[len(pts)-1].Value
}
