package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"alpaca-app/internal/model"
)

type BarsReq struct {
	g.Meta `path:"/bars" method:"get" tags:"Market" summary:"Daily candles for the last N days"`
	Symbol string `json:"symbol" v:"required|regex:^[A-Za-z.]{1,10}$#יש להזין סימול מניה|סימול המניה לא תקין"`
	Days   int    `json:"days" d:"30" v:"between:1,3650#מספר הימים חייב להיות בין 1 ל-3650"`
}

type BarsRes struct {
	Symbol  string         `json:"symbol"`
	Candles []model.Candle `json:"candles"`
}
