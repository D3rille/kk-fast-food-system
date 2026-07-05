package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProductNotFound = errors.New("product not found")

// ProductRepository defines all database access methods for products.
type ProductRepository interface {
	Create(ctx context.Context, p *models.Product) error
	GetByID(ctx context.Context, id string) (*models.Product, error)
	ListAll(ctx context.Context) ([]*models.Product, error)
	ListByCategory(ctx context.Context, categoryID string) ([]*models.Product, error)
	ListAvailableByCategory(ctx context.Context, categoryID string) ([]*models.Product, error)
	Update(ctx context.Context, p *models.Product) error
	Delete(ctx context.Context, id string) error
	ToggleAvailability(ctx context.Context, id string) (*models.Product, error)
}

type postgresProductRepository struct {
	db *pgxpool.Pool
}

// NewProductRepository creates a concrete repository implementation.
func NewProductRepository(db *pgxpool.Pool) ProductRepository {
	return &postgresProductRepository{db: db}
}

func (r *postgresProductRepository) Create(ctx context.Context, p *models.Product) error {
	query := `
		INSERT INTO products (id, category_id, name, description, base_price, image_url, is_available, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, query,
		p.ID, p.CategoryID, p.Name, p.Description, p.BasePrice, p.ImageURL, p.IsAvailable, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

func (r *postgresProductRepository) GetByID(ctx context.Context, id string) (*models.Product, error) {
	query := `
		SELECT id, category_id, name, description, base_price, image_url, is_available, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	var p models.Product
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.BasePrice, &p.ImageURL, &p.IsAvailable, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &p, nil
}

func (r *postgresProductRepository) ListAll(ctx context.Context) ([]*models.Product, error) {
	query := `
		SELECT id, category_id, name, description, base_price, image_url, is_available, created_at, updated_at
		FROM products
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	return scanProducts(rows)
}

func (r *postgresProductRepository) ListByCategory(ctx context.Context, categoryID string) ([]*models.Product, error) {
	query := `
		SELECT id, category_id, name, description, base_price, image_url, is_available, created_at, updated_at
		FROM products
		WHERE category_id = $1
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query products by category: %w", err)
	}
	return scanProducts(rows)
}

func (r *postgresProductRepository) ListAvailableByCategory(ctx context.Context, categoryID string) ([]*models.Product, error) {
	query := `
		SELECT id, category_id, name, description, base_price, image_url, is_available, created_at, updated_at
		FROM products
		WHERE category_id = $1 AND is_available = TRUE
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query available products: %w", err)
	}
	return scanProducts(rows)
}

func (r *postgresProductRepository) Update(ctx context.Context, p *models.Product) error {
	query := `
		UPDATE products
		SET category_id = $2, name = $3, description = $4, base_price = $5, image_url = $6, is_available = $7, updated_at = $8
		WHERE id = $1
	`
	commandTag, err := r.db.Exec(ctx, query,
		p.ID, p.CategoryID, p.Name, p.Description, p.BasePrice, p.ImageURL, p.IsAvailable, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *postgresProductRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *postgresProductRepository) ToggleAvailability(ctx context.Context, id string) (*models.Product, error) {
	query := `
		UPDATE products
		SET is_available = NOT is_available, updated_at = NOW()
		WHERE id = $1
		RETURNING id, category_id, name, description, base_price, image_url, is_available, created_at, updated_at
	`
	var p models.Product
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.BasePrice, &p.ImageURL, &p.IsAvailable, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to toggle product availability: %w", err)
	}
	return &p, nil
}

func scanProducts(rows pgx.Rows) ([]*models.Product, error) {
	defer rows.Close()
	var items []*models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID, &p.CategoryID, &p.Name, &p.Description, &p.BasePrice, &p.ImageURL, &p.IsAvailable, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product row: %w", err)
		}
		items = append(items, &p)
	}
	return items, nil
}
