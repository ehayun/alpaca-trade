package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func TestAuthToken(t *testing.T) {
	a := &auth{"eli", "pw"}
	now := time.Now()
	tok := a.token(now)
	if !a.valid(tok, now) {
		t.Error("fresh token rejected")
	}
	if a.valid(tok, now.Add(authTTL)) {
		t.Error("expired token accepted")
	}
	exp, sig, _ := strings.Cut(tok, ".")
	if a.valid(fmt.Sprint(now.Add(10*authTTL).Unix())+"."+sig, now) {
		t.Error("token with a changed expiry accepted")
	}
	if (&auth{"eli", "other"}).valid(tok, now) {
		t.Error("token accepted after a password change")
	}
	if a.valid(exp, now) || a.valid("", now) {
		t.Error("malformed token accepted")
	}
	if !a.check("eli", "pw") || a.check("eli", "PW") || a.check("", "") {
		t.Error("check wrong")
	}
}

func TestAuthMiddleware(t *testing.T) {
	a := &auth{"eli", "pw"}
	s := g.Server(t.Name())
	s.SetAddr("127.0.0.1:0")
	s.SetDumpRouterMap(false)
	s.Group("/", func(group *ghttp.RouterGroup) {
		group.POST("/login", a.login)
		group.Middleware(a.middleware)
		group.GET("/", func(r *ghttp.Request) { r.Response.Write("page") })
		group.GET("/api/x", func(r *ghttp.Request) { r.Response.Write("data") })
	})
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	defer s.Shutdown()
	base := fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort())
	c := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

	get := func(path, cookie string) *http.Response {
		req, _ := http.NewRequest("GET", base+path, nil)
		if cookie != "" {
			req.Header.Set("Cookie", cookie)
		}
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp
	}
	if r := get("/", ""); r.StatusCode != http.StatusFound || r.Header.Get("Location") != "/login" {
		t.Errorf("page without cookie: %d %s", r.StatusCode, r.Header.Get("Location"))
	}
	if r := get("/api/x", ""); r.StatusCode != http.StatusUnauthorized {
		t.Errorf("api without cookie: %d", r.StatusCode)
	}

	resp, err := c.PostForm(base+"/login", url.Values{"user": {"eli"}, "password": {"pw"}})
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	var cookie string
	for _, ck := range resp.Cookies() {
		if ck.Name == authCookie && ck.HttpOnly && time.Until(ck.Expires).Round(time.Minute) == authTTL {
			cookie = ck.Name + "=" + ck.Value
		}
	}
	if cookie == "" {
		t.Fatalf("login set no HttpOnly 60-minute cookie: %v", resp.Header["Set-Cookie"])
	}
	if r := get("/api/x", cookie); r.StatusCode != http.StatusOK {
		t.Errorf("api with cookie: %d", r.StatusCode)
	}
}
