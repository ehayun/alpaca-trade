package scanner

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "alpaca-app/api/scanner/v1"
	"alpaca-app/internal/logic/scanner"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }

func (c *ControllerV1) Start(ctx context.Context, req *v1.StartReq) (res *v1.StartRes, err error) {
	limit := g.Cfg().MustGet(ctx, "scanner.maxSpreadPct", 1).Float64()
	s := scanner.Start(limit)
	return &s, nil
}

func (c *ControllerV1) Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error) {
	s := scanner.Status()
	return &s, nil
}
