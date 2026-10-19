package trade

import (
	"context"

	v1 "alpaca-app/api/trade/v1"
	"alpaca-app/internal/logic/trade"
	"alpaca-app/internal/model"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }

func (c *ControllerV1) Account(ctx context.Context, req *v1.AccountReq) (res *v1.AccountRes, err error) {
	a, err := trade.Account(ctx)
	if err != nil {
		return nil, err
	}
	return (*v1.AccountRes)(a), nil
}

func (c *ControllerV1) Positions(ctx context.Context, req *v1.PositionsReq) (res *v1.PositionsRes, err error) {
	ps, err := trade.Positions(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.PositionsRes{Positions: ps}, nil
}

func (c *ControllerV1) Orders(ctx context.Context, req *v1.OrdersReq) (res *v1.OrdersRes, err error) {
	os, err := trade.OpenOrders(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.OrdersRes{Orders: os}, nil
}

func (c *ControllerV1) Fills(ctx context.Context, req *v1.FillsReq) (res *v1.FillsRes, err error) {
	fs, err := trade.Fills(ctx, req.Symbol, req.Days)
	if err != nil {
		return nil, err
	}
	return &v1.FillsRes{Fills: fs}, nil
}

func (c *ControllerV1) Closed(ctx context.Context, req *v1.ClosedReq) (res *v1.ClosedRes, err error) {
	cs, err := trade.Closed(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ClosedRes{Closed: cs}, nil
}

func (c *ControllerV1) PlaceOrder(ctx context.Context, req *v1.PlaceOrderReq) (res *v1.PlaceOrderRes, err error) {
	o, err := trade.PlaceOrder(ctx, model.OrderInput{
		Symbol: req.Symbol, Side: req.Side, Type: req.Type, Qty: req.Qty, LimitPrice: req.LimitPrice,
	})
	if err != nil {
		return nil, err
	}
	return (*v1.PlaceOrderRes)(o), nil
}

func (c *ControllerV1) CancelOrder(ctx context.Context, req *v1.CancelOrderReq) (res *v1.CancelOrderRes, err error) {
	if err := trade.CancelOrder(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.CancelOrderRes{}, nil
}

func (c *ControllerV1) ClosePosition(ctx context.Context, req *v1.ClosePositionReq) (res *v1.ClosePositionRes, err error) {
	o, err := trade.ClosePosition(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}
	return (*v1.ClosePositionRes)(o), nil
}
