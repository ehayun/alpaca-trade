// Package model holds the data shapes shared by logic and the API.
package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// Candle is one trading day, shaped for Lightweight Charts.
type Candle struct {
	Time   string  `json:"time"` // trading day, YYYY-MM-DD in New York
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume uint64  `json:"volume"`
}

// Point is one value of a line on the chart.
type Point struct {
	Time  string  `json:"time"`
	Value float64 `json:"value"`
}

// EMA is an exponential moving average of the close, aligned with the candles.
type EMA struct {
	Period int     `json:"period"`
	Points []Point `json:"points"`
}

// MACD is the MACD line (fast EMA - slow EMA of the close), its signal line (EMA of MACD)
// and the histogram (MACD - signal), aligned with the candles.
type MACD struct {
	MACD   []Point `json:"macd"`
	Signal []Point `json:"signal"`
	Hist   []Point `json:"hist"`
}

// ScanMatch is one symbol's last close and EMAs; SpreadPct = (highest EMA - lowest EMA) / close * 100.
// PriceSpreadPct is the same with the close as a fourth value.
type ScanMatch struct {
	Symbol         string  `json:"symbol"`
	Close          float64 `json:"close"`
	EMA20          float64 `json:"ema20"`
	EMA100         float64 `json:"ema100"`
	EMA200         float64 `json:"ema200"`
	SpreadPct      float64 `json:"spreadPct"`
	PriceSpreadPct float64 `json:"priceSpreadPct"`
}

// Spread is the value a scan compares with its limit: with or without the close.
func (m ScanMatch) Spread(withPrice bool) float64 {
	if withPrice {
		return m.PriceSpreadPct
	}
	return m.SpreadPct
}

// ScanStatus is the state of the EMA scanner: progress while running, matches when done.
type ScanStatus struct {
	Running      bool        `json:"running"`
	Failed       bool        `json:"failed"`
	StartedAt    *time.Time  `json:"startedAt"`
	FinishedAt   *time.Time  `json:"finishedAt"`
	Total        int         `json:"total"`   // symbols in the index lists
	Done         int         `json:"done"`    // symbols fetched so far
	Scanned      int         `json:"scanned"` // symbols with enough history for EMA 200
	MaxSpreadPct float64     `json:"maxSpreadPct"`
	PriceClose   bool        `json:"priceClose"` // the close must also be inside the spread limit
	Matches      []ScanMatch `json:"matches"`    // closest first
}

// Decimals marshal as JSON strings to keep exact values.

type Account struct {
	Equity      decimal.Decimal `json:"equity"`
	Cash        decimal.Decimal `json:"cash"`
	BuyingPower decimal.Decimal `json:"buyingPower"`
}

type Position struct {
	Symbol        string           `json:"symbol"`
	Qty           decimal.Decimal  `json:"qty"`
	AvgEntryPrice decimal.Decimal  `json:"avgEntryPrice"`
	CurrentPrice  *decimal.Decimal `json:"currentPrice"`
	MarketValue   *decimal.Decimal `json:"marketValue"`
	UnrealizedPL  *decimal.Decimal `json:"unrealizedPl"`
}

type Order struct {
	ID         string           `json:"id"`
	CreatedAt  time.Time        `json:"createdAt"`
	Symbol     string           `json:"symbol"`
	Side       string           `json:"side"`
	Type       string           `json:"type"`
	Qty        *decimal.Decimal `json:"qty"`
	LimitPrice *decimal.Decimal `json:"limitPrice"`
	Status     string           `json:"status"`
}

// OrderInput is a validated order from the user.
type OrderInput struct {
	Symbol     string
	Side       string // buy | sell
	Type       string // market | limit
	Qty        string
	LimitPrice string
}

// StockInfo feeds the info tab: company, today's price and recent news.
type StockInfo struct {
	Symbol string `json:"symbol"`
	Asset  Asset  `json:"asset"`
	Quote  Quote  `json:"quote"`
	News   []News `json:"news"`
}

type Asset struct {
	Name         string `json:"name"`
	Exchange     string `json:"exchange"`
	Tradable     bool   `json:"tradable"`
	Shortable    bool   `json:"shortable"`
	Fractionable bool   `json:"fractionable"`
}

// Quote fields are nil when Alpaca has no data for them.
type Quote struct {
	Price     *float64 `json:"price"`
	PrevClose *float64 `json:"prevClose"`
	Change    *float64 `json:"change"`
	ChangePct *float64 `json:"changePct"`
	Open      *float64 `json:"open"`
	High      *float64 `json:"high"`
	Low       *float64 `json:"low"`
	Volume    *uint64  `json:"volume"`
}

type News struct {
	Headline  string    `json:"headline"`
	Summary   string    `json:"-"` // only sent to Claude for the news summary
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
}
