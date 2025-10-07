package model

import (
	"time"
)

type Status string

const (
	Pending   Status = "PENDING"
	Paid      Status = "PAID"
	Confirmed Status = "CONFIRMED"
	Cancelled Status = "CANCELLED"
)

type Order struct {
	ID              int64        `json:"id"`
	UserID          int64        `json:"user_id"`
	Status          Status       `json:"status"`
	Items           []*OrderItem `json:"items"`
	TotalPrice      float64      `json:"total_price"`
	ShippingAddress string       `json:"shipping_address"`
	CreatedAt       time.Time    `json:"created_at"`
}

type OrderItem struct {
	ID       int64   `json:"id"`
	OrderID  int64   `json:"order_id"`
	Quantity int64   `json:"quantity"`
	Price    float64 `json:"price"`
	Sku      string  `json:"sku"`
}
