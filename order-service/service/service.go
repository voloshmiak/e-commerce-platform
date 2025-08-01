package service

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"order-service/data"
	pb "order-service/protobuf"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer
}

func (s *OrderService) CreateOrder(_ context.Context, r *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	items := make([]*data.OrderItem, len(r.GetItems()))
	for i, item := range r.GetItems() {
		items[i] = &data.OrderItem{
			ProductID: item.GetProductId(),
			Quantity:  item.GetQuantity(),
			Price:     item.GetPrice(),
		}
	}
	orderID := data.AddOrder(r.GetUserId(), items, r.GetShippingAddress())

	return &pb.CreateOrderResponse{
		OrderId: orderID,
	}, nil
}

func (s *OrderService) GetOrder(_ context.Context, r *pb.GetOrderRequest) (*pb.Order, error) {
	order := data.GetOrderByID(r.GetOrderId())
	if order == nil {
		return nil, nil
	}

	items := make([]*pb.OrderItem, len(order.Items))
	for i, item := range order.Items {
		items[i] = &pb.OrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &pb.Order{
		OrderId:         order.ID,
		UserId:          order.UserID,
		Status:          pb.Status(pb.Status_value[string(order.Status)]),
		Items:           items,
		TotalPrice:      order.TotalPrice,
		ShippingAddress: order.ShippingAddress,
	}, nil
}

func (s *OrderService) ListOrders(_ context.Context, r *pb.ListOrdersRequest) (*pb.ListOrdersResponse, error) {
	orders := data.GetOrdersByUserID(r.GetUserId())

	var ordersResponse []*pb.Order
	for _, order := range orders {
		items := make([]*pb.OrderItem, len(order.Items))
		for i, item := range order.Items {
			items[i] = &pb.OrderItem{
				ProductId: item.ProductID,
				Quantity:  item.Quantity,
				Price:     item.Price,
			}
		}
		ordersResponse = append(ordersResponse, &pb.Order{
			OrderId:         order.ID,
			UserId:          order.UserID,
			Status:          pb.Status(pb.Status_value[string(order.Status)]),
			Items:           items,
			TotalPrice:      order.TotalPrice,
			ShippingAddress: order.ShippingAddress,
		})
	}

	return &pb.ListOrdersResponse{
		Orders: ordersResponse,
	}, nil
}

func (s *OrderService) UpdateOrderStatus(_ context.Context, r *pb.UpdateStatusRequest) (*emptypb.Empty, error) {
	data.UpdateOrderStatus(r.GetOrderId(), data.Status(r.GetStatus()))

	return &emptypb.Empty{}, nil
}
