package market

import (
	"reflect"
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

func TestToQuote(t *testing.T) {
	q := ToQuote(&marketdata.Snapshot{
		LatestTrade:  &marketdata.Trade{Price: 110},
		DailyBar:     &marketdata.Bar{Open: 101, High: 112, Low: 99, Close: 108, Volume: 5},
		PrevDailyBar: &marketdata.Bar{Close: 100},
	})
	if *q.Price != 110 || *q.PrevClose != 100 || *q.Change != 10 || *q.ChangePct != 10 || *q.Open != 101 || *q.Volume != 5 {
		t.Fatalf("got %+v", q)
	}

	// No latest trade: the daily close is the price.
	q = ToQuote(&marketdata.Snapshot{DailyBar: &marketdata.Bar{Close: 95}, PrevDailyBar: &marketdata.Bar{Close: 100}})
	if *q.Price != 95 || *q.Change != -5 || *q.ChangePct != -5 {
		t.Fatalf("got %+v", q)
	}

	// No previous day (new listing) or no snapshot: no change, no panic.
	if q = ToQuote(&marketdata.Snapshot{LatestTrade: &marketdata.Trade{Price: 1}}); q.Change != nil || q.PrevClose != nil {
		t.Fatalf("got %+v", q)
	}
	if q = ToQuote(nil); q.Price != nil {
		t.Fatalf("got %+v", q)
	}
}

func TestEMA(t *testing.T) {
	c := func(day string, close float64) model.Candle { return model.Candle{Time: day, Close: close} }
	all := []model.Candle{c("2026-01-01", 1), c("2026-01-02", 2), c("2026-01-03", 3), c("2026-01-04", 6)}

	// Period 3: seed = (1+2+3)/3 = 2 on day 3; then k = 0.5: 6*0.5 + 2*0.5 = 4.
	want := []model.Point{{Time: "2026-01-03", Value: 2}, {Time: "2026-01-04", Value: 4}}
	if got := EMA(all, 3); !reflect.DeepEqual(got, want) {
		t.Fatalf("EMA = %+v, want %+v", got, want)
	}
	if got := EMA(all, 5); got != nil {
		t.Fatalf("too few candles: got %+v, want nil", got)
	}

	// Warm-up days are used for the EMA but cut from the result.
	candles, emas := WithEMAs(all, "2026-01-04", 3, 5)
	if len(candles) != 1 || candles[0].Time != "2026-01-04" {
		t.Fatalf("candles = %+v", candles)
	}
	if len(emas) != 2 || !reflect.DeepEqual(emas[0].Points, want[1:]) || len(emas[1].Points) != 0 {
		t.Fatalf("emas = %+v", emas)
	}
}

func TestStochastic(t *testing.T) {
	c := func(day string, high, low, close float64) model.Candle {
		return model.Candle{Time: day, High: high, Low: low, Close: close}
	}
	all := []model.Candle{
		c("d1", 10, 0, 5),   // range 0..10
		c("d2", 10, 0, 10),  // raw (period 2): (10-0)/(10-0) = 100
		c("d3", 10, 0, 0),   // raw 0
		c("d4", 20, 10, 20), // range 0..20 -> raw 100
		c("d5", 20, 20, 20), // range 10..20 -> raw 100
		c("d6", 7, 7, 7),    // range 7..20 -> raw 0
	}
	// period 2, smooth 1 => %K = raw; dPeriod 2 => %D = 2-day average of %K.
	k, d := Stochastic(all, 2, 1, 2)
	p := func(day string, v float64) model.Point { return model.Point{Time: day, Value: v} }
	wantK := []model.Point{p("d2", 100), p("d3", 0), p("d4", 100), p("d5", 100), p("d6", 0)}
	wantD := []model.Point{p("d3", 50), p("d4", 50), p("d5", 100), p("d6", 50)}
	if !reflect.DeepEqual(k, wantK) || !reflect.DeepEqual(d, wantD) {
		t.Fatalf("k = %v, d = %v", k, d)
	}

	// Flat range: raw %K is 50, not a division by zero.
	flat := []model.Candle{c("a", 5, 5, 5), c("b", 5, 5, 5), c("c", 5, 5, 5)}
	if k, _ := Stochastic(flat, 2, 2, 1); len(k) != 1 || k[0].Value != 50 {
		t.Fatalf("flat: %v", k)
	}

	// Too few candles.
	if k, d := Stochastic(all[:1], 2, 1, 1); k != nil || d != nil {
		t.Fatalf("short: %v %v", k, d)
	}
}
