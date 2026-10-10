package scanner

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	v1 "alpaca-app/api/scanner/v1"
	"alpaca-app/internal/logic/scanner"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 { return &ControllerV1{} }

func (c *ControllerV1) Start(ctx context.Context, req *v1.StartReq) (res *v1.StartRes, err error) {
	limit := g.Cfg().MustGet(ctx, "scanner.maxSpreadPct", 2).Float64()
	s := scanner.Start(limit, resultFile(ctx))
	return &s, nil
}

func (c *ControllerV1) Saved(ctx context.Context, req *v1.SavedReq) (res *v1.SavedRes, err error) {
	s, err := scanner.Load(resultFile(ctx))
	if err != nil {
		return nil, gerror.Wrap(err, "load saved scan")
	}
	return s, nil
}

func resultFile(ctx context.Context) string {
	return g.Cfg().MustGet(ctx, "scanner.resultFile", "data/scan.json").String()
}

func (c *ControllerV1) Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error) {
	s := scanner.Status()
	return &s, nil
}
