package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	pb "payment-service/protobuf"
	"strconv"
	"time"
)

func ListenForOrderCreated() error {
	time.Sleep(30 * time.Second)
	conn, err := grpc.NewClient("localhost:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	client := pb.NewPaymentServiceClient(conn)

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"kafka"},
		Topic:    "orders.created",
		MaxBytes: 10e6,
	})
	defer r.Close()

	log.Println("Listening for  order creation messages..")

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		var order map[string]interface{}
		err = json.Unmarshal(m.Value, &order)
		if err != nil {
			log.Println(err)
			break
		}

		fmt.Println(order)

		_, err = client.ProcessPayment(context.Background(), &pb.ProcessPaymentRequest{
			OrderId:       int64(order["id"].(float64)),
			Amount:        order["total_price"].(float64),
			Currency:      1,
			PaymentMethod: 1,
		})
		if err != nil {
			w := &kafka.Writer{
				Addr:                   kafka.TCP("kafka"),
				Topic:                  "payment.failed",
				Balancer:               &kafka.LeastBytes{},
				AllowAutoTopicCreation: true,
			}
			defer w.Close()

			err = w.WriteMessages(context.Background(), kafka.Message{
				Key:   []byte(strconv.Itoa(int(order["id"].(float64)))),
				Value: []byte(fmt.Sprintf("Payment failed for order %d: %v", int(order["id"].(float64)), err)),
			})

			if err != nil {
				return fmt.Errorf("failed to write message to Kafka: %v", err)
			}
		}

		w := &kafka.Writer{
			Addr:                   kafka.TCP("kafka"),
			Topic:                  "payment.succeeded",
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		}
		defer w.Close()

		encodedOrder, err := json.Marshal(order)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("failed to marshal order: %v", err)
		}

		err = w.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(strconv.Itoa(int(order["id"].(float64)))),
			Value: encodedOrder,
		})

		if err != nil {
			return fmt.Errorf("failed to write message to Kafka: %v", err)
		}
	}
	return nil
}
