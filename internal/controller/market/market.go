package market

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "alpaca-app/api/market/v1"
	"alpaca-app/internal/consts"
	"alpaca-app/internal/logic/market"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }

func (c *ControllerV1) Bars(ctx context.Context, req *v1.BarsReq) (res *v1.BarsRes, err error) {
	symbol := strings.ToUpper(req.Symbol)
	candles, err := market.DailyCandles(ctx, symbol, req.Days)
	if err != nil {
		return nil, gerror.Wrapf(err, "bars %s", symbol)
	}
	if len(candles) == 0 {
		return nil, gerror.NewCode(consts.CodeUser, "לא נמצאו נתונים עבור "+symbol+" בטווח הזה")
	}
	return &v1.BarsRes{Symbol: symbol, Candles: candles}, nil
}
