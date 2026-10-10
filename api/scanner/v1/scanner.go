package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"alpaca-app/internal/model"
)

type StartReq struct {
	g.Meta `path:"/scan" method:"post" tags:"Scanner" summary:"Start an EMA scan of the S&P 500, Nasdaq-100 and Dow 30"`
}

type StartRes = model.ScanStatus

type StatusReq struct {
	g.Meta `path:"/scan" method:"get" tags:"Scanner" summary:"Scan progress, or the result of the last scan"`
}

type StatusRes = model.ScanStatus
