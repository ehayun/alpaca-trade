package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"alpaca-app/internal/model"
)

type AccountReq struct {
	g.Meta `path:"/account" method:"get" tags:"Trade" summary:"Paper account summary"`
}

type AccountRes model.Account

type PositionsReq struct {
	g.Meta `path:"/positions" method:"get" tags:"Trade" summary:"Open positions"`
}

type PositionsRes struct {
	Positions []model.Position `json:"positions"`
}

type OrdersReq struct {
	g.Meta `path:"/orders" method:"get" tags:"Trade" summary:"Open orders"`
}

type OrdersRes struct {
	Orders []model.Order `json:"orders"`
}

type FillsReq struct {
	g.Meta `path:"/fills" method:"get" tags:"Trade" summary:"Filled orders of one symbol in the last N days"`
	Symbol string `json:"symbol" v:"required|regex:^[A-Za-z.]{1,10}$#יש להזין סימול מניה|סימול המניה לא תקין"`
	Days   int    `json:"days" v:"required|between:1,3650#יש להזין מספר ימים|מספר הימים חייב להיות בין 1 ל-3650"`
}

type FillsRes struct {
	Fills []model.Fill `json:"fills"`
}

type PlaceOrderReq struct {
	g.Meta     `path:"/orders" method:"post" tags:"Trade" summary:"Place a day order (market or limit)"`
	Symbol     string `json:"symbol" v:"required|regex:^[A-Za-z.]{1,10}$#יש להזין סימול מניה|סימול המניה לא תקין"`
	Side       string `json:"side" v:"required|in:buy,sell#יש לבחור קנייה או מכירה|יש לבחור קנייה או מכירה"`
	Type       string `json:"type" v:"required|in:market,limit#יש לבחור סוג פקודה|סוג פקודה לא נתמך"`
	Qty        string `json:"qty" v:"required#יש להזין כמות"`
	LimitPrice string `json:"limitPrice" v:"required-if:type,limit#בפקודת לימיט יש להזין מחיר"`
}

type PlaceOrderRes model.Order

type CancelOrderReq struct {
	g.Meta `path:"/orders/{id}" method:"delete" tags:"Trade" summary:"Cancel an open order"`
	Id     string `json:"id" in:"path" v:"required|regex:^[0-9a-fA-F-]{36}$#חסר מזהה פקודה|מזהה פקודה לא תקין"`
}

type CancelOrderRes struct{}

type ClosePositionReq struct {
	g.Meta `path:"/positions/{symbol}" method:"delete" tags:"Trade" summary:"Close a whole position at market price"`
	Symbol string `json:"symbol" in:"path" v:"required|regex:^[A-Za-z.]{1,10}$#יש להזין סימול מניה|סימול המניה לא תקין"`
}

type ClosePositionRes model.Order
