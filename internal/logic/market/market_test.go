package market

import (
	"math"
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

func TestMACD(t *testing.T) {
	closes := func(vals ...float64) []model.Candle {
		out := make([]model.Candle, len(vals))
		for i, v := range vals {
			out[i] = model.Candle{Time: string(rune('a' + i)), Close: v}
		}
		return out
	}
	// fast 2, slow 3, signal 2 on closes 1,2,3,4,5,9:
	// EMA2 (k=2/3): b 1.5, c 2.5, d 3.5, e 4.5, f 7.5 · EMA3 (k=1/2): c 2, d 3, e 4, f 6.5
	// MACD: c 0.5, d 0.5, e 0.5, f 1 · signal (k=2/3): d 0.5, e 0.5, f 0.8333 · hist: d 0, e 0, f 0.1667
	line, sig, hist := MACD(closes(1, 2, 3, 4, 5, 9), 2, 3, 2)
	near := func(got []model.Point, want ...float64) bool {
		if len(got) != len(want) {
			return false
		}
		for i := range want {
			if math.Abs(got[i].Value-want[i]) > 1e-9 {
				return false
			}
		}
		return true
	}
	if !near(line, 0.5, 0.5, 0.5, 1) || line[0].Time != "c" {
		t.Fatalf("macd = %v", line)
	}
	if !near(sig, 0.5, 0.5, 2.0/3+0.5/3) || sig[0].Time != "d" {
		t.Fatalf("signal = %v", sig)
	}
	if !near(hist, 0, 0, 1-(2.0/3+0.5/3)) || hist[2].Time != "f" {
		t.Fatalf("hist = %v", hist)
	}
	// Too few candles for the slow EMA.
	if l, s, h := MACD(closes(1, 2), 2, 3, 2); l != nil || s != nil || h != nil {
		t.Fatalf("short: %v %v %v", l, s, h)
	}
}
