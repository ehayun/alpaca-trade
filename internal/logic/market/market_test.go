package market

import (
	"testing"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"

	"alpaca-app/internal/model"
)

func TestRange(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	start, end := Range(now, 30)
	if want := now.Add(-16 * time.Minute); !end.Equal(want) {
		t.Fatalf("end = %v, want %v", end, want)
	}
	if want := end.AddDate(0, 0, -30); !start.Equal(want) {
		t.Fatalf("start = %v, want %v", start, want)
	}
}

func TestToCandles(t *testing.T) {
	// Alpaca stamps daily bars at 04:00 UTC (midnight New York); the date must stay the NY trading day.
	ts := time.Date(2026, 9, 8, 4, 0, 0, 0, time.UTC)
	got := ToCandles([]marketdata.Bar{{Timestamp: ts, Open: 1, High: 3, Low: 0.5, Close: 2, Volume: 10}})
	want := model.Candle{Time: "2026-09-08", Open: 1, High: 3, Low: 0.5, Close: 2, Volume: 10}
	if len(got) != 1 || got[0] != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
