package data

import "time"

type Status string

const (
	Pending   Status = "pending"
	Completed Status = "completed"
	Failed    Status = "failed"
	Refunded  Status = "refunded"
)

type PaymentMethod string

const (
	CreditCard PaymentMethod = "credit_card"
	PayPal     PaymentMethod = "paypal"
)

type Transaction struct {
	ID                   int64         `json:"id"`
	OrderID              int64         `json:"order_id"`
	Amount               float64       `json:"amount"`
	Currency             string        `json:"currency"`
	Status               Status        `json:"Status"`
	GatewayTransactionID string        `json:"gateway_transaction_id"`
	PaymentMethod        PaymentMethod `json:"payment_method"`
	CreatedAt            time.Time     `json:"created_at"`
}

var Transactions = []*Transaction{
	{
		ID:                   1,
		OrderID:              101,
		Amount:               1000,
		Currency:             "USD",
		Status:               Completed,
		GatewayTransactionID: "txn_123456789",
		PaymentMethod:        CreditCard,
		CreatedAt:            time.Now(),
	},
	{
		ID:                   2,
		OrderID:              102,
		Amount:               2000,
		Currency:             "USD",
		Status:               Pending,
		GatewayTransactionID: "txn_987654321",
		PaymentMethod:        PayPal,
		CreatedAt:            time.Now(),
	},
}

func GetTransactionByID(id int64) *Transaction {
	for _, transaction := range Transactions {
		if transaction.ID == id {
			return transaction
		}
	}
	return nil
}

func AddTransaction(orderID int64, amount float64, currency, paymentMethod string) int64 {
	transaction := &Transaction{
		ID:                   int64(len(Transactions) + 1),
		OrderID:              orderID,
		Amount:               amount,
		Currency:             currency,
		Status:               Pending,
		GatewayTransactionID: "txn_" + time.Now().Format("20060102150405"),
		PaymentMethod:        PaymentMethod(paymentMethod),
		CreatedAt:            time.Now(),
	}

	Transactions = append(Transactions, transaction)

	return transaction.ID
}
