package repository

import (
	"context"
	"fmt"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StoreRepository defines database access methods for Stores.
type StoreRepository interface {
	GetByID(ctx context.Context, id string) (*models.Store, error)
	Create(ctx context.Context, store *models.Store) error
}

type postgresStoreRepository struct {
	db *pgxpool.Pool
}

// NewStoreRepository creates a concrete PostgreSQL implementation of StoreRepository.
func NewStoreRepository(db *pgxpool.Pool) StoreRepository {
	return &postgresStoreRepository{db: db}
}

func (r *postgresStoreRepository) GetByID(ctx context.Context, id string) (*models.Store, error) {
	query := `
		SELECT id, name, address, timezone, is_active, created_at, updated_at
		FROM stores
		WHERE id = $1
	`
	var store models.Store
	err := r.db.QueryRow(ctx, query, id).Scan(
		&store.ID,
		&store.Name,
		&store.Address,
		&store.Timezone,
		&store.IsActive,
		&store.CreatedAt,
		&store.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("store not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get store: %w", err)
	}

	return &store, nil
}

func (r *postgresStoreRepository) Create(ctx context.Context, store *models.Store) error {
	query := `
		INSERT INTO stores (id, name, address, timezone, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		store.ID,
		store.Name,
		store.Address,
		store.Timezone,
		store.IsActive,
		store.CreatedAt,
		store.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create store: %w", err)
	}

	return nil
}
