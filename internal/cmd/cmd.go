package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"alpaca-app/internal/controller/market"
	"alpaca-app/internal/controller/trade"
)

var Main = gcmd.Command{
	Name:  "main",
	Usage: "main",
	Brief: "start http server",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		s := g.Server()
		s.Group("/", func(group *ghttp.RouterGroup) {
			group.GET("/", func(r *ghttp.Request) {
				if err := r.Response.WriteTpl("index.html"); err != nil {
					g.Log().Errorf(r.Context(), "%+v", err)
				}
			})
			group.GET("/healthz", func(r *ghttp.Request) { r.Response.Write("ok") })
		})
		s.Group("/api", func(group *ghttp.RouterGroup) {
			group.Middleware(responseMiddleware)
			group.Bind(market.NewV1(), trade.NewV1())
		})
		s.Run()
		return nil
	},
}
