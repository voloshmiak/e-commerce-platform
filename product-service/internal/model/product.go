package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Currency int

const (
	EUR Currency = 0
	USD Currency = 1
)

func (c Currency) String() string {
	return [...]string{"EUR", "USD"}[c]
}

type Product struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Sku           string             `json:"sku" bson:"sku"`
	Name          string             `json:"name" bson:"name"`
	Description   string             `json:"description" bson:"description"`
	Price         float64            `json:"price" bson:"price"`
	Currency      Currency           `json:"currency" bson:"currency"`
	StockQuantity int32              `json:"stock_quantity" bson:"stock_quantity"`
	Category      string             `json:"category" bson:"category"`
	ImageURL      string             `json:"image_url" bson:"image_url"`
	Attributes    map[string]string  `json:"attributes,omitempty" bson:"attributes"`
	IsActive      bool               `json:"is_active" bson:"is_active"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at"`
}
