package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Health struct {
	Failing         bool `json:"failing"`
	MinResponseTime int  `json:"minResponseTime"`
}

type PaymentRequest struct {
	CorrelationId string          `json:"correlationId"`
	Amount        decimal.Decimal `json:"amount"`
	RequestedAt   time.Time       `json:"requestedAt"`
}

type PaymentBasic struct {
	CorrelationId string          `json:"correlationId"`
	Amount        decimal.Decimal `json:"amount"`
}

type PaymentDb struct {
	CorrelationId string
	Amount        decimal.Decimal
	Fallback      bool
	CreatedAt     time.Time
}

type PaymentSummary struct {
	TotalRequests int     `json:"totalRequests"`
	TotalAmount   float64 `json:"totalAmount"`
}

type SummaryResponse struct {
	Default  PaymentSummary `json:"default"`
	Fallback PaymentSummary `json:"fallback"`
}
