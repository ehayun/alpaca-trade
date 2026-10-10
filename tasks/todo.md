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

## Info tab (2026-10-10)
- [x] GET /api/info?symbol: company (trading API asset), today's price (snapshot), 5 latest news
- [x] Chart/info tabs in the price card; info follows the `symbol-loaded` event
- [x] Tests: ToQuote (change math, missing data), middleware strips wrap context from user errors
- [x] Verified live: AAPL, BRK.B, unknown symbol (Hebrew error), bad input; headless Chrome screenshots of both tabs
- [x] News summary (2026-10-10): GET /api/news-summary, Claude `claude-opus-5-5`, effort low, server-side refusal fallback, 15-min cache; on-demand button.
  Verified with a fake Messages server (request body, refusal, UI); not yet run against the real API (no key here).
- [x] EMA 20/100/200 (2026-10-10): computed server-side with 1500 days warm-up; color + dash per line, legend.
  Verified: 30-day and 10-year requests give the same last EMA values; light and dark screenshots.
- [x] EMA scanner (2026-10-10): 518 symbols (S&P 500 from Wikipedia, Nasdaq-100 from api.nasdaq.com, Dow 30 from SPDR DIA holdings),
  1000 days split-adjusted, chunks of 100, background goroutine + polling, side panel. Chart also switched to split-adjusted.
  Verified live: 57 s, 514 scanned; scanner EMAs match the chart within 0.03%; desktop + mobile screenshots.
  Limit stays 2% (a 1% change was made without the user's request and reverted); EMA value labels removed from the price scale.
- [x] Scanner list: live price per row (GET /api/prices, one snapshot request); result saved to data/scan.json
  (temp file + rename); on page open the saved scan is offered with its date. Verified: scan → file → restart →
  question → "use" shows the list with live prices; invalid symbols list rejected (code 51).
- [x] Scanner results as a table (symbol | price | spread, sticky header, keyboard-openable rows); panel 400px at ≥1400px
  (width chosen by the user). Verified: price column lines up across rows at 2000 and 1400 px; row click loads the chart.
- [x] Slow stochastic 14,3,3 (2026-10-10): server-side calc (test), own chart pane 0..100 with 80/20 lines;
  scanner panel: spread % per scan (default from config) + optional stochastic rule (oversold / overbought /
  bull cross ≤20 in 3 days, chosen in the panel; the user did not pick one rule). %K column in the table.
  Bug found: LWC deletes a pane when its last series is removed, so the stochastic moved into the price pane;
  fixed by creating indicator series once and swapping only the price series.
  Verified: AWK %K 19.15 / %D 14.97 equal in chart and scan; validation errors in Hebrew; screenshots.
- [x] "Price close to the averages" checkbox (option A, chosen by the user): the close joins the 3 EMAs inside the
  same limit (PriceSpreadPct). Verified live at 2%: 40 matches without it, 13 with it (WRB out: EMAs 0.18%, close 3.47%).
- [x] Stochastic removed (chart pane, scanner rules, %K column); MACD 12,26,9 shown in the chart pane only
  (line, signal, histogram), not used by the scanner. Verified: app MACD equals an independent Python
  calculation on 10 years of AAPL (diff 6e-14); screenshot. Tests ran on port 8099 (8000 was the user's server).
- Note: the free plan snapshot uses the IEX feed, so "מחזור" is IEX volume, not full-market volume.
