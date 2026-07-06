package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrModifierGroupNotFound  = errors.New("modifier group not found")
	ErrModifierOptionNotFound = errors.New("modifier option not found")
)

// ModifierRepository defines all database access methods for modifier groups and options.
type ModifierRepository interface {
	CreateGroup(ctx context.Context, g *models.ModifierGroup) error
	GetGroupByID(ctx context.Context, id string) (*models.ModifierGroup, error)
	ListGroups(ctx context.Context) ([]*models.ModifierGroup, error)
	UpdateGroup(ctx context.Context, g *models.ModifierGroup) error
	DeleteGroup(ctx context.Context, id string) error

	CreateOption(ctx context.Context, o *models.ModifierOption) error
	GetOptionByID(ctx context.Context, id string) (*models.ModifierOption, error)
	ListOptionsByGroup(ctx context.Context, groupID string) ([]models.ModifierOption, error)
	ListOptionsByIDs(ctx context.Context, ids []string) ([]models.ModifierOption, error)
	UpdateOption(ctx context.Context, o *models.ModifierOption) error
	DeleteOption(ctx context.Context, id string) error

	AttachToProduct(ctx context.Context, productID, groupID string) error
	DetachFromProduct(ctx context.Context, productID, groupID string) error
	ListGroupsForProduct(ctx context.Context, productID string) ([]models.ModifierGroupWithOptions, error)
}

type postgresModifierRepository struct {
	db *pgxpool.Pool
}

// NewModifierRepository creates a concrete repository implementation.
func NewModifierRepository(db *pgxpool.Pool) ModifierRepository {
	return &postgresModifierRepository{db: db}
}

func (r *postgresModifierRepository) CreateGroup(ctx context.Context, g *models.ModifierGroup) error {
	query := `
		INSERT INTO modifier_groups (id, name, min_selection, max_selection, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(ctx, query, g.ID, g.Name, g.MinSelection, g.MaxSelection, g.CreatedAt, g.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create modifier group: %w", err)
	}
	return nil
}

func (r *postgresModifierRepository) GetGroupByID(ctx context.Context, id string) (*models.ModifierGroup, error) {
	query := `
		SELECT id, name, min_selection, max_selection, created_at, updated_at
		FROM modifier_groups
		WHERE id = $1
	`
	var g models.ModifierGroup
	err := r.db.QueryRow(ctx, query, id).Scan(&g.ID, &g.Name, &g.MinSelection, &g.MaxSelection, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrModifierGroupNotFound
		}
		return nil, fmt.Errorf("failed to get modifier group: %w", err)
	}
	return &g, nil
}

func (r *postgresModifierRepository) ListGroups(ctx context.Context) ([]*models.ModifierGroup, error) {
	query := `
		SELECT id, name, min_selection, max_selection, created_at, updated_at
		FROM modifier_groups
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query modifier groups: %w", err)
	}
	defer rows.Close()

	var groups []*models.ModifierGroup
	for rows.Next() {
		var g models.ModifierGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.MinSelection, &g.MaxSelection, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan modifier group row: %w", err)
		}
		groups = append(groups, &g)
	}
	return groups, nil
}

func (r *postgresModifierRepository) UpdateGroup(ctx context.Context, g *models.ModifierGroup) error {
	query := `
		UPDATE modifier_groups
		SET name = $2, min_selection = $3, max_selection = $4, updated_at = $5
		WHERE id = $1
	`
	commandTag, err := r.db.Exec(ctx, query, g.ID, g.Name, g.MinSelection, g.MaxSelection, g.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update modifier group: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrModifierGroupNotFound
	}
	return nil
}

func (r *postgresModifierRepository) DeleteGroup(ctx context.Context, id string) error {
	commandTag, err := r.db.Exec(ctx, `DELETE FROM modifier_groups WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete modifier group: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrModifierGroupNotFound
	}
	return nil
}

func (r *postgresModifierRepository) CreateOption(ctx context.Context, o *models.ModifierOption) error {
	query := `
		INSERT INTO modifier_options (id, modifier_group_id, name, extra_price, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query, o.ID, o.ModifierGroupID, o.Name, o.ExtraPrice, o.IsDefault, o.CreatedAt, o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create modifier option: %w", err)
	}
	return nil
}

func (r *postgresModifierRepository) GetOptionByID(ctx context.Context, id string) (*models.ModifierOption, error) {
	query := `
		SELECT id, modifier_group_id, name, extra_price, is_default, created_at, updated_at
		FROM modifier_options
		WHERE id = $1
	`
	var o models.ModifierOption
	err := r.db.QueryRow(ctx, query, id).Scan(&o.ID, &o.ModifierGroupID, &o.Name, &o.ExtraPrice, &o.IsDefault, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrModifierOptionNotFound
		}
		return nil, fmt.Errorf("failed to get modifier option: %w", err)
	}
	return &o, nil
}

