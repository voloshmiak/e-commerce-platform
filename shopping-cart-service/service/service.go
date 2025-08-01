package service

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"shopping-cart-service/data"
	pb "shopping-cart-service/protobuf"
)

type ShoppingCartService struct {
	pb.UnimplementedShoppingCartServiceServer
}

func (s *ShoppingCartService) GetCart(_ context.Context, r *pb.GetCartRequest) (*pb.GetCartResponse, error) {
	cart := data.GetCart(r.UserId)

	items := make([]*pb.CartItem, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = &pb.CartItem{
			ProductId: item.ProductID,
			Quantity:  int32(item.Quantity),
			Price:     item.Price,
		}
	}

	return &pb.GetCartResponse{
		Items: items,
	}, nil
}

func (s *ShoppingCartService) AddItem(_ context.Context, r *pb.AddItemRequest) (*emptypb.Empty, error) {
	item := r.GetItem()
	data.AddItem(r.GetUserId(), item.GetProductId(), item.GetQuantity(), item.GetPrice())

	return &emptypb.Empty{}, nil
}

func (s *ShoppingCartService) UpdateItem(_ context.Context, r *pb.UpdateItemRequest) (*emptypb.Empty, error) {
	item := r.GetItem()
	data.UpdateItem(r.GetUserId(), item.GetProductId(), item.GetQuantity(), item.GetPrice())

	return &emptypb.Empty{}, nil
}

func (s *ShoppingCartService) RemoveItem(_ context.Context, r *pb.RemoveItemRequest) (*emptypb.Empty, error) {
	data.RemoveItem(r.GetUserId(), r.GetProductId())

	return &emptypb.Empty{}, nil
}

func (s *ShoppingCartService) ClearCart(_ context.Context, r *pb.ClearCartRequest) (*emptypb.Empty, error) {
	data.ClearCart(r.GetUserId())

	return &emptypb.Empty{}, nil
}
