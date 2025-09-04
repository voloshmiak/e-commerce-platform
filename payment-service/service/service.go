package service

import "payment-service/data"

type PaymentService struct{}

func (s *PaymentService) ProcessPayment(orderID int64, amount float64, currency string, paymentMethod string) (int64, error) {
	transaction := data.AddTransaction(orderID, amount, currency, paymentMethod)
	return transaction.ID, nil
}

func (s *PaymentService) GetPaymentStatus(transactionID int64) data.Status {
	transaction := data.GetTransactionByID(transactionID)
	return transaction.Status
}
