package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderRepository defines all database access methods for order.
type OrderRepository interface {
	Create(ctx context.Context, item *models.Order) error
	GetByID(ctx context.Context, id string) (*models.Order, error)
	// List returns orders sorted oldest-first. Pass a non-empty status to filter by that status.
	List(ctx context.Context, status string) ([]*models.Order, error)
	Update(ctx context.Context, item *models.Order) error
	UpdateStatus(ctx context.Context, id string, status models.OrderStatus, paymentStatus models.PaymentStatus) error
	Delete(ctx context.Context, id string) error
}

type postgresOrderRepository struct {
	db *pgxpool.Pool
}

// NewOrderRepository creates a concrete repository implementation.
func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return &postgresOrderRepository{db: db}
}

func (r *postgresOrderRepository) Create(ctx context.Context, item *models.Order) error {
	query := `
		INSERT INTO orders (id, store_id, order_number, source, status, payment_status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		item.ID,
		item.StoreID,
		item.OrderNumber,
		item.Source,
		item.Status,
		item.PaymentStatus,
		item.TotalAmount,
		item.CreatedAt,
		item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}
	return nil
}

func (r *postgresOrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	query := `
		SELECT id, store_id, order_number, source, status, payment_status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	var item models.Order
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.StoreID,
		&item.OrderNumber,
		&item.Source,
		&item.Status,
		&item.PaymentStatus,
		&item.TotalAmount,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order by id: %w", err)
	}
	return &item, nil
}

func (r *postgresOrderRepository) List(ctx context.Context, status string) ([]*models.Order, error) {
	var (
		query string
		args  []any
	)
	if status != "" {
		query = `
			SELECT id, store_id, order_number, source, status, payment_status, total_amount, created_at, updated_at
			FROM orders
			WHERE status = $1
			ORDER BY created_at ASC
		`
		args = []any{status}
	} else {
		query = `
			SELECT id, store_id, order_number, source, status, payment_status, total_amount, created_at, updated_at
			FROM orders
			ORDER BY created_at ASC
		`
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query order: %w", err)
	}
	defer rows.Close()

	var items []*models.Order
	for rows.Next() {
		var item models.Order
		if err := rows.Scan(
			&item.ID,
			&item.StoreID,
			&item.OrderNumber,
			&item.Source,
			&item.Status,
			&item.PaymentStatus,
			&item.TotalAmount,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan order row: %w", err)
		}
		items = append(items, &item)
	}
	return items, nil
}

func (r *postgresOrderRepository) Update(ctx context.Context, item *models.Order) error {
	query := `
		UPDATE orders
		SET store_id = $2, source = $3, status = $4, payment_status = $5, total_amount = $6, updated_at = $7
		WHERE id = $1
	`
	commandTag, err := r.db.Exec(ctx, query,
		item.ID, item.StoreID, item.Source, item.Status, item.PaymentStatus, item.TotalAmount, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no order rows updated")
	}
	return nil
}

func (r *postgresOrderRepository) UpdateStatus(ctx context.Context, id string, status models.OrderStatus, paymentStatus models.PaymentStatus) error {
	query := `
		UPDATE orders
		SET status = $2, payment_status = $3, updated_at = $4
		WHERE id = $1
	`
	commandTag, err := r.db.Exec(ctx, query, id, status, paymentStatus, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("order not found: %s", id)
	}
	return nil
}

func (r *postgresOrderRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM orders
		WHERE id = $1
	`
	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no order rows deleted")
	}
	return nil
}
