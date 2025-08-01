package data

import (
	"time"
)

type Status string

const (
	Pending   Status = "pending"
	Paid      Status = "paid"
	Shipped   Status = "shipped"
	Cancelled Status = "cancelled"
)

type Order struct {
	ID              int64       `json:"id"`
	UserID          int64       `json:"user_id"`
	Status          Status      `json:"status"`
	Items           []OrderItem `json:"items"`
	TotalPrice      float64     `json:"total_price"`
	ShippingAddress string      `json:"shipping_address"`
	CreatedAt       time.Time   `json:"created_at"`
}

type OrderItem struct {
	ID        int64   `json:"id"`
	OrderID   int64   `json:"order_id"`
	ProductID int64   `json:"product_id"`
	Quantity  int64   `json:"quantity"`
	Price     float64 `json:"price"`
}

var Orders = []*Order{
	{
		ID:              1,
		UserID:          1,
		Status:          Paid,
		Items:           []OrderItem{{ID: 1, OrderID: 1, ProductID: 101, Quantity: 2, Price: 50.00}},
		TotalPrice:      100.00,
		ShippingAddress: "123 Main St, Anytown, USA",
		CreatedAt:       time.Now(),
	},
	{
		ID:              2,
		UserID:          2,
		Status:          Shipped,
		Items:           []OrderItem{{ID: 2, OrderID: 2, ProductID: 102, Quantity: 1, Price: 75.00}},
		TotalPrice:      75.00,
		ShippingAddress: "456 Elm St, Othertown, USA",
		CreatedAt:       time.Now(),
	},
}
