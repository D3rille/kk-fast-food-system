package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrCategoryNotFound = errors.New("category not found")

// CategoryRepository defines data access rules for Categories.
type CategoryRepository interface {
	Create(ctx context.Context, cat *models.Category) error
	GetByID(ctx context.Context, id string) (*models.Category, error)
	ListActiveByStore(ctx context.Context, storeID string) ([]*models.Category, error)
	ListAllByStore(ctx context.Context, storeID string) ([]*models.Category, error)
	Update(ctx context.Context, cat *models.Category) error
	Delete(ctx context.Context, id string) error
}

type postgresCategoryRepository struct {
	db *pgxpool.Pool
}

// NewCategoryRepository creates a concrete implementation of CategoryRepository.
func NewCategoryRepository(db *pgxpool.Pool) CategoryRepository {
	return &postgresCategoryRepository{db: db}
}

func (r *postgresCategoryRepository) Create(ctx context.Context, cat *models.Category) error {
	query := `
		INSERT INTO categories (id, store_id, name, sort_order, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		cat.ID,
		cat.StoreID,
		cat.Name,
		cat.SortOrder,
		cat.IsActive,
		cat.CreatedAt,
		cat.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}
	return nil
}

func (r *postgresCategoryRepository) GetByID(ctx context.Context, id string) (*models.Category, error) {
	query := `
		SELECT id, store_id, name, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE id = $1
	`
	var cat models.Category
	err := r.db.QueryRow(ctx, query, id).Scan(
		&cat.ID,
		&cat.StoreID,
		&cat.Name,
		&cat.SortOrder,
		&cat.IsActive,
		&cat.CreatedAt,
		&cat.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return &cat, nil
}

func (r *postgresCategoryRepository) ListActiveByStore(ctx context.Context, storeID string) ([]*models.Category, error) {
	query := `
		SELECT id, store_id, name, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE store_id = $1 AND is_active = TRUE
		ORDER BY sort_order ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query active categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var cat models.Category
		err := rows.Scan(
			&cat.ID,
			&cat.StoreID,
			&cat.Name,
			&cat.SortOrder,
			&cat.IsActive,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		categories = append(categories, &cat)
	}
	return categories, nil
}

func (r *postgresCategoryRepository) ListAllByStore(ctx context.Context, storeID string) ([]*models.Category, error) {
	query := `
		SELECT id, store_id, name, sort_order, is_active, created_at, updated_at
		FROM categories
		WHERE store_id = $1
		ORDER BY sort_order ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query all categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		var cat models.Category
		err := rows.Scan(
			&cat.ID,
			&cat.StoreID,
			&cat.Name,
			&cat.SortOrder,
			&cat.IsActive,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		categories = append(categories, &cat)
	}
	return categories, nil
}

func (r *postgresCategoryRepository) Update(ctx context.Context, cat *models.Category) error {
	query := `
		UPDATE categories
		SET name = $2, sort_order = $3, is_active = $4, updated_at = $5
		WHERE id = $1
	`
	commandTag, err := r.db.Exec(ctx, query,
		cat.ID,
		cat.Name,
		cat.SortOrder,
		cat.IsActive,
		cat.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no category was updated")
	}
	return nil
}

func (r *postgresCategoryRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM categories
		WHERE id = $1
	`
	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no category was deleted")
	}
	return nil
}
