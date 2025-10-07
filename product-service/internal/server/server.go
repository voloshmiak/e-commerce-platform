package server

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"product-catalog-service/internal/service"
	pb "product-catalog-service/protobuf"
)

type Server struct {
	pb.UnimplementedProductCatalogServiceServer
	service *service.Service
}

func NewProductCatalogServer(s *service.Service) *Server {
	return &Server{
		service: s,
	}
}

func (s *Server) GetProduct(ctx context.Context, r *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product, err := s.service.GetProduct(ctx, r.GetId())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("product not found: %v", err))
		}
		return nil, err
	}

	return &pb.GetProductResponse{Product: &pb.Product{
		Id:            product.ID.Hex(),
		Sku:           product.Sku,
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		Currency:      pb.Currency(product.Currency),
		StockQuantity: product.StockQuantity,
		Category:      product.Category,
		ImageUrl:      product.ImageURL,
		IsActive:      product.IsActive,
		Attributes:    product.Attributes,
	}}, nil
}

func (s *Server) ListProducts(ctx context.Context, r *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	var query string
	if r.GetQuery() != "" {
		query = r.GetQuery()
	}

	products, err := s.service.ListProducts(ctx, query, r.GetPage(), r.GetPageSize())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var productList []*pb.Product
	for _, product := range products {
		productList = append(productList, &pb.Product{
			Id:            product.ID.Hex(),
			Sku:           product.Sku,
			Name:          product.Name,
			Description:   product.Description,
			Price:         product.Price,
			Currency:      pb.Currency(product.Currency),
			StockQuantity: product.StockQuantity,
			Category:      product.Category,
			ImageUrl:      product.ImageURL,
			IsActive:      product.IsActive,
			Attributes:    product.Attributes,
		})
	}

	return &pb.ListProductsResponse{Products: productList}, nil
}

func (s *Server) CreateProduct(ctx context.Context, r *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	id, err := s.service.CreateProduct(ctx, r)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &pb.CreateProductResponse{Id: id}, nil
}

func (s *Server) UpdateProduct(ctx context.Context, r *pb.UpdateProductRequest) (*pb.UpdateProductResponse, error) {
	err := s.service.UpdateProduct(ctx, r)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &pb.UpdateProductResponse{}, nil
}

func (s *Server) DeleteProduct(ctx context.Context, r *pb.DeleteProductRequest) (*pb.DeleteProductResponse, error) {
	err := s.service.DeleteProduct(ctx, r.GetId())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("product not found: %v", err))
		}
		log.Println(err)
		return nil, err
	}

	return &pb.DeleteProductResponse{}, nil
}

func (s *Server) GetProductBySKU(ctx context.Context, r *pb.GetProductBySKURequest) (*pb.GetProductBySKUResponse, error) {
	product, err := s.service.GetProductBySKU(ctx, r.GetSku())
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, fmt.Sprintf("product not found: %v", err))
		}
		return nil, err
	}

	return &pb.GetProductBySKUResponse{Product: &pb.Product{
		Id:            product.ID.Hex(),
		Sku:           product.Sku,
		Name:          product.Name,
		Description:   product.Description,
		Price:         product.Price,
		Currency:      pb.Currency(product.Currency),
		StockQuantity: product.StockQuantity,
		Category:      product.Category,
		ImageUrl:      product.ImageURL,
		IsActive:      product.IsActive,
		Attributes:    product.Attributes,
	}}, nil
}
