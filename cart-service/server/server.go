package server

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
	"shopping-cart-service/data"
	pb "shopping-cart-service/protobuf"
	"shopping-cart-service/service"
	"strconv"
)

type ShoppingCartServer struct {
	pb.UnimplementedShoppingCartServiceServer
	Service *service.ShoppingCartService
}

func (s *ShoppingCartServer) GetCart(ctx context.Context, r *emptypb.Empty) (*pb.GetCartResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	fmt.Println(3)

	resItems, err := s.Service.GetCart(int64(userID))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get cart: %v", err))
	}

	items := make([]*pb.CartItem, 0, len(resItems))
	for product, details := range resItems {
		productID, err := strconv.ParseInt(product, 10, 64)
		if err != nil {
			log.Println("error parsing product ID:", err)
			continue
		}

		var jsonMap map[string]interface{}
		err = json.Unmarshal([]byte(details), &jsonMap)
		if err != nil {
			return nil, err
		}

		quantity, err := strconv.ParseInt(jsonMap["Quantity"].(string), 10, 64)
		if err != nil {
			log.Println("error parsing quantity:", err)
		}

		price, err := strconv.ParseFloat(jsonMap["Price"].(string), 64)
		if err != nil {
			log.Println("error parsing price:", err)
		}

		items = append(items, &pb.CartItem{
			ProductId: productID,
			Quantity:  int32(quantity),
			Price:     price,
		})
	}

	return &pb.GetCartResponse{
		Items: items,
	}, nil
}

func (s *ShoppingCartServer) AddItem(ctx context.Context, r *pb.AddItemRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	err = s.Service.AddItem(int64(userID), r.ProductId, r.Quantity, r.Price)
	if err != nil {
		log.Println(err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to add item: %v", err))
	}

	return &emptypb.Empty{}, nil
}

func (s *ShoppingCartServer) UpdateItem(ctx context.Context, r *pb.UpdateItemRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	data.UpdateItemQuantity(int64(userID), r.GetId(), r.GetQuantity())

	return &emptypb.Empty{}, nil
}

func (s *ShoppingCartServer) RemoveItem(ctx context.Context, r *pb.RemoveItemRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}
	data.RemoveItem(int64(userID), r.GetId())

	return &emptypb.Empty{}, nil
}

func (s *ShoppingCartServer) ClearCart(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}
	data.ClearCart(int64(userID))

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
