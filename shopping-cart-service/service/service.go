package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"shopping-cart-service/data"
	pb "shopping-cart-service/protobuf"
	"strconv"
)

type ShoppingCartService struct {
	pb.UnimplementedShoppingCartServiceServer
}

func (s *ShoppingCartService) GetCart(ctx context.Context, _ *emptypb.Empty) (*pb.GetCartResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	cart := data.GetCart(int64(userID))

	items := make([]*pb.CartItem, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = &pb.CartItem{
			Id:        item.ID,
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
	}

	return &pb.GetCartResponse{
		Items: items,
	}, nil
}

func (s *ShoppingCartService) AddItem(ctx context.Context, r *pb.AddItemRequest) (*pb.AddItemResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	itemID := data.AddItem(int64(userID), r.GetProductId(), r.GetQuantity(), r.GetPrice())

	return &pb.AddItemResponse{
		Id: itemID,
	}, nil
}

func (s *ShoppingCartService) UpdateItem(ctx context.Context, r *pb.UpdateItemRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}

	data.UpdateItemQuantity(int64(userID), r.GetId(), r.GetQuantity())

	return &emptypb.Empty{}, nil
}

func (s *ShoppingCartService) RemoveItem(ctx context.Context, r *pb.RemoveItemRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to get user ID from context")
	}
	data.RemoveItem(int64(userID), r.GetId())

	return &emptypb.Empty{}, nil
}

func (s *ShoppingCartService) ClearCart(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
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
