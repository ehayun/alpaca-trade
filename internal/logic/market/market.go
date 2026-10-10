// Package market fetches prices and news from Alpaca market data.
package market

import (
	"context"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"

	"alpaca-app/internal/model"
)

// Reads APCA_API_KEY_ID / APCA_API_SECRET_KEY from the environment.
var client = marketdata.NewClient(marketdata.ClientOpts{})

var newYork = mustLoad("America/New_York")

// EMAPeriods are the moving averages drawn on the chart.
var EMAPeriods = []int{20, 100, 200}

// MACD settings (the usual defaults): EMA 12 - EMA 26, signal = EMA 9 of the MACD line.
const (
	MACDFast   = 12
	MACDSlow   = 26
	MACDSignal = 9
)

// emaWarmupDays of extra history (~1030 trading days) let EMA 200 settle before the first shown day:
// the starting value's weight drops to (1-2/201)^1030 ≈ 0.003%, so values match a full-history EMA.
const emaWarmupDays = 1500

// DailyCandles returns daily candles for the last `days` calendar days, with their EMAs and MACD.
func DailyCandles(ctx context.Context, symbol string, days int) ([]model.Candle, []model.EMA, model.MACD, error) {
	start, end := Range(time.Now(), days)
	bars, err := client.GetBars(symbol, marketdata.GetBarsRequest{
		TimeFrame:  marketdata.OneDay,
		Adjustment: marketdata.AdjustmentSplit,
		Start:      start.AddDate(0, 0, -emaWarmupDays),
		End:        end,
	})
	if err != nil {
		return nil, nil, model.MACD{}, err
	}
	all := ToCandles(bars)
	from := start.In(newYork).Format(time.DateOnly)
	candles, emas := WithEMAs(all, from, EMAPeriods...)
	line, signal, hist := MACD(all, MACDFast, MACDSlow, MACDSignal)
	return candles, emas, model.MACD{MACD: since(line, from), Signal: since(signal, from), Hist: since(hist, from)}, nil
}

// DailyCandlesMulti returns daily candles for the last `days` calendar days for many symbols at once.
// Symbols without data are missing from the map.
func DailyCandlesMulti(ctx context.Context, symbols []string, days int) (map[string][]model.Candle, error) {
	start, end := Range(time.Now(), days)
	bars, err := client.GetMultiBars(symbols, marketdata.GetBarsRequest{
		TimeFrame:  marketdata.OneDay,
		Adjustment: marketdata.AdjustmentSplit,
		Start:      start,
		End:        end,
	})
	if err != nil {
		return nil, err
	}
	out := make(map[string][]model.Candle, len(bars))
	for s, b := range bars {
		out[s] = ToCandles(b)
	}
	return out, nil
}

// WithEMAs computes EMAs over all candles, then keeps only the days from `from` (YYYY-MM-DD) on.
func WithEMAs(all []model.Candle, from string, periods ...int) ([]model.Candle, []model.EMA) {
	first := len(all)
	for i, c := range all {
		if c.Time >= from {
			first = i
			break
		}
	}
	emas := make([]model.EMA, len(periods))
	for i, p := range periods {
		emas[i] = model.EMA{Period: p, Points: since(EMA(all, p), from)}
	}
	return all[first:], emas
}

// since keeps the points from day `from` (YYYY-MM-DD) on; never nil, so it marshals as [].
func since(pts []model.Point, from string) []model.Point {
	out := []model.Point{}
	for _, p := range pts {
		if p.Time >= from {
			out = append(out, p)
		}
	}
	return out
}

// EMA of the closes, seeded with the simple average of the first `period` closes.
// It has no points when there are fewer than `period` candles.
func EMA(candles []model.Candle, period int) []model.Point {
	closes := make([]model.Point, len(candles))
	for i, c := range candles {
		closes[i] = model.Point{Time: c.Time, Value: c.Close}
	}
	return emaOf(closes, period)
}

