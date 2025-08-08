package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"order-service/data"
	pb "order-service/protobuf"
	"strconv"
)

type OrderService struct {
	pb.UnimplementedOrderServiceServer
}

func (s *OrderService) CreateOrder(ctx context.Context, r *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	items := make([]*data.OrderItem, len(r.GetItems()))
	for i, item := range r.GetItems() {
		items[i] = &data.OrderItem{
			ProductID: item.GetProductId(),
			Quantity:  item.GetQuantity(),
			Price:     item.GetPrice(),
		}
	}
	orderID := data.AddOrder(int64(userID), items, r.GetShippingAddress())

	return &pb.CreateOrderResponse{
		Id: orderID,
	}, nil
}

func (s *OrderService) GetOrder(ctx context.Context, r *pb.GetOrderRequest) (*pb.Order, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	order := data.GetUserOrderByID(int64(userID), r.GetId())
	if order == nil {
		return nil, status.Error(codes.NotFound, "order not found")
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
		Id:              order.ID,
		UserId:          order.UserID,
		Status:          pb.Status(pb.Status_value[string(order.Status)]),
		Items:           items,
		TotalPrice:      order.TotalPrice,
		ShippingAddress: order.ShippingAddress,
	}, nil
}

func (s *OrderService) ListUserOrders(ctx context.Context, _ *emptypb.Empty) (*pb.ListOrdersResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	orders := data.GetOrdersByUserID(int64(userID))

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
			Id:              order.ID,
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
	data.UpdateOrderStatus(r.GetId(), data.Status(r.GetStatus()))

	return &emptypb.Empty{}, nil
}

func getUserIDFromContext(ctx context.Context) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "missing metadata in context")
	}

	userID := md["user-id"][0]

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return 0, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user ID: %v", err))
	}

	return userIDInt, nil
}
