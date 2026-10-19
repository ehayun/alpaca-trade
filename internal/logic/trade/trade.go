// Package trade wraps the Alpaca trading API. Paper trading only.
package trade

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"alpaca-app/internal/consts"
	"alpaca-app/internal/model"
)

// PaperURL is hard-coded on purpose: the SDK defaults to the live (real money) API.
const PaperURL = "https://paper-api.alpaca.markets"

// Reads APCA_API_KEY_ID / APCA_API_SECRET_KEY from the environment.
var client = alpaca.NewClient(alpaca.ClientOpts{BaseURL: PaperURL})

var newYork = mustLoad("America/New_York")

// userErr marks a Hebrew message as safe to show the user.
func userErr(msg string) error { return gerror.NewCode(consts.CodeUser, msg) }

func Account(ctx context.Context) (*model.Account, error) {
	a, err := client.GetAccount()
	if err != nil {
		return nil, err
	}
	return &model.Account{Equity: a.Equity, Cash: a.Cash, BuyingPower: a.BuyingPower}, nil
}

// Asset returns company data; an unknown symbol is a user error.
func Asset(ctx context.Context, symbol string) (*model.Asset, error) {
	a, err := client.GetAsset(symbol)
	var apiErr *alpaca.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
		return nil, userErr("הסימול " + symbol + " לא נמצא")
	}
	if err != nil {
		return nil, err
	}
	return &model.Asset{
		Name: a.Name, Exchange: a.Exchange, Tradable: a.Tradable, Shortable: a.Shortable, Fractionable: a.Fractionable,
	}, nil
}

func Positions(ctx context.Context) ([]model.Position, error) {
	ps, err := client.GetPositions()
	if err != nil {
		return nil, err
	}
	out := make([]model.Position, len(ps))
	for i, p := range ps {
		out[i] = model.Position{
			Symbol: p.Symbol, Qty: p.Qty, AvgEntryPrice: p.AvgEntryPrice,
			CurrentPrice: p.CurrentPrice, MarketValue: p.MarketValue, UnrealizedPL: p.UnrealizedPL,
		}
	}
	return out, nil
}

func OpenOrders(ctx context.Context) ([]model.Order, error) {
	os, err := client.GetOrders(alpaca.GetOrdersRequest{Status: "open", Limit: 100})
	if err != nil {
		return nil, err
	}
	out := make([]model.Order, len(os))
	for i := range os {
		out[i] = toOrder(&os[i])
	}
	return out, nil
}

// Fills returns the filled orders of one symbol in the last N days, oldest first.
// ponytail: one page of 500 closed orders; page with Until if a symbol gets more trades than that.
func Fills(ctx context.Context, symbol string, days int) ([]model.Fill, error) {
	os, err := client.GetOrders(alpaca.GetOrdersRequest{
		Status: "closed", Limit: 500, Direction: "asc", Symbols: []string{strings.ToUpper(symbol)},
		After: time.Now().AddDate(0, 0, -days),
	})
	if err != nil {
		return nil, err
	}
	return toFills(os), nil
}

// toFills keeps the orders that filled at least in part (a canceled order can have a partial fill).
func toFills(os []alpaca.Order) []model.Fill {
	out := []model.Fill{}
	for _, o := range os {
		if o.FilledAt == nil || o.FilledAvgPrice == nil || !o.FilledQty.IsPositive() {
			continue
		}
		out = append(out, model.Fill{
			Day: o.FilledAt.In(newYork).Format(time.DateOnly), Side: string(o.Side), Qty: o.FilledQty, Price: *o.FilledAvgPrice,
		})
	}
	return out
}

// Closed returns the stocks with no open position whose filled buys and sells are equal, last closed first.
// ponytail: one page of the newest 500 closed orders; page with Until if the account gets more trades than that.
func Closed(ctx context.Context) ([]model.Closed, error) {
	os, err := client.GetOrders(alpaca.GetOrdersRequest{Status: "closed", Limit: 500, Direction: "desc"})
	if err != nil {
		return nil, err
	}
	ps, err := client.GetPositions()
	if err != nil {
		return nil, err
	}
	open := map[string]bool{}
	for _, p := range ps {
		open[p.Symbol] = true
	}
	return toClosed(os, open), nil
}

