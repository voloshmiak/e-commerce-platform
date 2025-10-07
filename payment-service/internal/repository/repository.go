package repository

import (
	"context"
	"database/sql"
	"payment-service/internal/model"
)

type Repository struct {
	conn *sql.DB
}

func New(conn *sql.DB) *Repository {
	return &Repository{conn: conn}
}

func (r *Repository) CreateTransaction(ctx context.Context, transaction *model.Transaction) (int64, error) {
	query := `INSERT INTO transactions (order_id, amount, currency, status, gateway_transaction_id, payment_method) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := r.conn.QueryRowContext(ctx, query,
		transaction.OrderID,
		transaction.Amount,
		transaction.Currency,
		transaction.Status,
		transaction.GatewayTransactionID,
		transaction.PaymentMethod,
	).Scan(&transaction.ID)
	if err != nil {
		return 0, err
	}

	return transaction.ID, nil
}

func (r *Repository) GetTransactionByPaymentIntentID(ctx context.Context, paymentIntentID string) (*model.Transaction, error) {
	query := `SELECT id, order_id, amount, currency, status, gateway_transaction_id, payment_method, created_at 
			  FROM transactions WHERE gateway_transaction_id = $1`

	row := r.conn.QueryRowContext(ctx, query, paymentIntentID)
	var transaction model.Transaction
	err := row.Scan(
		&transaction.ID,
		&transaction.OrderID,
		&transaction.Amount,
		&transaction.Currency,
		&transaction.Status,
		&transaction.GatewayTransactionID,
		&transaction.PaymentMethod,
		&transaction.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

func (r *Repository) UpdateTransaction(ctx context.Context, transaction *model.Transaction) error {
	query := `UPDATE transactions SET order_id = $1, amount = $2, currency = $3, status = $4, 
			  gateway_transaction_id = $5, payment_method = $6 WHERE id = $7`

	_, err := r.conn.ExecContext(ctx, query,
		transaction.OrderID,
		transaction.Amount,
		transaction.Currency,
		transaction.Status,
		transaction.GatewayTransactionID,
		transaction.PaymentMethod,
		transaction.ID,
	)
	return err
}

func (r *Repository) UpdateTransactionStatus(ctx context.Context, paymentIntentID string, status model.Status) error {
	query := `UPDATE transactions SET status = $1 WHERE gateway_transaction_id = $2`
	_, err := r.conn.ExecContext(ctx, query, status, paymentIntentID)
	return err
}

func (r *Repository) UpdateTransactionOrderID(ctx context.Context, paymentIntentID string, orderID int64) error {
	query := `UPDATE transactions SET order_id = $1 WHERE gateway_transaction_id = $2`
	_, err := r.conn.ExecContext(ctx, query, orderID, paymentIntentID)
	return err
}
