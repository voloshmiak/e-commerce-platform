package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type CartRepository interface {
	GetCart(userID string) (map[string]string, error)
	AddItem(userID, productID, quantity, price string) error
	UpdateItem()
	RemoveItem()
	ClearCart()
}

type CartStorage struct {
	client *redis.Client
	ttl    int64
}

func NewCartStorage(client *redis.Client, ttl int64) *CartStorage {
	return &CartStorage{
		client: client,
		ttl:    ttl,
	}
}

type CartItem struct {
	Quantity string
	Price    string
}

func (c *CartStorage) GetCart(userID string) (map[string]string, error) {
	key := fmt.Sprintf("cart:%s", userID)
	res, err := c.client.HGetAll(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}
	fmt.Println(res)
	return res, err
}

func (c *CartStorage) AddItem(userID, productID string, quantity string, price string) error {
	ctx := context.Background()
	key := fmt.Sprintf("cart:%s", userID)

	value := CartItem{
		Quantity: quantity,
		Price:    price,
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = c.client.HSet(ctx, key, productID, valueJSON).Result()
	if err != nil {
		return err
	}
	c.client.Expire(ctx, key, time.Duration(c.ttl)*time.Second)
	return nil
}

func (c *CartStorage) UpdateItem() {
}

func (c *CartStorage) RemoveItem() {
}

func (c *CartStorage) ClearCart() {
}
