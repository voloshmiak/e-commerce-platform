package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	pb "order-service/protobuf"
	"order-service/service"
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
	pb.RegisterOrderServiceServer(s, &service.OrderService{})

	log.Println("Starting gRPC order service server on port :8080")

	if err = s.Serve(listener); err != nil {
		return err
	}

	return nil
}
