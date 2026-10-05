// Package market fetches historical daily prices from Alpaca market data.
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

// DailyCandles returns daily candles for the last `days` calendar days.
func DailyCandles(ctx context.Context, symbol string, days int) ([]model.Candle, error) {
	start, end := Range(time.Now(), days)
	bars, err := client.GetBars(symbol, marketdata.GetBarsRequest{
		TimeFrame: marketdata.OneDay,
		Start:     start,
		End:       end,
	})
	if err != nil {
		return nil, err
	}
	return ToCandles(bars), nil
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

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
