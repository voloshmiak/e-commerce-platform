package service

import (
	"product-catalog-service/data"
)

type ProductCatalogService struct {
}

func (s *ProductCatalogService) GetProduct(id int64) *data.Product {
	product := data.GetProductByID(id)

	return product
}

func (s *ProductCatalogService) ListProducts() []*data.Product {
	products := data.GetAllProducts()

	return products
}

func (s *ProductCatalogService) CreateProduct(name, description string, price float64, currency string, stockQuantity int32, category, imageURL string, attributes map[string]string) int64 {
	productID := data.AddProduct(
		name,
		description,
		price,
		data.Currency(currency),
		stockQuantity,
		category,
		imageURL,
		attributes,
	)

	return productID
}

func (s *ProductCatalogService) UpdateProduct(id int64, name, description string, price float64, currency string, stockQuantity int32, category, imageURL string, attributes map[string]string) {
	data.UpdateProduct(
		id,
		name,
		description,
		price,
		data.Currency(currency),
		stockQuantity,
		category,
		imageURL,
		attributes,
	)
}

func (s *ProductCatalogService) DeleteProduct(id int64) {
	data.DeleteProduct(id)
}

func (s *ProductCatalogService) CheckStock(id int64, stock int32) bool {
	return data.CheckAvailability(id, stock)
}
