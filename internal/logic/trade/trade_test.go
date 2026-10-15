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
