package trade

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"alpaca-app/internal/consts"
	"alpaca-app/internal/model"
)

func TestBuildOrder(t *testing.T) {
	req, err := BuildOrder(model.OrderInput{Symbol: " aapl ", Side: "buy", Type: "limit", Qty: "2", LimitPrice: "150.5"})
	if err != nil {
		t.Fatal(err)
	}
	if req.Symbol != "AAPL" || req.Qty.String() != "2" || req.LimitPrice.String() != "150.5" ||
		req.Side != alpaca.Buy || req.TimeInForce != alpaca.Day {
		t.Fatalf("unexpected request: %+v", req)
	}

	if req, err = BuildOrder(model.OrderInput{Symbol: "MSFT", Side: "sell", Type: "market", Qty: "0.5"}); err != nil || req.LimitPrice != nil {
		t.Fatalf("market order: %+v %v", req, err)
	}

	bad := []model.OrderInput{
		{Symbol: "A", Side: "buy", Type: "market", Qty: "0"},
		{Symbol: "A", Side: "buy", Type: "market", Qty: "abc"},
		{Symbol: "A", Side: "buy", Type: "limit", Qty: "1"},
		{Symbol: "A", Side: "buy", Type: "limit", Qty: "1", LimitPrice: "-3"},
		{Symbol: "A", Side: "buy", Type: "stop", Qty: "1"},
		{Symbol: "A", Side: "short", Type: "market", Qty: "1"},
	}
	for _, in := range bad {
		if _, err := BuildOrder(in); gerror.Code(err) != consts.CodeUser || err.Error() == "" {
			t.Errorf("%+v: want user error, got %v", in, err)
		}
	}
}

func TestRejected(t *testing.T) {
	buyingPower := fmt.Errorf("wrapped: %w", &alpaca.APIError{StatusCode: 403, Message: "insufficient buying power"})
	if err := rejected(buyingPower, "נדחה"); gerror.Code(err) != consts.CodeUser || err.Error() != "נדחה: insufficient buying power" {
		t.Fatalf("4xx must become a user error, got %v", err)
	}
	for _, err := range []error{
		&alpaca.APIError{StatusCode: 401, Message: "request is not authorized"},
		&alpaca.APIError{StatusCode: 500, Message: "boom"},
		errors.New("network down"),
	} {
		if gerror.Code(rejected(err, "x")) == consts.CodeUser {
			t.Errorf("%v must stay hidden from the user", err)
		}
	}
	if rejected(nil, "x") != nil {
		t.Fatal("nil must stay nil")
	}
}

func TestPaperOnly(t *testing.T) {
	if PaperURL != "https://paper-api.alpaca.markets" {
		t.Fatal("trading must point at the paper API")
	}
}

func TestToFills(t *testing.T) {
	late := time.Date(2026, 3, 3, 2, 30, 0, 0, time.UTC) // 21:30 on 2 March in New York
	price := decimal.RequireFromString("101.25")
	got := toFills([]alpaca.Order{
		{Side: alpaca.Buy, FilledAt: &late, FilledQty: decimal.NewFromInt(3), FilledAvgPrice: &price},
		{Side: alpaca.Sell, Status: "canceled"}, // no fill: dropped
		{Side: alpaca.Sell, FilledAt: &late, FilledQty: decimal.Zero, FilledAvgPrice: &price},
	})
	if len(got) != 1 || got[0].Day != "2026-03-02" || got[0].Side != "buy" || got[0].Qty.String() != "3" || got[0].Price.String() != "101.25" {
		t.Fatalf("unexpected fills: %+v", got)
	}
	if toFills(nil) == nil {
		t.Fatal("no fills must be [] in JSON, not null")
	}
}

func TestToClosed(t *testing.T) {
	day1 := time.Date(2026, 3, 2, 15, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 3, 5, 15, 0, 0, 0, time.UTC)
	p := func(v string) *decimal.Decimal { d := decimal.RequireFromString(v); return &d }
	fill := func(sym string, side alpaca.Side, at *time.Time, qty int64, price string) alpaca.Order {
		return alpaca.Order{Symbol: sym, Side: side, FilledAt: at, FilledQty: decimal.NewFromInt(qty), FilledAvgPrice: p(price)}
	}
	got := toClosed([]alpaca.Order{
		fill("AAPL", alpaca.Sell, &day2, 3, "110"),
		fill("AAPL", alpaca.Buy, &day1, 1, "100"),
		fill("AAPL", alpaca.Buy, &day1, 2, "103"),
		fill("MSFT", alpaca.Sell, &day1, 1, "90"),
		fill("MSFT", alpaca.Buy, &day1, 1, "100"),
		fill("NVDA", alpaca.Buy, &day1, 2, "50"), // only part sold: left out
		fill("NVDA", alpaca.Sell, &day1, 1, "55"),
		fill("TSLA", alpaca.Buy, &day1, 1, "10"), // open position: left out
		fill("TSLA", alpaca.Sell, &day1, 1, "12"),
		{Symbol: "AMD", Side: alpaca.Buy, Status: "canceled"}, // no fill
	}, map[string]bool{"TSLA": true})
	if len(got) != 2 {
		t.Fatalf("want AAPL and MSFT, got %+v", got)
	}
	a, m := got[0], got[1]
	if a.Symbol != "AAPL" || a.Day != "2026-03-05" || a.Qty.String() != "3" || a.AvgBuy.String() != "102" ||
		a.AvgSell.String() != "110" || a.RealizedPL.String() != "24" {
		t.Fatalf("AAPL: %+v", a)
	}
	if m.Symbol != "MSFT" || m.RealizedPL.String() != "-10" {
		t.Fatalf("MSFT: %+v", m)
	}
	if toClosed(nil, nil) == nil {
		t.Fatal("no trades must be [] in JSON, not null")
	}
}
