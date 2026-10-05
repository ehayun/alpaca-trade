package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	"alpaca-app/internal/consts"
)

func TestResponseMiddleware(t *testing.T) {
	s := g.Server(t.Name())
	s.SetAddr("127.0.0.1:0")
	s.SetDumpRouterMap(false)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(responseMiddleware)
		group.GET("/ok", func(r *ghttp.Request) {})
		group.GET("/user", func(r *ghttp.Request) { r.SetError(gerror.NewCode(consts.CodeUser, "הודעה למשתמש")) })
		group.GET("/internal", func(r *ghttp.Request) { r.SetError(errors.New("db password=secret")) })
		group.GET("/panic", func(r *ghttp.Request) { panic("secret panic") })
	})
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	defer s.Shutdown()
	base := fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort())

	cases := []struct {
		path string
		code int
		msg  string
	}{
		{"/ok", gcode.CodeOK.Code(), "OK"},
		{"/user", consts.CodeUser.Code(), "הודעה למשתמש"},
		{"/internal", gcode.CodeInternalError.Code(), consts.MsgInternal},
		{"/panic", gcode.CodeInternalError.Code(), consts.MsgInternal},
	}
	for _, c := range cases {
		resp, err := http.Get(base + c.path)
		if err != nil {
			t.Fatal(err)
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var body ghttp.DefaultHandlerResponse
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Fatalf("%s: not JSON: %s", c.path, raw)
		}
		if body.Code != c.code || body.Message != c.msg {
			t.Errorf("%s: got code=%d msg=%q, want code=%d msg=%q", c.path, body.Code, body.Message, c.code, c.msg)
		}
	}
}
