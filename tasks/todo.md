# Alpaca stock app — GoFrame + Bootstrap RTL (Hebrew)

Decisions: paper trading only (hard-coded paper URL), no DB, Lightweight Charts 5.2.1 (candles default, line toggle),
Bootstrap 5.3.8 RTL, Heebo, Hebrew UI, `$` money, dates dd/MM/yyyy Asia/Jerusalem, English README.

## Plan
- [x] go.mod + GoFrame v2.10.3 + alpaca-trade-api-go v3.11.0
- [x] manifest/config/config.yaml (address, viewer autoEncode, serverRoot; no secrets)
- [x] api/market/v1: GET /api/bars?symbol&days (validated, Hebrew messages)
- [x] api/trade/v1: account, positions, orders, place, cancel
- [x] internal/logic (Alpaca clients, bars → candles, order building, 4xx → user error)
- [x] internal/controller (thin)
- [x] internal/cmd: routes, GET / page, GET /healthz, safe response middleware
- [x] resource/template/index.html + resource/public/js/{common,chart,trade}.js + css
- [x] Resource packing into binary (hack/pack + -tags packed)
- [x] Tests for logic + middleware; go vet / go test
- [x] .env.example, .gitignore, README.md
- [x] Run, screenshot RTL page (desktop + 500px), place + cancel a paper limit order end-to-end

## Review
- Verified live against paper API: bars, validation (Hebrew), account, insufficient-buying-power rejection,
  limit order placed (AAPL 1 @ $1) and cancelled; account left clean.
- Bugs found while verifying, all fixed:
  - GoFrame wrote panic text into the response; the middleware now clears the buffer on error (covered by test).
  - GoFrame has no `uuid` validation rule; replaced with a regex.
  - The chart title's bidi order was scrambled; the symbol is now wrapped in `<bdi>`.
- GoFrame docs (goframe.org) block automated fetches (403); they were read from the gf-site source repo instead.
- Not done: live trading switch, stop orders, crypto, port override via env (port is in config.yaml).
