package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"product-catalog-service/internal/config"
	"product-catalog-service/internal/consumer"
	"product-catalog-service/internal/repository"
	"product-catalog-service/internal/server"
	"product-catalog-service/internal/service"
	pb "product-catalog-service/protobuf"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Config
	cfg, err := config.New()
	if err != nil {
		return err
	}

	// MongoDB client
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=admin", cfg.Mongo.User,
		cfg.Mongo.Password, cfg.Mongo.Host, cfg.Mongo.Port)
	mongoClient, err := newMongoClient(uri)
	if err != nil {
		return err
	}
	defer mongoClient.Disconnect(context.Background())

	// Elasticsearch client
	addr := fmt.Sprintf("http://%s:%s", cfg.Elastic.Host, cfg.Elastic.Port)
	elasticClient, err := newElasticClient(addr)
	if err != nil {
		return err
	}

	// Listener
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Server.Port))
	if err != nil {
		return err
	}

	// Repository
	mongoRepo := repository.NewMongoRepository(mongoClient.Database(cfg.Mongo.DBName).Collection("products"))
	elasticRepo := repository.NewElasticRepository(elasticClient, "products_idx")

	// Kafka writers
	kafkaAddr := fmt.Sprintf("%s:%s", cfg.Kafka.Host, cfg.Kafka.Port)
	stockReservedWriter := &kafka.Writer{
		Addr:                   kafka.TCP(kafkaAddr),
		Topic:                  "stock.reserved",
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer stockReservedWriter.Close()

	stockFailedWriter := &kafka.Writer{
		Addr:                   kafka.TCP(kafkaAddr),
		Topic:                  "stock.reservation.failed",
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer stockFailedWriter.Close()

	// Service
	svc := service.New(mongoRepo, elasticRepo, stockReservedWriter, stockFailedWriter)

	// gRPC server
	s := grpc.NewServer()
	pb.RegisterProductCatalogServiceServer(s, server.NewProductCatalogServer(svc))

	// Kafka consumer
	cons := consumer.New(svc, cfg)
	cons.Start()

	// Serving gRPC server
	go func() {
		log.Printf("Starting gRPC user service server on port %s\n", cfg.Server.Port)

		if err = s.Serve(listener); err != nil {
			log.Println(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Received shutdown signal, stopping server...")

	s.GracefulStop()
	cons.Stop()

	log.Println("Application stopped")

	return nil
}

func newMongoClient(uri string) (*mongo.Client, error) {
	ctx := context.Background()
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func newElasticClient(addr string) (*elasticsearch.Client, error) {
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{addr},
	})
	if err != nil {
		return nil, err
	}

	res, err := client.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.New(res.String())
	}

	_, err = client.Indices.Create("products_idx")
	if err != nil {
		return nil, err
	}

	return client, nil
}
