package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"product-catalog-service/consumer"
	pb "product-catalog-service/protobuf"
	"product-catalog-service/server"
	"product-catalog-service/service"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	svc := &service.ProductCatalogService{}
	cons := &consumer.ProductConsumer{Service: svc}

	s := grpc.NewServer()
	pb.RegisterProductCatalogServiceServer(s, server.NewProductCatalogServer(svc))

	// Kafka consumer
	go cons.ListenForOrderCreated()
	go cons.ListenForFailedPayment()

	// gRPC server
	go func() {
		log.Println("Starting gRPC product catalog service server on port :8080")

		if err = s.Serve(listener); err != nil {
			log.Println(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Received shutdown signal, stopping server...")

	s.GracefulStop()

	return nil
}
