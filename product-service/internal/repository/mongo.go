package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"product-catalog-service/internal/model"
	"time"
)

type MongoRepository struct {
	MongoCollection *mongo.Collection
}

func NewMongoRepository(mongoCollection *mongo.Collection) *MongoRepository {
	return &MongoRepository{
		MongoCollection: mongoCollection,
	}
}

func (r *MongoRepository) GetProductByID(ctx context.Context, id string) (*model.Product, error) {
	var product model.Product
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	err = r.MongoCollection.FindOne(ctx, primitive.M{"_id": objID}).Decode(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *MongoRepository) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {
	var product model.Product

	fmt.Println(sku)

	err := r.MongoCollection.FindOne(ctx, primitive.M{"sku": sku}).Decode(&product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (r *MongoRepository) CreateProduct(ctx context.Context, product *model.Product) (string, error) {
	_, err := r.MongoCollection.InsertOne(ctx, product)
	if err != nil {
		return "", err
	}

	hexedObjectID := product.ID.Hex()

	return hexedObjectID, nil
}

func (r *MongoRepository) UpdateProduct(ctx context.Context, product *model.Product) (*mongo.UpdateResult, error) {
	update := primitive.M{
		"$set": primitive.M{
			"sku":            product.Sku,
			"name":           product.Name,
			"description":    product.Description,
			"price":          product.Price,
			"currency":       product.Currency,
			"stock_quantity": product.StockQuantity,
			"category":       product.Category,
			"image_url":      product.ImageURL,
			"attributes":     product.Attributes,
			"is_active":      product.IsActive,
			"updated_at":     time.Now(),
		},
	}

	result, err := r.MongoCollection.UpdateByID(ctx, product.ID, update)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *MongoRepository) DeleteProductByID(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.MongoCollection.DeleteOne(ctx, primitive.M{"_id": objID})
	if err != nil {
		return err
	}

	return nil
}

func (r *MongoRepository) BulkGetBySKUs(ctx context.Context, skus []string) ([]*model.Product, error) {
	filter := bson.M{"sku": bson.M{"$in": skus}}

	cursor, err := r.MongoCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*model.Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	if err = cursor.Err(); err != nil {
		return nil, err
	}

	if products == nil {
		return []*model.Product{}, nil
	}

	return products, nil
}
