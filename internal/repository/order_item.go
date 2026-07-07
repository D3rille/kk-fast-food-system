package repository

import (
	"context"
	"fmt"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderItemRepository defines database access for order line items.
type OrderItemRepository interface {
	CreateBatch(ctx context.Context, items []*models.OrderItem) error
	GetByOrderIDWithProducts(ctx context.Context, orderID string) ([]models.OrderItemResponse, error)
}

type postgresOrderItemRepository struct {
	db *pgxpool.Pool
}

// NewOrderItemRepository creates a concrete repository implementation.
func NewOrderItemRepository(db *pgxpool.Pool) OrderItemRepository {
	return &postgresOrderItemRepository{db: db}
}

func (r *postgresOrderItemRepository) CreateBatch(ctx context.Context, items []*models.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	query := `
		INSERT INTO order_items (id, order_id, product_id, quantity, unit_price, calculated_subtotal, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	modifierQuery := `
		INSERT INTO order_item_modifiers (id, order_item_id, modifier_option_id, created_at)
		VALUES ($1, $2, $3, $4)
	`
	for _, item := range items {
		if _, err := r.db.Exec(ctx, query,
			item.ID,
			item.OrderID,
			item.ProductID,
			item.Quantity,
			item.UnitPrice,
			item.CalculatedSubtotal,
			item.CreatedAt,
		); err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}

		for _, optionID := range item.ModifierOptionIDs {
			modID, err := uuid.NewV7()
			if err != nil {
				return fmt.Errorf("failed to generate order item modifier ID: %w", err)
			}
			if _, err := r.db.Exec(ctx, modifierQuery, modID.String(), item.ID, optionID, item.CreatedAt); err != nil {
				return fmt.Errorf("failed to insert order item modifier: %w", err)
			}
		}
	}
	return nil
}

func (r *postgresOrderItemRepository) GetByOrderIDWithProducts(ctx context.Context, orderID string) ([]models.OrderItemResponse, error) {
	query := `
		SELECT oi.id, oi.product_id, p.name, oi.quantity, oi.unit_price, oi.calculated_subtotal, oi.created_at
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = $1
		ORDER BY oi.created_at ASC
	`
	rows, err := r.db.Query(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []models.OrderItemResponse
	for rows.Next() {
		var item models.OrderItemResponse
		if scanErr := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
			&item.Subtotal,
			&item.CreatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan order item row: %w", scanErr)
		}
		item.Modifiers = []models.OrderItemModifierResponse{}
		items = append(items, item)
	}
	if items == nil {
		items = []models.OrderItemResponse{}
		return items, nil
	}

	itemIDs := make([]string, len(items))
	indexByItemID := make(map[string]int, len(items))
	for i, item := range items {
		itemIDs[i] = item.ID
		indexByItemID[item.ID] = i
	}

	modifierQuery := `
		SELECT oim.order_item_id, oim.modifier_option_id, mo.name, mo.extra_price, mg.name
		FROM order_item_modifiers oim
		JOIN modifier_options mo ON mo.id = oim.modifier_option_id
		JOIN modifier_groups mg ON mg.id = mo.modifier_group_id
		WHERE oim.order_item_id = ANY($1)
	`
	modRows, err := r.db.Query(ctx, modifierQuery, itemIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query order item modifiers: %w", err)
	}
	defer modRows.Close()

	for modRows.Next() {
		var orderItemID string
		var m models.OrderItemModifierResponse
		if err := modRows.Scan(&orderItemID, &m.ModifierOptionID, &m.Name, &m.ExtraPrice, &m.GroupName); err != nil {
			return nil, fmt.Errorf("failed to scan order item modifier row: %w", err)
		}
		idx := indexByItemID[orderItemID]
		items[idx].Modifiers = append(items[idx].Modifiers, m)
	}

	return items, nil
}
