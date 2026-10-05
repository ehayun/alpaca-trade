# Alpaca stock app

Hebrew (RTL) web app for US stocks: daily price charts and **paper** trading through Alpaca.

- Price history: enter a symbol and a number of days (1 to 3650); candlestick chart by default, with OHLC bars and line as alternatives.
- Paper trading: account summary, market and limit day orders, open positions, open orders with cancel.
- Shareable links: `/?symbol=AAPL&days=90&chart=bars` opens with that chart loaded (`chart` = candles, bars, line).

**Trading always goes to `https://paper-api.alpaca.markets`.** The URL is hard-coded in
`internal/logic/trade/trade.go`, because the SDK otherwise defaults to the live (real money) API.

## Stack

- Go, [GoFrame v2](https://goframe.org/docs) (standard router, templates, validation)
- [alpaca-trade-api-go v3](https://github.com/alpacahq/alpaca-trade-api-go): `marketdata` for prices, `alpaca` for trading
- Bootstrap 5.3 RTL, Heebo font, [Lightweight Charts](https://tradingview.github.io/lightweight-charts/) 5 (all from CDN)
- No database: Alpaca holds orders and positions

## Run

```sh
cp .env.example .env    # fill in your Alpaca PAPER keys
set -a; . ./.env; set +a
make run                # http://localhost:8000
```

| Env var | Purpose |
|---|---|
| `APCA_API_KEY_ID` | Alpaca paper API key ID |
| `APCA_API_SECRET_KEY` | Alpaca paper API secret |

The port and other server settings live in `manifest/config/config.yaml`. No secrets go in that file.

## Test and build

```sh
make test     # go vet + go test
make build    # bin/alpaca-app, a self-contained binary
```

`make build` packs `manifest/config` and `resource/` into the binary with GoFrame's `gres`
(`hack/pack`, the same result as `gf pack` but without installing the CLI), then builds with `-tags packed`.
`make run` reads templates and static files from disk, so edits show up without repacking.

## Layout

```
api/{market,trade}/v1      request/response structs, routes, validation (Hebrew messages)
internal/cmd               server setup, routes, response middleware
internal/controller        thin handlers: request → logic → response
internal/logic             Alpaca calls and business rules (market, trade)
internal/model             shared data shapes
internal/consts            error codes and messages
resource/template          index.html
resource/public            css/, js/ (chart.js, trade.js, common.js)
manifest/config            config.yaml
hack/pack                  resource packer for production builds
```

## API

All JSON responses use the shape `{code, message, data}`. `code` 0 means OK, 51 is a validation error,
and 1001 is a user-facing error, for example an order Alpaca rejected. Any other error is logged
server-side, and the client only gets a generic Hebrew message.

| Method | Path | |
|---|---|---|
| GET | `/api/bars?symbol=AAPL&days=30` | daily candles |
| GET | `/api/account` | equity, cash, buying power |
| GET | `/api/positions` | open positions |
| GET | `/api/orders` | open orders |
| POST | `/api/orders` | `{symbol, side: buy\|sell, type: market\|limit, qty, limitPrice?}` |
| DELETE | `/api/orders/{id}` | cancel an order |
| GET | `/healthz` | liveness |