// emaOf is the EMA of point values, seeded with the simple average of the first `period` values.
func emaOf(pts []model.Point, period int) []model.Point {
	if period < 1 || len(pts) < period {
		return nil
	}
	sum := 0.0
	for _, p := range pts[:period] {
		sum += p.Value
	}
	ema := sum / float64(period)
	k := 2 / float64(period+1)
	out := []model.Point{{Time: pts[period-1].Time, Value: ema}}
	for _, p := range pts[period:] {
		ema = p.Value*k + ema*(1-k)
		out = append(out, model.Point{Time: p.Time, Value: ema})
	}
	return out
}

// MACD returns the MACD line (fast EMA - slow EMA, from the day the slow EMA starts),
// the signal line (EMA of the MACD line) and the histogram (MACD - signal, from the day the signal starts).
func MACD(candles []model.Candle, fast, slow, signal int) (line, sig, hist []model.Point) {
	f, s := EMA(candles, fast), EMA(candles, slow)
	if len(s) == 0 || len(f) < len(s) {
		return nil, nil, nil
	}
	f = f[len(f)-len(s):] // both now end on the last candle and start on the slow EMA's first day
	line = make([]model.Point, len(s))
	for i := range s {
		line[i] = model.Point{Time: s[i].Time, Value: f[i].Value - s[i].Value}
	}
	sig = emaOf(line, signal)
	hist = make([]model.Point, len(sig))
	for i, p := range sig {
		hist[i] = model.Point{Time: p.Time, Value: line[len(line)-len(sig)+i].Value - p.Value}
	}
	return line, sig, hist
}

// Range ends 16 minutes before now: free plans can't query the latest 15 minutes of SIP data.
func Range(now time.Time, days int) (start, end time.Time) {
	end = now.Add(-16 * time.Minute)
	return end.AddDate(0, 0, -days), end
}

// ToCandles converts SDK bars to chart candles.
func ToCandles(bars []marketdata.Bar) []model.Candle {
	out := make([]model.Candle, len(bars))
	for i, b := range bars {
		out[i] = model.Candle{
			Time:   b.Timestamp.In(newYork).Format(time.DateOnly),
			Open:   b.Open,
			High:   b.High,
			Low:    b.Low,
			Close:  b.Close,
			Volume: b.Volume,
		}
	}
	return out
}

// Quote returns today's price data for the info tab.
func Quote(ctx context.Context, symbol string) (model.Quote, error) {
	s, err := client.GetSnapshot(symbol, marketdata.GetSnapshotRequest{})
	if err != nil {
		return model.Quote{}, err
	}
	return ToQuote(s), nil
}

// LatestPrices returns the latest price per symbol in one request; symbols without data are missing.
func LatestPrices(ctx context.Context, symbols []string) (map[string]float64, error) {
	snaps, err := client.GetSnapshots(symbols, marketdata.GetSnapshotRequest{})
	if err != nil {
		return nil, err
	}
	out := make(map[string]float64, len(snaps))
	for s, snap := range snaps {
		if p := ToQuote(snap).Price; p != nil {
			out[s] = *p
		}
	}
	return out, nil
}

// ToQuote takes the latest trade as the price (daily close if none) and compares it to the previous close.
func ToQuote(s *marketdata.Snapshot) model.Quote {
	var q model.Quote
	if s == nil {
		return q
	}
	if d := s.DailyBar; d != nil {
		q.Price, q.Open, q.High, q.Low, q.Volume = &d.Close, &d.Open, &d.High, &d.Low, &d.Volume
	}
	if t := s.LatestTrade; t != nil {
		q.Price = &t.Price
	}
	if p := s.PrevDailyBar; p != nil && p.Close != 0 {
		q.PrevClose = &p.Close
		if q.Price != nil {
			change := *q.Price - p.Close
			pct := change / p.Close * 100
			q.Change, q.ChangePct = &change, &pct
		}
	}
	return q
}

// News returns the `limit` most recent articles for the symbol.
func News(ctx context.Context, symbol string, limit int) ([]model.News, error) {
	ns, err := client.GetNews(marketdata.GetNewsRequest{
		Symbols:    []string{symbol},
		Sort:       marketdata.SortDesc,
		TotalLimit: limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]model.News, len(ns))
	for i, n := range ns {
		out[i] = model.News{Headline: n.Headline, Summary: n.Summary, URL: n.URL, CreatedAt: n.CreatedAt}
	}
	return out, nil
}

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
