package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"alpaca-app/internal/model"
)

type StartReq struct {
	g.Meta       `path:"/scan" method:"post" tags:"Scanner" summary:"Start an EMA scan of the S&P 500, Nasdaq-100 and Dow 30"`
	MaxSpreadPct float64 `json:"maxSpreadPct" v:"between:0.1,20#הפער חייב להיות בין 0.1% ל-20%"` // 0 or missing = config default
	PriceClose   bool    `json:"priceClose"`                                                     // the close must also be inside the limit
}

type StartRes = model.ScanStatus

type StatusReq struct {
	g.Meta `path:"/scan" method:"get" tags:"Scanner" summary:"Scan progress, or the result of the last scan"`
}

type StatusRes = model.ScanStatus

type SavedReq struct {
	g.Meta `path:"/scan/saved" method:"get" tags:"Scanner" summary:"The scan saved in the result file; null if there is none"`
}

type SavedRes = model.ScanStatus

type SymbolsReq struct {
	g.Meta `path:"/symbols" method:"get" tags:"Scanner" summary:"All symbols of the S&P 500, Nasdaq-100 and Dow 30, sorted"`
}

type SymbolsRes struct {
	Symbols []string `json:"symbols"`
}
