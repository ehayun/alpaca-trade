package market

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "alpaca-app/api/market/v1"
	"alpaca-app/internal/consts"
	"alpaca-app/internal/logic/market"
	"alpaca-app/internal/logic/summary"
	"alpaca-app/internal/logic/trade"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }

func (c *ControllerV1) Bars(ctx context.Context, req *v1.BarsReq) (res *v1.BarsRes, err error) {
	symbol := strings.ToUpper(req.Symbol)
	candles, emas, err := market.DailyCandles(ctx, symbol, req.Days)
	if err != nil {
		return nil, gerror.Wrapf(err, "bars %s", symbol)
	}
	if len(candles) == 0 {
		return nil, gerror.NewCode(consts.CodeUser, "לא נמצאו נתונים עבור "+symbol+" בטווח הזה")
	}
	return &v1.BarsRes{Symbol: symbol, Candles: candles, EMAs: emas}, nil
}

// ponytail: three sequential Alpaca calls; run them concurrently if the tab feels slow.
func (c *ControllerV1) Info(ctx context.Context, req *v1.InfoReq) (res *v1.InfoRes, err error) {
	symbol := strings.ToUpper(req.Symbol)
	asset, err := trade.Asset(ctx, symbol)
	if err != nil {
		return nil, gerror.Wrapf(err, "asset %s", symbol)
	}
	quote, err := market.Quote(ctx, symbol)
	if err != nil {
		return nil, gerror.Wrapf(err, "snapshot %s", symbol)
	}
	news, err := market.News(ctx, symbol, 5)
	if err != nil {
		return nil, gerror.Wrapf(err, "news %s", symbol)
	}
	return &v1.InfoRes{Symbol: symbol, Asset: *asset, Quote: quote, News: news}, nil
}

func (c *ControllerV1) NewsSummary(ctx context.Context, req *v1.NewsSummaryReq) (res *v1.NewsSummaryRes, err error) {
	symbol := strings.ToUpper(req.Symbol)
	text, err := summary.Cached(ctx, symbol)
	if err != nil {
		return nil, gerror.Wrapf(err, "news summary %s", symbol)
	}
	return &v1.NewsSummaryRes{Symbol: symbol, Summary: text}, nil
}

func (c *ControllerV1) Prices(ctx context.Context, req *v1.PricesReq) (res *v1.PricesRes, err error) {
	prices, err := market.LatestPrices(ctx, strings.Split(strings.ToUpper(req.Symbols), ","))
	if err != nil {
		return nil, gerror.Wrap(err, "latest prices")
	}
	return &v1.PricesRes{Prices: prices}, nil
}
