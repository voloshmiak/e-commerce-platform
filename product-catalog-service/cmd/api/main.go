package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	pb "product-catalog-service/protobuf"
	"product-catalog-service/service"
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

	s := grpc.NewServer()
	pb.RegisterProductCatalogServiceServer(s, &service.ProductCatalogService{})

	log.Println("Starting gRPC product catalog service server on port :8080")

	if err = s.Serve(listener); err != nil {
		return err
	}

	return nil
}
