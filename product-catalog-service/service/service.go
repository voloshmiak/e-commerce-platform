package service

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	"product-catalog-service/data"
	pb "product-catalog-service/protobuf"
)

type ProductCatalogService struct {
	pb.UnimplementedProductCatalogServiceServer
}

func (s *ProductCatalogService) GetProduct(_ context.Context, r *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	product := data.GetProductByID(r.GetId())

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

func (s *ProductCatalogService) ListProducts(_ context.Context, r *emptypb.Empty) (*pb.ListProductsResponse, error) {
	products := data.GetAllProducts()

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

func (s *ProductCatalogService) CreateProduct(_ context.Context, r *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	productID := data.AddProduct(
		r.GetName(),
		r.GetDescription(),
		r.GetPrice(),
		data.Currency(r.GetCurrency().String()),
		r.GetStock(),
		r.GetCategory(),
		r.GetImageUrl(),
		r.GetAttributes(),
	)

	return &pb.CreateProductResponse{Id: productID}, nil
}

func (s *ProductCatalogService) UpdateProduct(_ context.Context, r *pb.UpdateProductRequest) (*emptypb.Empty, error) {
	data.UpdateProduct(
		r.GetId(),
		r.GetName(),
		r.GetDescription(),
		r.GetPrice(),
		data.Currency(r.GetCurrency().String()),
		r.GetStock(),
		r.GetCategory(),
		r.GetImageUrl(),
		r.GetAttributes(),
	)

	return &emptypb.Empty{}, nil
}

func (s *ProductCatalogService) DeleteProduct(_ context.Context, r *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	data.DeleteProduct(r.GetId())

	return &emptypb.Empty{}, nil
}

func (s *ProductCatalogService) CheckStock(_ context.Context, r *pb.CheckStockRequest) (*pb.CheckStockResponse, error) {
	inStock := data.CheckAvailability(r.GetId(), r.GetQuantity())
	return &pb.CheckStockResponse{InStock: inStock}, nil
}
