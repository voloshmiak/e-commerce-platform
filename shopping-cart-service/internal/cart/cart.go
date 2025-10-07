package cart

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	pb "shopping-cart-service/protobuf"
	"time"
)

type ShoppingCart struct {
	redisClient   *redis.Client
	productClient pb.ProductCatalogServiceClient
	ttl           int64
}

type Item struct {
	Quantity       int32
	Price          float64
	Name           string
	ImageURL       string
	ItemTotalPrice float64
}

func New(redisClient *redis.Client, ttl int64, productClient pb.ProductCatalogServiceClient) *ShoppingCart {
	return &ShoppingCart{
		redisClient:   redisClient,
		ttl:           ttl,
		productClient: productClient,
	}
}

func (c *ShoppingCart) GetCart(ctx context.Context, userID int64) (map[string]string, float64, int32, error) {
	key := fmt.Sprintf("cart:%d", userID)
	res, err := c.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, 0.0, 0, err
	}

	var totalPrice float64
	var totalItems int32
	for _, details := range res {
		var CartItem Item
		err = json.Unmarshal([]byte(details), &CartItem)
		if err != nil {
			return nil, 0.0, 0, err
		}

		totalPrice += CartItem.Price * float64(CartItem.Quantity)
		totalItems += CartItem.Quantity
	}

	return res, totalPrice, totalItems, err
}

func (c *ShoppingCart) AddItem(ctx context.Context, userID int64, quantity int32, sku string) error {
	key := fmt.Sprintf("cart:%d", userID)

	getProductResp, err2 := c.productClient.GetProductBySKU(ctx, &pb.GetProductBySKURequest{Sku: sku})
	if err2 != nil {
		return err2
	}

	value := Item{
		Quantity:       quantity,
		Price:          getProductResp.Product.Price,
		Name:           getProductResp.Product.Name,
		ImageURL:       getProductResp.Product.ImageUrl,
		ItemTotalPrice: float64(quantity) * getProductResp.Product.Price,
	}

	encodedValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = c.redisClient.HSet(ctx, key, sku, encodedValue).Result()
	if err != nil {
		return err
	}

	_, err = c.redisClient.Expire(ctx, key, time.Duration(c.ttl)*time.Second).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *ShoppingCart) UpdateItemQuantity(ctx context.Context, userID int64, quantity int32, sku string) error {
	key := fmt.Sprintf("cart:%d", userID)
	item, err := c.redisClient.HGet(ctx, key, sku).Result()
	if err != nil {
		return err
	}

	var value Item

	err = json.Unmarshal([]byte(item), &value)
	if err != nil {
		return err
	}

	value.Quantity = quantity

	encodedValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = c.redisClient.HSet(ctx, key, sku, encodedValue).Result()
	if err != nil {
		return err
	}

	_, err = c.redisClient.Expire(ctx, key, time.Duration(c.ttl)*time.Second).Result()
	if err != nil {
		return err
	}

	return nil

}

func (c *ShoppingCart) RemoveItem(ctx context.Context, userID int64, sku string) {
	key := fmt.Sprintf("cart:%d", userID)
	c.redisClient.HDel(ctx, key, sku)
}

func (c *ShoppingCart) ClearCart(ctx context.Context, userID int64) {
	key := fmt.Sprintf("cart:%d", userID)
	c.redisClient.Del(ctx, key)
}
