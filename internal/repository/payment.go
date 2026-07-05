package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PaymentRepository defines all database access methods for payment.
type PaymentRepository interface {
	Create(ctx context.Context, p *models.Payment) error
	UpdateStatus(ctx context.Context, id string, status models.PaymentStatus, transactionRef string) error
}

type postgresPaymentRepository struct {
	db *pgxpool.Pool
}

// NewPaymentRepository creates a concrete repository implementation.
func NewPaymentRepository(db *pgxpool.Pool) PaymentRepository {
	return &postgresPaymentRepository{db: db}
}

func (r *postgresPaymentRepository) Create(ctx context.Context, p *models.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, provider, amount, status, transaction_ref, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.OrderID, p.Provider, p.Amount, p.Status, p.TransactionRef, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}
	return nil
}

func (r *postgresPaymentRepository) UpdateStatus(ctx context.Context, id string, status models.PaymentStatus, transactionRef string) error {
	query := `
		UPDATE payments
		SET status = $2, transaction_ref = $3, updated_at = $4
		WHERE id = $1
	`
	commandTag, err := r.db.Exec(ctx, query, id, status, transactionRef, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("payment not found: %s", id)
	}
	return nil
}
