package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"product-catalog-service/internal/config"
	"product-catalog-service/internal/service"
	"sync"
)

type HandlerFunc func(ctx context.Context, m *kafka.Message) error

type Consumer struct {
	service *service.Service
	config  *config.Config
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func New(svc *service.Service, cfg *config.Config) *Consumer {
	return &Consumer{
		service: svc,
		config:  cfg,
	}
}

func (c *Consumer) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	listeners := map[string]HandlerFunc{
		"payment.succeeded": c.handlePaymentSucceed,
	}

	for topic, handler := range listeners {
		groupID := "product-catalog-service-" + topic

		c.wg.Add(1)
		go c.listen(ctx, topic, groupID, handler)
	}
}

func (c *Consumer) Stop() {
	log.Println("Stopping Kafka consumers...")
	c.cancel()
	c.wg.Wait()
	log.Println("All Kafka consumers stopped.")
}

func (c *Consumer) listen(ctx context.Context, topic, groupID string, handler HandlerFunc) {
	defer c.wg.Done()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{fmt.Sprintf("%s:%s", c.config.Kafka.Host, c.config.Kafka.Port)},
		Topic:    topic,
		GroupID:  groupID,
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Printf("Listening for %s messages...\n", topic)

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			continue
		}

		go func(m kafka.Message) {
			err = handler(ctx, &m)
			if err != nil {
				log.Println(err)
			}
		}(m)
	}
}

func (c *Consumer) handlePaymentSucceed(ctx context.Context, m *kafka.Message) error {
	var event service.OrderCreatedEvent
	err := json.Unmarshal(m.Value, &event)
	if err != nil {
		return err
	}

	err = c.service.CheckAndReserveStock(ctx, event.Data)

	return err
}

func (c *Consumer) handlePaymentFailed(ctx context.Context, m *kafka.Message) error {
	var event service.OrderCreatedEvent
	err := json.Unmarshal(m.Value, &event)
	if err != nil {
		return err
	}

	err = c.service.CompensateStock(ctx, event.Data)

	return err
}
