package scanner

import (
	"math"
	"slices"
	"testing"
	"time"

	"alpaca-app/internal/model"
)

func flat(n int, close float64) []model.Candle {
	out := make([]model.Candle, n)
	for i := range out {
		out[i] = model.Candle{Time: "d", Close: close}
	}
	return out
}

func TestEvaluate(t *testing.T) {
	// Flat price: all EMAs equal the price, spread 0.
	m, ok, _ := Evaluate("FLAT", flat(250, 50), "")
	if !ok || m.SpreadPct != 0 || m.EMA200 != 50 || m.Close != 50 {
		t.Fatalf("flat: %+v ok=%v", m, ok)
	}

	// A jump at the end moves EMA 20 more than EMA 200, so the spread grows.
	c := append(flat(240, 100), flat(10, 120)...)
	m, ok, _ = Evaluate("JUMP", c, "")
	if !ok || !(m.EMA20 > m.EMA100 && m.EMA100 > m.EMA200) {
		t.Fatalf("jump: %+v", m)
	}
	want := (m.EMA20 - m.EMA200) / 120 * 100
	if math.Abs(m.SpreadPct-want) > 1e-9 || m.SpreadPct < 2 {
		t.Fatalf("spread = %v, want %v (> 2)", m.SpreadPct, want)
	}

	// EMAs close together, but the last close is far away: only the price spread is large.
	c = append(flat(249, 100), flat(1, 110)...)
	m, _, _ = Evaluate("GAP", c, "")
	if m.SpreadPct > 1 || m.PriceSpreadPct < 8 || m.Spread(false) != m.SpreadPct || m.Spread(true) != m.PriceSpreadPct {
		t.Fatalf("gap: spread %.2f, with price %.2f", m.SpreadPct, m.PriceSpreadPct)
	}
	want = (110 - m.EMA200) / 110 * 100 // close is the highest value, EMA 200 the lowest
	if math.Abs(m.PriceSpreadPct-want) > 1e-9 {
		t.Fatalf("price spread = %v, want %v", m.PriceSpreadPct, want)
	}

	// Too little history for EMA 200.
	if _, ok, _ := Evaluate("NEW", flat(199, 10), ""); ok {
		t.Fatal("199 candles: want ok=false")
	}
}

func TestSymbols(t *testing.T) {
	s := Symbols()
	if len(s) < 500 || !slices.IsSorted(s) || len(slices.Compact(slices.Clone(s))) != len(s) {
		t.Fatalf("got %d symbols, want 500+ sorted and unique", len(s))
	}
	for _, want := range []string{"AAPL", "BRK.B", "MSFT"} {
		if !slices.Contains(s, want) {
			t.Fatalf("missing %s", want)
		}
	}
}

func TestSaveLoad(t *testing.T) {
	path := t.TempDir() + "/sub/scan.json"
	if s, err := Load(path); s != nil || err != nil {
		t.Fatalf("no file: got %+v, %v; want nil, nil", s, err)
	}
	now := time.Date(2026, 10, 10, 5, 44, 0, 0, time.UTC)
	in := model.ScanStatus{FinishedAt: &now, Total: 518, Scanned: 514, MaxSpreadPct: 2,
		Matches: []model.ScanMatch{{Symbol: "WRB", Close: 71.91, SpreadPct: 0.18}}}
	if err := Save(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := Load(path)
	if err != nil || out == nil || !out.FinishedAt.Equal(now) || out.Scanned != 514 || len(out.Matches) != 1 || out.Matches[0] != in.Matches[0] {
		t.Fatalf("got %+v, %v", out, err)
	}
}

func TestStochPass(t *testing.T) {
	pts := func(vals ...float64) []model.Point {
		out := make([]model.Point, len(vals))
		for i, v := range vals {
			out[i] = model.Point{Time: string(rune('a' + i)), Value: v}
		}
		return out
	}
	// %K has one more point than %D at the start; both end on the same day.
	k := pts(50, 15, 10, 25, 30)
	d := pts(20, 15, 20, 28) // aligned: k 15/20, 10/15, 25/20 (cross up from 10), 30/28

	cases := []struct {
		rule string
		k, d []model.Point
		want bool
	}{
		{"", nil, nil, true},
		{StochOversold, pts(30, 20), pts(25, 22), true},
		{StochOversold, pts(30, 21), pts(25, 22), false},
		{StochOverbought, pts(70, 80), pts(60, 75), true},
		{StochOverbought, pts(70, 79), pts(60, 75), false},
		{StochBullCross, k, d, true},
		{StochBullCross, pts(50, 30, 35, 40, 45), pts(40, 36, 38, 42), false},              // cross from 30: not oversold
		{StochBullCross, pts(9, 5, 10, 25, 20, 18, 17), pts(8, 12, 20, 22, 21, 19), false}, // cross 4 days ago
		{"unknown", k, d, false},
	}
	for _, c := range cases {
		if got := StochPass(c.rule, c.k, c.d); got != c.want {
			t.Errorf("StochPass(%q, %v, %v) = %v, want %v", c.rule, c.k, c.d, got, c.want)
		}
	}
}
