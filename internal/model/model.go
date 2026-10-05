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