// toClosed groups the fills by symbol. A symbol whose buys and sells differ (still open, or older
// fills out of the page) has no true realized P/L, so it is left out.
func toClosed(os []alpaca.Order, open map[string]bool) []model.Closed {
	type sums struct {
		buyQty, buyCost, sellQty, sellCost decimal.Decimal
		day                                string
	}
	bySymbol := map[string]*sums{}
	for _, o := range os {
		if open[o.Symbol] || o.FilledAt == nil || o.FilledAvgPrice == nil || !o.FilledQty.IsPositive() {
			continue
		}
		s := bySymbol[o.Symbol]
		if s == nil {
			s = &sums{}
			bySymbol[o.Symbol] = s
		}
		if day := o.FilledAt.In(newYork).Format(time.DateOnly); day > s.day {
			s.day = day
		}
		cost := o.FilledQty.Mul(*o.FilledAvgPrice)
		if o.Side == alpaca.Buy {
			s.buyQty, s.buyCost = s.buyQty.Add(o.FilledQty), s.buyCost.Add(cost)
		} else {
			s.sellQty, s.sellCost = s.sellQty.Add(o.FilledQty), s.sellCost.Add(cost)
		}
	}
	out := []model.Closed{}
	for sym, s := range bySymbol {
		if !s.buyQty.IsPositive() || !s.buyQty.Equal(s.sellQty) {
			continue
		}
		out = append(out, model.Closed{
			Symbol: sym, Qty: s.buyQty, Day: s.day,
			AvgBuy: s.buyCost.Div(s.buyQty).Round(4), AvgSell: s.sellCost.Div(s.sellQty).Round(4),
			RealizedPL: s.sellCost.Sub(s.buyCost),
		})
	}
	slices.SortFunc(out, func(a, b model.Closed) int {
		if c := strings.Compare(b.Day, a.Day); c != 0 {
			return c
		}
		return strings.Compare(a.Symbol, b.Symbol)
	})
	return out
}

func PlaceOrder(ctx context.Context, in model.OrderInput) (*model.Order, error) {
	req, err := BuildOrder(in)
	if err != nil {
		return nil, err
	}
	o, err := client.PlaceOrder(req)
	if err != nil {
		return nil, rejected(err, "הפקודה נדחתה")
	}
	res := toOrder(o)
	return &res, nil
}

func CancelOrder(ctx context.Context, id string) error {
	return rejected(client.CancelOrder(id), "לא ניתן לבטל את הפקודה")
}

// ClosePosition closes the whole position with a market order (Alpaca sells a long, buys back a short).
func ClosePosition(ctx context.Context, symbol string) (*model.Order, error) {
	symbol = strings.ToUpper(symbol)
	o, err := client.ClosePosition(symbol, alpaca.ClosePositionRequest{})
	if err != nil {
		return nil, rejected(err, "לא ניתן לסגור את הפוזיציה "+symbol)
	}
	res := toOrder(o)
	return &res, nil
}

// BuildOrder turns user input into an SDK order request. Day orders only.
func BuildOrder(in model.OrderInput) (alpaca.PlaceOrderRequest, error) {
	qty, err := decimal.NewFromString(strings.TrimSpace(in.Qty))
	if err != nil || !qty.IsPositive() {
		return alpaca.PlaceOrderRequest{}, userErr("הכמות חייבת להיות מספר חיובי")
	}
	req := alpaca.PlaceOrderRequest{
		Symbol:      strings.ToUpper(strings.TrimSpace(in.Symbol)),
		Qty:         &qty,
		Side:        alpaca.Side(in.Side),
		Type:        alpaca.OrderType(in.Type),
		TimeInForce: alpaca.Day,
	}
	if req.Side != alpaca.Buy && req.Side != alpaca.Sell {
		return alpaca.PlaceOrderRequest{}, userErr("כיוון פקודה לא נתמך")
	}
	switch req.Type {
	case alpaca.Market:
	case alpaca.Limit:
		price, err := decimal.NewFromString(strings.TrimSpace(in.LimitPrice))
		if err != nil || !price.IsPositive() {
			return alpaca.PlaceOrderRequest{}, userErr("בפקודת לימיט יש להזין מחיר חיובי")
		}
		req.LimitPrice = &price
	default:
		return alpaca.PlaceOrderRequest{}, userErr("סוג פקודה לא נתמך")
	}
	return req, nil
}

// rejected turns an Alpaca 4xx (e.g. insufficient buying power) into a user error with Alpaca's reason.
// Other errors (network, 5xx, bad API keys) pass through and stay hidden from the user.
func rejected(err error, prefix string) error {
	var apiErr *alpaca.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode >= 400 && apiErr.StatusCode < 500 && apiErr.StatusCode != 401 {
		return userErr(prefix + ": " + apiErr.Message)
	}
	return err
}

func toOrder(o *alpaca.Order) model.Order {
	return model.Order{
		ID: o.ID, CreatedAt: o.CreatedAt, Symbol: o.Symbol, Side: string(o.Side), Type: string(o.Type),
		Qty: o.Qty, LimitPrice: o.LimitPrice, Status: o.Status,
	}
}

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
