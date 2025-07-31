package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	pb "user-service/protobuf"
	"user-service/service"
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
	pb.RegisterUserServiceServer(s, &service.UserService{})

	log.Println("Starting gRPC user service server on port :8080")

	if err = s.Serve(listener); err != nil {
		return err
	}

	return nil
}
