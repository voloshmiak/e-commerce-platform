package service

import (
	"context"
	"payment-service/data"
	pb "payment-service/protobuf"
)

type PaymentService struct {
	pb.UnimplementedPaymentServiceServer
}

func (s *PaymentService) ProcessPayment(_ context.Context, r *pb.ProcessPaymentRequest) (*pb.ProcessPaymentResponse, error) {
	transaction := data.AddTransaction(r.GetOrderId(), r.GetAmount(), r.GetCurrency().String(), r.GetPaymentMethod().String())

	return &pb.ProcessPaymentResponse{
		TransactionId: transaction.ID,
	}, nil
}

func (s *PaymentService) GetPaymentStatus(_ context.Context, r *pb.GetPaymentStatusRequest) (*pb.GetPaymentStatusResponse, error) {
	transaction := data.GetTransactionByID(r.GetTransactionId())
	var status pb.Status
	switch transaction.Status {
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
