package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"order-service/data"
	"strconv"
)

type OrderService struct{}

func (s *OrderService) CreateOrder(ctx context.Context, userID int, items []*data.OrderItem, shippingAddress string) (int, error) {
	order := data.AddOrder(int64(userID), items, shippingAddress)

	if err := produceMessage(ctx, order, strconv.Itoa(int(order.ID))); err != nil {
		log.Println(err)
		return 0, err
	}

	return int(order.ID), nil
}

func (s *OrderService) GetOrder(userID int, orderID int64) (*data.Order, error) {
	order := data.GetUserOrderByID(int64(userID), orderID)
	if order == nil {
		return nil, fmt.Errorf("order with ID %d not found for user %d", orderID, userID)
	}

	return order, nil
}

func (s *OrderService) GetUserOrders(userID int) []*data.Order {
	orders := data.GetOrdersByUserID(int64(userID))

	return orders
}

func (s *OrderService) UpdateOrderStatus(orderID int64, status string) {
	data.UpdateOrderStatus(orderID, data.Status(status))
}

func produceMessage(ctx context.Context, toSend any, key string) error {
	encodedOrder, err := json.Marshal(toSend)
	if err != nil {
		return fmt.Errorf("failed to encode order: %v", err)
	}

	w := &kafka.Writer{
		Addr:                   kafka.TCP("kafka"),
		Topic:                  "orders.created",
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer w.Close()

	err = w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: encodedOrder,
	})

	if err != nil {
		log.Println(err)
		return fmt.Errorf("failed to write message to Kafka: %v", err)
	}

	return nil
}
