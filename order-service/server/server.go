package server

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"order-service/data"
	pb "order-service/protobuf"
	"order-service/service"
	"strconv"
	"strings"
)

const (
	secret = "my-secret-key"
)

type OrderServer struct {
	pb.UnimplementedOrderServiceServer
	Service    *service.OrderService
	CartClient pb.ShoppingCartServiceClient
}

func NewOrderServer(s *service.OrderService, c pb.ShoppingCartServiceClient) *OrderServer {
	return &OrderServer{
		Service:    s,
		CartClient: c,
	}
}

func (s *OrderServer) CreateOrder(ctx context.Context, r *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}
	md := metadata.Pairs("user-id", strconv.Itoa(userID))
	ctx = metadata.NewOutgoingContext(ctx, md)
	resp, err := s.CartClient.GetCart(ctx, &emptypb.Empty{})
	if err != nil {
		log.Println("error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get cart: %v", err))
	}

	items := make([]*data.OrderItem, len(resp.GetItems()))
	for i, item := range resp.GetItems() {
		items[i] = &data.OrderItem{
			ProductID: item.GetProductId(),
			Quantity:  int64(item.GetQuantity()),
			Price:     item.GetPrice(),
		}
	}

	orderID, orderStatus, err := s.Service.CreateOrderSaga(ctx, userID, items, r.GetShippingAddress())
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create order: %v", err))
	}

	return &pb.CreateOrderResponse{
		Id:     int64(orderID),
		Status: orderStatus,
	}, nil
}

func (s *OrderServer) GetOrder(ctx context.Context, r *pb.GetOrderRequest) (*pb.Order, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	order, err := s.Service.GetOrder(userID, r.GetId())
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.NotFound, fmt.Sprintf("failed to get order: %v", err))
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

func (s *OrderServer) ListUserOrders(ctx context.Context, _ *emptypb.Empty) (*pb.ListOrdersResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	orders := s.Service.GetUserOrders(userID)

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

func (s *OrderServer) UpdateOrderStatus(_ context.Context, r *pb.UpdateStatusRequest) (*emptypb.Empty, error) {
	s.Service.UpdateOrderStatus(r.GetId(), string(r.GetStatus()))

	return &emptypb.Empty{}, nil
}

func getUserIDFromContext(ctx context.Context) (int, error) {
	mt, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err := fmt.Errorf("missing metadata in context")
		log.Println(err)
		return 0, err
	}

	bearedToken := mt.Get("Authorization")[0]

	tokenString := strings.TrimPrefix(bearedToken, "Bearer ")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		log.Println(err)
		return 0, err
	}

	claims := token.Claims.(jwt.MapClaims)

	userID, ok := claims["user-id"]
	if !ok {
		err := fmt.Errorf("user-id not found in token claims")
		log.Println(err)
		return 0, err
	}

	userIDFloat, ok := userID.(float64)
	if !ok {
		err := fmt.Errorf("user-id not found in token claims")
		log.Println(err)
		return 0, err
	}

	userIDInt := int(userIDFloat)

	return userIDInt, nil
}
