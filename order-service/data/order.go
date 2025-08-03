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
	ID              int64        `json:"id"`
	UserID          int64        `json:"user_id"`
	Status          Status       `json:"status"`
	Items           []*OrderItem `json:"items"`
	TotalPrice      float64      `json:"total_price"`
	ShippingAddress string       `json:"shipping_address"`
	CreatedAt       time.Time    `json:"created_at"`
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
		UserID:          2,
		Status:          Paid,
		Items:           []*OrderItem{{ID: 1, OrderID: 1, ProductID: 101, Quantity: 2, Price: 50.00}},
		TotalPrice:      100.00,
		ShippingAddress: "123 Main St, Anytown, USA",
		CreatedAt:       time.Now(),
	},
	{
		ID:              2,
		UserID:          2,
		Status:          Shipped,
		Items:           []*OrderItem{{ID: 2, OrderID: 2, ProductID: 102, Quantity: 1, Price: 75.00}},
		TotalPrice:      75.00,
		ShippingAddress: "456 Elm St, Othertown, USA",
		CreatedAt:       time.Now(),
	},
}

func AddOrder(userID int64, items []*OrderItem, shippingAddress string) int64 {
	order := &Order{
		ID:              int64(len(Orders) + 1),
		UserID:          userID,
		Status:          Pending,
		Items:           items,
		TotalPrice:      calculateTotalPrice(items),
		ShippingAddress: shippingAddress,
		CreatedAt:       time.Now(),
	}

	Orders = append(Orders, order)
	return order.ID
}

func GetOrderByID(orderID int64) *Order {
	for _, order := range Orders {
		if order.ID == orderID {
			return order
		}
	}
	return nil
}

func GetOrdersByUserID(userID int64) []*Order {
	var userOrders []*Order
	for _, order := range Orders {
		if order.UserID == userID {
			userOrders = append(userOrders, order)
		}
	}
	return userOrders
}

func UpdateOrderStatus(orderID int64, status Status) {
	for _, order := range Orders {
		if order.ID == orderID {
			order.Status = status
		}
	}
}

func calculateTotalPrice(items []*OrderItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}
