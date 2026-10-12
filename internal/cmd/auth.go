package cmd

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

const (
	authCookie = "auth"
	authTTL    = 60 * time.Minute
)

// auth holds the login from APP_USER / APP_PASSWORD.
// The cookie is "<expiry unix>.<HMAC(password, user|expiry)>": no server state,
// and a new password makes all old cookies invalid.
// ponytail: one user, no logout list; a stolen cookie is good until it expires.
type auth struct{ user, password string }

func newAuth() (*auth, error) {
	a := &auth{os.Getenv("APP_USER"), os.Getenv("APP_PASSWORD")}
	if a.user == "" || a.password == "" {
		return nil, errors.New("APP_USER and APP_PASSWORD must be set (see .env.example)")
	}
	return a, nil
}

func (a *auth) sign(exp string) string {
	m := hmac.New(sha256.New, []byte(a.password))
	m.Write([]byte(a.user + "|" + exp))
	return hex.EncodeToString(m.Sum(nil))
}

func (a *auth) token(now time.Time) string {
	exp := strconv.FormatInt(now.Add(authTTL).Unix(), 10)
	return exp + "." + a.sign(exp)
}

func (a *auth) valid(token string, now time.Time) bool {
	exp, sig, ok := strings.Cut(token, ".")
	if !ok {
		return false
	}
	t, err := strconv.ParseInt(exp, 10, 64)
	if err != nil || now.Unix() >= t {
		return false
	}
	return hmac.Equal([]byte(sig), []byte(a.sign(exp)))
}

func (a *auth) check(user, password string) bool {
	u := subtle.ConstantTimeCompare([]byte(user), []byte(a.user))
	p := subtle.ConstantTimeCompare([]byte(password), []byte(a.password))
	return u&p == 1
}

// middleware lets a request through only with a valid cookie.
// Pages go to /login; API calls get HTTP 401.
func (a *auth) middleware(r *ghttp.Request) {
	if a.valid(r.Cookie.Get(authCookie).String(), time.Now()) {
		r.Middleware.Next()
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/") {
		r.Response.WriteHeader(http.StatusUnauthorized)
		r.Response.WriteJson(ghttp.DefaultHandlerResponse{Code: http.StatusUnauthorized, Message: "נדרשת התחברות"})
		return
	}
	r.Response.RedirectTo("/login")
}

func (a *auth) loginPage(r *ghttp.Request) {
	if err := r.Response.WriteTpl("login.html", g.Map{"failed": r.Get("failed").Bool()}); err != nil {
		g.Log().Errorf(r.Context(), "%+v", err)
	}
}

func (a *auth) login(r *ghttp.Request) {
	if !a.check(r.Get("user").String(), r.Get("password").String()) {
		time.Sleep(time.Second) // ponytail: slows guessing; add a lockout if this faces the internet
		r.Response.RedirectTo("/login?failed=1")
		return
	}
	r.Cookie.SetCookie(authCookie, a.token(time.Now()), "", "/", authTTL, ghttp.CookieOptions{
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
	})
	r.Response.RedirectTo("/")
}
