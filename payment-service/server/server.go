package server

import (
	"context"
	"payment-service/data"
	pb "payment-service/protobuf"
	"payment-service/service"
)

type PaymentServer struct {
	pb.UnimplementedPaymentServiceServer
	service *service.PaymentService
}

func (s *PaymentServer) ProcessPayment(_ context.Context, r *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	transactionID, _ := s.service.ProcessPayment(r.GetOrderId(), r.GetAmount(), r.GetCurrency().String(), r.GetPaymentMethod().String())

	return &pb.ProcessPaymentResponse{
		TransactionId: transactionID,
	}, nil
}

func (s *PaymentServer) GetPaymentStatus(_ context.Context, r *pb.GetPaymentStatusRequest) (*pb.GetPaymentStatusResponse, error) {
	transactionStatus := s.service.GetPaymentStatus(r.GetTransactionId())
	var status pb.Status
	switch transactionStatus {
	case data.Pending:
		status = pb.Status_PENDING
	case data.Completed:
		status = pb.Status_COMPLETED
	case data.Failed:
		status = pb.Status_FAILED
	case data.Refunded:
		status = pb.Status_REFUNDED
	}

	return &pb.GetPaymentStatusResponse{
		Status: status,
	}, nil
}
