package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"alpaca-app/internal/model"
)

type BarsReq struct {
	g.Meta `path:"/bars" method:"get" tags:"Market" summary:"Daily candles for the last N days"`
	Symbol string `json:"symbol" v:"required|regex:^[A-Za-z.]{1,10}$#יש להזין סימול מניה|סימול המניה לא תקין"`
	Days   int    `json:"days" d:"300" v:"between:1,3650#מספר הימים חייב להיות בין 1 ל-3650"`
}

type BarsRes struct {
	Symbol  string         `json:"symbol"`
	Candles []model.Candle `json:"candles"`
	EMAs    []model.EMA    `json:"emas"`
	MACD    model.MACD     `json:"macd"`
}

type InfoReq struct {
	g.Meta `path:"/info" method:"get" tags:"Market" summary:"Company data, today's price and recent news"`
	Symbol string `json:"symbol" v:"required|regex:^[A-Za-z.]{1,10}$#יש להזין סימול מניה|סימול המניה לא תקין"`
}

type InfoRes = model.StockInfo

type NewsSummaryReq struct {
	g.Meta `path:"/news-summary" method:"get" tags:"Market" summary:"Short Hebrew summary of recent news, by Claude"`
	Symbol string `json:"symbol" v:"required|regex:^[A-Za-z.]{1,10}$#יש להזין סימול מניה|סימול המניה לא תקין"`
}

type NewsSummaryRes struct {
	Symbol  string `json:"symbol"`
	Summary string `json:"summary"`
}

type PricesReq struct {
	g.Meta  `path:"/prices" method:"get" tags:"Market" summary:"Latest price for up to 100 symbols"`
	Symbols string `json:"symbols" v:"required|regex:^[A-Za-z.]{1,10}(,[A-Za-z.]{1,10}){0,99}$#יש להזין סימולים|רשימת הסימולים לא תקינה"`
}

type PricesRes struct {
	Prices map[string]float64 `json:"prices"`
}
