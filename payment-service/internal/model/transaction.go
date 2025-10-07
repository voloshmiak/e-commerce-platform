package model

import (
	"time"
)

type Status string

const (
	Pending   Status = "pending"
	Completed Status = "completed"
	Failed    Status = "failed"
	Refunded  Status = "refunded"
)

type PaymentMethod string

const (
	Card       PaymentMethod = "CARD"
	OnDelivery PaymentMethod = "ON_DELIVERY"
)

type Currency string

const (
	EUR Currency = "EUR"
	USD Currency = "USD"
)

type Transaction struct {
	ID                   int64         `json:"id"`
	OrderID              *int64        `json:"order_id"`
	Amount               float64       `json:"amount"`
	Currency             Currency      `json:"currency"`
	Status               Status        `json:"Status"`
	GatewayTransactionID *string       `json:"gateway_transaction_id"`
	PaymentMethod        PaymentMethod `json:"payment_method"`
	CreatedAt            time.Time     `json:"created_at"`
}
