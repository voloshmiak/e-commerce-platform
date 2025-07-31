package service

import (
	"context"
	"order-service/data"
	pb "order-service/protobuf"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer
}

func (s *OrderService) CreateOrder(_ context.Context, r *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	orderID, status := data.CreateOrder(r.GetUserId(), r.GetItems())

	return &pb.CreateOrderResponse{
		OrderId: orderID,
		Status:  status,
	}, nil
}

func (s *OrderService) GetOrder(_ context.Context, r *pb.GetOrderRequest) (*pb.OrderSummary, error) {
	order := data.GetOrderByID(r.GetOrderId())
	if order == nil {
		return nil, nil
	}

	return &pb.OrderSummary{
		OrderId: order.OrderId,
		UserId:  order.UserId,
		Items:   order.Items,
		Status:  order.Status,
	}, nil
}

func (s *OrderService) ListOrders(_ context.Context, r *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders := data.GetOrdersByUserID(r.GetUserId())

	var orderSummaries []*pb.OrderSummary
	for _, order := range orders {
		orderSummaries = append(orderSummaries, &pb.OrderSummary{
			OrderId: order.OrderId,
			UserId:  order.UserId,
			Items:   order.Items,
			Status:  order.Status,
		})
	}

	return &pb.ListOrdersResponse{
		Orders: orderSummaries,
	}, nil
}

func (s *OrderService) UpdateOrderStatus(_ context.Context, r *pb.UpdateStatusRequest) (*pb.UpdateStatusResponse, error) {
	orderID, status := data.UpdateOrderStatus(r.GetOrderId(), r.GetStatus())

	return &pb.UpdateStatusResponse{
		OrderId: orderID,
		Status:  status,
	}, nil
}