func (r *postgresModifierRepository) ListOptionsByGroup(ctx context.Context, groupID string) ([]models.ModifierOption, error) {
	query := `
		SELECT id, modifier_group_id, name, extra_price, is_default, created_at, updated_at
		FROM modifier_options
		WHERE modifier_group_id = $1
		ORDER BY extra_price ASC, name ASC
	`
	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to query modifier options: %w", err)
	}
	return scanModifierOptions(rows)
}

func (r *postgresModifierRepository) ListOptionsByIDs(ctx context.Context, ids []string) ([]models.ModifierOption, error) {
	if len(ids) == 0 {
		return []models.ModifierOption{}, nil
	}
	query := `
		SELECT id, modifier_group_id, name, extra_price, is_default, created_at, updated_at
		FROM modifier_options
		WHERE id = ANY($1)
	`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to query modifier options by id: %w", err)
	}
	return scanModifierOptions(rows)
}

func (r *postgresModifierRepository) UpdateOption(ctx context.Context, o *models.ModifierOption) error {
	query := `
		UPDATE modifier_options
		SET name = $2, extra_price = $3, is_default = $4, updated_at = $5
		WHERE id = $1
	`
	commandTag, err := r.db.Exec(ctx, query, o.ID, o.Name, o.ExtraPrice, o.IsDefault, o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update modifier option: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrModifierOptionNotFound
	}
	return nil
}

func (r *postgresModifierRepository) DeleteOption(ctx context.Context, id string) error {
	commandTag, err := r.db.Exec(ctx, `DELETE FROM modifier_options WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete modifier option: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrModifierOptionNotFound
	}
	return nil
}

func (r *postgresModifierRepository) AttachToProduct(ctx context.Context, productID, groupID string) error {
	query := `
		INSERT INTO product_modifier_groups (product_id, modifier_group_id)
		VALUES ($1, $2)
		ON CONFLICT (product_id, modifier_group_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, productID, groupID)
	if err != nil {
		return fmt.Errorf("failed to attach modifier group to product: %w", err)
	}
	return nil
}

func (r *postgresModifierRepository) DetachFromProduct(ctx context.Context, productID, groupID string) error {
	commandTag, err := r.db.Exec(ctx, `
		DELETE FROM product_modifier_groups WHERE product_id = $1 AND modifier_group_id = $2
	`, productID, groupID)
	if err != nil {
		return fmt.Errorf("failed to detach modifier group from product: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return ErrModifierGroupNotFound
	}
	return nil
}

// ListGroupsForProduct returns every modifier group attached to a product, each with its options,
// ordered so required groups (min_selection >= 1) appear first.
func (r *postgresModifierRepository) ListGroupsForProduct(ctx context.Context, productID string) ([]models.ModifierGroupWithOptions, error) {
	query := `
		SELECT g.id, g.name, g.min_selection, g.max_selection, g.created_at, g.updated_at,
		       o.id, o.modifier_group_id, o.name, o.extra_price, o.is_default, o.created_at, o.updated_at
		FROM modifier_groups g
		JOIN product_modifier_groups pmg ON pmg.modifier_group_id = g.id
		LEFT JOIN modifier_options o ON o.modifier_group_id = g.id
		WHERE pmg.product_id = $1
		ORDER BY g.min_selection DESC, g.name ASC, o.extra_price ASC, o.name ASC
	`
	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to query modifier groups for product: %w", err)
	}
	defer rows.Close()

	var groups []models.ModifierGroupWithOptions
	index := make(map[string]int)
	for rows.Next() {
		var g models.ModifierGroup
		var opt models.ModifierOption
		var optID, optGroupID, optName *string
		var optExtraPrice *int64
		var optIsDefault *bool
		var optCreatedAt, optUpdatedAt *time.Time

		if err := rows.Scan(
			&g.ID, &g.Name, &g.MinSelection, &g.MaxSelection, &g.CreatedAt, &g.UpdatedAt,
			&optID, &optGroupID, &optName, &optExtraPrice, &optIsDefault, &optCreatedAt, &optUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan modifier group row: %w", err)
		}

		idx, ok := index[g.ID]
		if !ok {
			groups = append(groups, models.ModifierGroupWithOptions{ModifierGroup: g, Options: []models.ModifierOption{}})
			idx = len(groups) - 1
			index[g.ID] = idx
		}

		if optID != nil {
			opt = models.ModifierOption{
				ID:              *optID,
				ModifierGroupID: *optGroupID,
				Name:            *optName,
				ExtraPrice:      *optExtraPrice,
				IsDefault:       *optIsDefault,
				CreatedAt:       *optCreatedAt,
				UpdatedAt:       *optUpdatedAt,
			}
			groups[idx].Options = append(groups[idx].Options, opt)
		}
	}
	return groups, nil
}

func scanModifierOptions(rows pgx.Rows) ([]models.ModifierOption, error) {
	defer rows.Close()
	var options []models.ModifierOption
	for rows.Next() {
		var o models.ModifierOption
		if err := rows.Scan(&o.ID, &o.ModifierGroupID, &o.Name, &o.ExtraPrice, &o.IsDefault, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan modifier option row: %w", err)
		}
		options = append(options, o)
	}
	return options, nil
}
