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

// emaWarmupDays of extra history (~1030 trading days) let EMA 200 settle before the first shown day:
// the starting value's weight drops to (1-2/201)^1030 ≈ 0.003%, so values match a full-history EMA.
const emaWarmupDays = 1500

// DailyCandles returns daily candles for the last `days` calendar days, with their EMAs.
func DailyCandles(ctx context.Context, symbol string, days int) ([]model.Candle, []model.EMA, error) {
	start, end := Range(time.Now(), days)
	bars, err := client.GetBars(symbol, marketdata.GetBarsRequest{
		TimeFrame:  marketdata.OneDay,
		Adjustment: marketdata.AdjustmentSplit,
		Start:      start.AddDate(0, 0, -emaWarmupDays),
		End:        end,
	})
	if err != nil {
		return nil, nil, err
	}
	candles, emas := WithEMAs(ToCandles(bars), start.In(newYork).Format(time.DateOnly), EMAPeriods...)
	return candles, emas, nil
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
		emas[i] = model.EMA{Period: p, Points: []model.Point{}}
		for _, pt := range EMA(all, p) {
			if pt.Time >= from {
				emas[i].Points = append(emas[i].Points, pt)
			}
		}
	}
	return all[first:], emas
}

// EMA of the closes, seeded with the simple average of the first `period` closes.
// It has no points when there are fewer than `period` candles.
func EMA(candles []model.Candle, period int) []model.Point {
	if period < 1 || len(candles) < period {
		return nil
	}
	sum := 0.0
	for _, c := range candles[:period] {
		sum += c.Close
	}
	ema := sum / float64(period)
	k := 2 / float64(period+1)
	out := []model.Point{{Time: candles[period-1].Time, Value: ema}}
	for _, c := range candles[period:] {
		ema = c.Close*k + ema*(1-k)
		out = append(out, model.Point{Time: c.Time, Value: ema})
	}
	return out
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
