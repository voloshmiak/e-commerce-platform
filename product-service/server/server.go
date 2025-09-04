package server

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	pb "product-catalog-service/protobuf"
	"product-catalog-service/service"
)

type ProductCatalogServer struct {
	pb.UnimplementedProductCatalogServiceServer
	service *service.ProductCatalogService
}

func NewProductCatalogServer(s *service.ProductCatalogService) *ProductCatalogServer {
	return &ProductCatalogServer{
		service: s,
	}
}

func (s *ProductCatalogServer) GetProduct(_ context.Context, r *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product := s.service.GetProduct(r.GetId())

	return &pb.GetProductResponse{Product: &pb.Product{
		Id:          product.ID,
		Sku:         product.Sku,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Currency:    pb.Currency(pb.Currency_value[string(product.Currency)]),
		Stock:       product.StockQuantity,
		Category:    product.Category,
		ImageUrl:    product.ImageURL,
		IsActive:    product.IsActive,
	}}, nil
}

func (s *ProductCatalogServer) ListProducts(_ context.Context, _ *emptypb.Empty) (*pb.ListProductsResponse, error) {
	products := s.service.ListProducts()

	var productList []*pb.Product
	for _, product := range products {
		productList = append(productList, &pb.Product{
			Id:          product.ID,
			Sku:         product.Sku,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Currency:    pb.Currency(pb.Currency_value[string(product.Currency)]),
			Stock:       product.StockQuantity,
			Category:    product.Category,
			ImageUrl:    product.ImageURL,
			IsActive:    product.IsActive,
		})
	}

	return &pb.ListProductsResponse{Products: productList}, nil
}

func (s *ProductCatalogServer) CreateProduct(_ context.Context, r *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	productID := s.service.CreateProduct(
		r.GetName(),
		r.GetDescription(),
		r.GetPrice(),
		r.GetCurrency().String(),
		r.GetStock(),
		r.GetCategory(),
		r.GetImageUrl(),
		r.GetAttributes(),
	)

	return &pb.CreateProductResponse{Id: productID}, nil
}

func (s *ProductCatalogServer) UpdateProduct(_ context.Context, r *pb.UpdateProductRequest) (*emptypb.Empty, error) {
	s.service.UpdateProduct(
		r.GetId(),
		r.GetName(),
		r.GetDescription(),
		r.GetPrice(),
		r.GetCurrency().String(),
		r.GetStock(),
		r.GetCategory(),
		r.GetImageUrl(),
		r.GetAttributes(),
	)

	return &emptypb.Empty{}, nil
}

func (s *ProductCatalogServer) DeleteProduct(_ context.Context, r *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	s.service.DeleteProduct(r.GetId())

	return &emptypb.Empty{}, nil
}

func (s *ProductCatalogServer) CheckStock(_ context.Context, r *pb.CheckStockRequest) (*pb.CheckStockResponse, error) {
	inStock := s.service.CheckStock(r.GetId(), r.GetQuantity())
	return &pb.CheckStockResponse{InStock: inStock}, nil
}
