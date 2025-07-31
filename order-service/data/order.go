package data

import (
	pb "order-service/protobuf"
)

var orders = []*pb.OrderSummary{
	{
		OrderId: 1,
		UserId:  123,
		Items: []*pb.OrderItem{
			{ProductId: 1, Quantity: 2, Price: 19.99},
			{ProductId: 2, Quantity: 1, Price: 29.99},
		},
		Status: "Pending",
	},
	{
		OrderId: 2,
		UserId:  789,
		Items: []*pb.OrderItem{
			{ProductId: 3, Quantity: 1, Price: 49.99},
			{ProductId: 4, Quantity: 3, Price: 15.99},
		},
		Status: "Shipped",
	},
}

func GetOrderByID(id int64) *pb.OrderSummary {
	for _, order := range orders {
		if order.OrderId == id {
			return order
		}
	}
	return nil
}

func GetOrdersByUserID(userID int64) []*pb.OrderSummary {
	var userOrders []*pb.OrderSummary
	for _, order := range orders {
		if order.UserId == userID {
			userOrders = append(userOrders, order)
		}
	}
	return userOrders
}

func CreateOrder(userID int64, items []*pb.OrderItem) (int64, string) {
	order := &pb.OrderSummary{
		OrderId: int64(len(orders) + 1),
		UserId:  userID,
		Items:   items,
		Status:  "Pending",
	}

	orders = append(orders, order)

	return order.OrderId, order.Status
}

func UpdateOrderStatus(orderID int64, status string) (int64, string) {
	for _, order := range orders {
		if order.OrderId == orderID {
			order.Status = status
			return order.OrderId, status
		}
	}
	return 0, ""
}
