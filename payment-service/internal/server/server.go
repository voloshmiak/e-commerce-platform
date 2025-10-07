package server

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"payment-service/internal/service"
	pb "payment-service/protobuf"
)

type Server struct {
	pb.UnimplementedPaymentServiceServer
	Service *service.Service
}

func (s *Server) ProcessPayment(ctx context.Context, _ *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	pi, err := s.Service.ProcessPayment(ctx)
	if err != nil {
		log.Println(err)
		switch {
		case errors.Is(err, service.ErrEmptyCart):
			return nil, status.Error(codes.FailedPrecondition, "cart is empty")
		case errors.Is(err, service.ErrProcessingPayment):
			return nil, status.Error(codes.Internal, "error processing payment")
		}
		return nil, status.Error(codes.Internal, "unknown error processing payment")
	}

	return &pb.ProcessPaymentResponse{
		ClientSecret: pi.ClientSecret,
		Amount:       float32(pi.Amount),
		Currency:     pi.Currency,
	}, nil
}
