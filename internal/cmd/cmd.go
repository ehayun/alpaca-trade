package cmd

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"

	"alpaca-app/internal/controller/market"
	"alpaca-app/internal/controller/scanner"
	"alpaca-app/internal/controller/trade"
)

var Main = gcmd.Command{
	Name:  "main",
	Usage: "main",
	Brief: "start http server",
	Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
		a, err := newAuth()
		if err != nil {
			return err
		}
		s := g.Server()
		s.Group("/", func(group *ghttp.RouterGroup) {
			group.GET("/healthz", func(r *ghttp.Request) { r.Response.Write("ok") })
			group.GET("/login", a.loginPage)
			group.POST("/login", a.login)
		})
		s.Group("/", func(group *ghttp.RouterGroup) {
			group.Middleware(a.middleware)
			group.GET("/", func(r *ghttp.Request) {
				tpl := g.Map{"maxSpreadPct": scanner.DefaultMaxSpreadPct(r.Context())}
				if err := r.Response.WriteTpl("index.html", tpl); err != nil {
					g.Log().Errorf(r.Context(), "%+v", err)
				}
			})
		})
		s.Group("/api", func(group *ghttp.RouterGroup) {
			group.Middleware(a.middleware, responseMiddleware)
			group.Bind(market.NewV1(), trade.NewV1(), scanner.NewV1())
		})
		s.Run()
		return nil
	},
}
