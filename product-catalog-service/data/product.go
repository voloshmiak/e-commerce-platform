package data

import (
	"fmt"
	"time"
)

type Currency string

const (
	EUR Currency = "EUR"
	USD Currency = "USD"
	JPY Currency = "JPY"
	BGN Currency = "BGN"
	CZK Currency = "CZK"
	DKK Currency = "DKK"
	GBP Currency = "GBP"
	HUF Currency = "HUF"
	PLN Currency = "PLN"
	RON Currency = "RON"
	SEK Currency = "SEK"
	CHF Currency = "CHF"
	ISK Currency = "ISK"
	NOK Currency = "NOK"
	HRK Currency = "HRK"
	RUB Currency = "RUB"
	TRY Currency = "TRY"
	AUD Currency = "AUD"
	BRL Currency = "BRL"
	CAD Currency = "CAD"
	CNY Currency = "CNY"
	HKD Currency = "HKD"
	IDR Currency = "IDR"
	ILS Currency = "ILS"
	INR Currency = "INR"
	KRW Currency = "KRW"
	MXN Currency = "MXN"
	MYR Currency = "MYR"
	NZD Currency = "NZD"
	PHP Currency = "PHP"
	SGD Currency = "SGD"
	THB Currency = "THB"
	ZAR Currency = "ZAR"
)

type Product struct {
	ID            int64             `json:"id"`
	Sku           string            `json:"sku"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Price         float64           `json:"price"`
	Currency      Currency          `json:"currency"`
	StockQuantity int32             `json:"stock_quantity"`
	Category      string            `json:"category"`
	ImageURL      string            `json:"image_url"`
	Attributes    map[string]string `json:"attributes,omitempty"`
	IsActive      bool              `json:"is_active"`
	CreatedAt     time.Time         `json:"created_at"`
}

var Products = []*Product{
	{
		ID:            1,
		Sku:           "PROD001",
		Name:          "Wireless Mouse",
		Description:   "Ergonomic wireless mouse with adjustable DPI.",
		Price:         29.99,
		Currency:      USD,
		StockQuantity: 150,
		Category:      "Electronics",
		ImageURL:      "https://example.com/images/wireless-mouse.jpg",
		Attributes: map[string]string{
			"color":        "black",
			"connectivity": "wireless",
		},
		IsActive:  true,
		CreatedAt: time.Now(),
	},
	{
		ID:            2,
		Sku:           "PROD002",
		Name:          "Bluetooth Headphones",
		Description:   "Over-ear Bluetooth headphones with noise cancellation.",
		Price:         89.99,
		Currency:      USD,
		StockQuantity: 75,
		Category:      "Electronics",
		ImageURL:      "https://example.com/images/bluetooth-headphones.jpg",
		Attributes: map[string]string{
			"color":        "blue",
			"connectivity": "Bluetooth",
		},
		IsActive:  true,
		CreatedAt: time.Now(),
	},
}

func GetProductByID(id int64) *Product {
	for _, product := range Products {
		if product.ID == id {
			return product
		}
	}
	return nil
}

func GetAllProducts() []*Product {
	return Products
}

func AddProduct(name, description string, price float64, currency Currency, stockQuantity int32, category, imageURL string, attributes map[string]string) int64 {
	newID := int64(len(Products) + 1)
	newProduct := &Product{
		ID:            newID,
		Sku:           "PROD" + fmt.Sprintf("%03d", newID),
		Name:          name,
		Description:   description,
		Price:         price,
		Currency:      currency,
		StockQuantity: stockQuantity,
		Category:      category,
		ImageURL:      imageURL,
		Attributes:    attributes,
		IsActive:      true,
		CreatedAt:     time.Now(),
	}
	Products = append(Products, newProduct)
	return newID
}

func UpdateProduct(id int64, name, description string, price float64, currency Currency, stockQuantity int32, category, imageURL string, attributes map[string]string) {
	for _, product := range Products {
		if product.ID == id {
			product.Name = name
			product.Description = description
			product.Price = price
			product.Currency = currency
			product.StockQuantity = stockQuantity
			product.Category = category
			product.ImageURL = imageURL
			product.Attributes = attributes
			return
		}
	}
}

func DeleteProduct(id int64) {
	for i, product := range Products {
		if product.ID == id {
			Products = append(Products[:i], Products[i+1:]...)
			return
		}
	}
}

func CheckAvailability(id int64, quantity int32) bool {
	for _, product := range Products {
		if product.ID == id {
			return product.StockQuantity >= quantity
		}
	}
	return false
}
