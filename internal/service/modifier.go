package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrModifierGroupNotFound    = errors.New("modifier group not found")
	ErrModifierOptionNotFound   = errors.New("modifier option not found")
	ErrInvalidModifierSelection = errors.New("invalid modifier selection")
)

// ModifierService defines business operations for modifier groups and options.
type ModifierService interface {
	CreateGroup(ctx context.Context, req *models.CreateModifierGroupRequest) (*models.ModifierGroup, error)
	ListGroups(ctx context.Context) ([]*models.ModifierGroup, error)
	GetGroup(ctx context.Context, id string) (*models.ModifierGroup, error)
	UpdateGroup(ctx context.Context, id string, req *models.UpdateModifierGroupRequest) (*models.ModifierGroup, error)
	DeleteGroup(ctx context.Context, id string) error

	CreateOption(ctx context.Context, groupID string, req *models.CreateModifierOptionRequest) (*models.ModifierOption, error)
	ListOptions(ctx context.Context, groupID string) ([]models.ModifierOption, error)
	UpdateOption(ctx context.Context, id string, req *models.UpdateModifierOptionRequest) (*models.ModifierOption, error)
	DeleteOption(ctx context.Context, id string) error

	AttachToProduct(ctx context.Context, productID, groupID string) error
	DetachFromProduct(ctx context.Context, productID, groupID string) error
	GetGroupsForProduct(ctx context.Context, productID string) ([]models.ModifierGroupResponse, error)

	// ValidateAndPrice checks the selected option IDs against a product's modifier groups
	// (required groups must be satisfied within min/max bounds, no unknown options) and
	// returns the combined extra price to add on top of the product's base price.
	ValidateAndPrice(ctx context.Context, productID string, selectedOptionIDs []string) (int64, error)
}

type modifierService struct {
	repo repository.ModifierRepository
}

// NewModifierService instantiates a new ModifierService.
func NewModifierService(repo repository.ModifierRepository) ModifierService {
	return &modifierService{repo: repo}
}

func (s *modifierService) CreateGroup(ctx context.Context, req *models.CreateModifierGroupRequest) (*models.ModifierGroup, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.MinSelection < 0 || req.MaxSelection < 1 || req.MinSelection > req.MaxSelection {
		return nil, fmt.Errorf("invalid min_selection/max_selection")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}
	now := time.Now()
	g := &models.ModifierGroup{
		ID:           id.String(),
		Name:         req.Name,
		MinSelection: req.MinSelection,
		MaxSelection: req.MaxSelection,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.CreateGroup(ctx, g); err != nil {
		return nil, fmt.Errorf("failed to create modifier group: %w", err)
	}
	return g, nil
}

func (s *modifierService) ListGroups(ctx context.Context) ([]*models.ModifierGroup, error) {
	groups, err := s.repo.ListGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list modifier groups: %w", err)
	}
	return groups, nil
}

func (s *modifierService) GetGroup(ctx context.Context, id string) (*models.ModifierGroup, error) {
	g, err := s.repo.GetGroupByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrModifierGroupNotFound) {
			return nil, ErrModifierGroupNotFound
		}
		return nil, fmt.Errorf("failed to get modifier group: %w", err)
	}
	return g, nil
}

func (s *modifierService) UpdateGroup(ctx context.Context, id string, req *models.UpdateModifierGroupRequest) (*models.ModifierGroup, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.MinSelection < 0 || req.MaxSelection < 1 || req.MinSelection > req.MaxSelection {
		return nil, fmt.Errorf("invalid min_selection/max_selection")
	}

	g, err := s.repo.GetGroupByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrModifierGroupNotFound) {
			return nil, ErrModifierGroupNotFound
		}
		return nil, fmt.Errorf("failed to find modifier group: %w", err)
	}
	g.Name = req.Name
	g.MinSelection = req.MinSelection
	g.MaxSelection = req.MaxSelection
	g.UpdatedAt = time.Now()

	if err := s.repo.UpdateGroup(ctx, g); err != nil {
		if errors.Is(err, repository.ErrModifierGroupNotFound) {
			return nil, ErrModifierGroupNotFound
		}
		return nil, fmt.Errorf("failed to update modifier group: %w", err)
	}
	return g, nil
}

func (s *modifierService) DeleteGroup(ctx context.Context, id string) error {
	if err := s.repo.DeleteGroup(ctx, id); err != nil {
		if errors.Is(err, repository.ErrModifierGroupNotFound) {
			return ErrModifierGroupNotFound
		}
		return fmt.Errorf("failed to delete modifier group: %w", err)
	}
	return nil
}

func (s *modifierService) CreateOption(ctx context.Context, groupID string, req *models.CreateModifierOptionRequest) (*models.ModifierOption, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if _, err := s.repo.GetGroupByID(ctx, groupID); err != nil {
		if errors.Is(err, repository.ErrModifierGroupNotFound) {
			return nil, ErrModifierGroupNotFound
		}
		return nil, fmt.Errorf("failed to find modifier group: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}
	now := time.Now()
	o := &models.ModifierOption{
		ID:              id.String(),
		ModifierGroupID: groupID,
		Name:            req.Name,
		ExtraPrice:      req.ExtraPrice,
		IsDefault:       req.IsDefault,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.repo.CreateOption(ctx, o); err != nil {
		return nil, fmt.Errorf("failed to create modifier option: %w", err)
	}
	return o, nil
}

func (s *modifierService) ListOptions(ctx context.Context, groupID string) ([]models.ModifierOption, error) {
	options, err := s.repo.ListOptionsByGroup(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to list modifier options: %w", err)
	}
	return options, nil
}

func (s *modifierService) UpdateOption(ctx context.Context, id string, req *models.UpdateModifierOptionRequest) (*models.ModifierOption, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	o, err := s.repo.GetOptionByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrModifierOptionNotFound) {
			return nil, ErrModifierOptionNotFound
		}
		return nil, fmt.Errorf("failed to find modifier option: %w", err)
	}
	o.Name = req.Name
	o.ExtraPrice = req.ExtraPrice
	o.IsDefault = req.IsDefault
	o.UpdatedAt = time.Now()

	if err := s.repo.UpdateOption(ctx, o); err != nil {
		if errors.Is(err, repository.ErrModifierOptionNotFound) {
			return nil, ErrModifierOptionNotFound
		}
		return nil, fmt.Errorf("failed to update modifier option: %w", err)
	}
	return o, nil
}

func (s *modifierService) DeleteOption(ctx context.Context, id string) error {
	if err := s.repo.DeleteOption(ctx, id); err != nil {
		if errors.Is(err, repository.ErrModifierOptionNotFound) {
			return ErrModifierOptionNotFound
		}
		return fmt.Errorf("failed to delete modifier option: %w", err)
	}
	return nil
}

func (s *modifierService) AttachToProduct(ctx context.Context, productID, groupID string) error {
	if err := s.repo.AttachToProduct(ctx, productID, groupID); err != nil {
		return fmt.Errorf("failed to attach modifier group to product: %w", err)
	}
	return nil
}

func (s *modifierService) DetachFromProduct(ctx context.Context, productID, groupID string) error {
	if err := s.repo.DetachFromProduct(ctx, productID, groupID); err != nil {
		if errors.Is(err, repository.ErrModifierGroupNotFound) {
			return ErrModifierGroupNotFound
		}
		return fmt.Errorf("failed to detach modifier group from product: %w", err)
	}
	return nil
}

func (s *modifierService) GetGroupsForProduct(ctx context.Context, productID string) ([]models.ModifierGroupResponse, error) {
	groups, err := s.repo.ListGroupsForProduct(ctx, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get modifier groups for product: %w", err)
	}

	resp := make([]models.ModifierGroupResponse, len(groups))
	for i, g := range groups {
		options := make([]models.ModifierOptionResponse, len(g.Options))
		for j, o := range g.Options {
			options[j] = models.ModifierOptionResponse{
				ID:         o.ID,
				Name:       o.Name,
				ExtraPrice: o.ExtraPrice,
				IsDefault:  o.IsDefault,
			}
		}
		resp[i] = models.ModifierGroupResponse{
			ID:           g.ID,
			Name:         g.Name,
			MinSelection: g.MinSelection,
			MaxSelection: g.MaxSelection,
			IsRequired:   g.MinSelection >= 1,
			Options:      options,
		}
	}
	return resp, nil
}

func (s *modifierService) ValidateAndPrice(ctx context.Context, productID string, selectedOptionIDs []string) (int64, error) {
	groups, err := s.repo.ListGroupsForProduct(ctx, productID)
	if err != nil {
		return 0, fmt.Errorf("failed to load modifier groups for product: %w", err)
	}
	return ValidateAndPriceSelections(groups, selectedOptionIDs)
}

// ValidateAndPriceSelections checks selectedOptionIDs against a product's modifier groups —
// every group's selection count must fall within [MinSelection, MaxSelection], and every
// selected option must belong to one of the product's groups — and returns the sum of the
// selected options' extra prices. Exported so OrderService can reuse it without depending on
// ModifierService (services only depend on repository interfaces in this codebase).
func ValidateAndPriceSelections(groups []models.ModifierGroupWithOptions, selectedOptionIDs []string) (int64, error) {
	selected := make(map[string]bool, len(selectedOptionIDs))
	for _, id := range selectedOptionIDs {
		selected[id] = true
	}

	optionToGroup := make(map[string]string)
	var extraPrice int64
	for _, g := range groups {
		count := 0
		for _, o := range g.Options {
			optionToGroup[o.ID] = g.ID
			if selected[o.ID] {
				count++
				extraPrice += o.ExtraPrice
			}
		}
		if count < g.MinSelection || count > g.MaxSelection {
			return 0, fmt.Errorf("%w: group %q requires between %d and %d selection(s), got %d", ErrInvalidModifierSelection, g.Name, g.MinSelection, g.MaxSelection, count)
		}
	}

	for _, id := range selectedOptionIDs {
		if _, ok := optionToGroup[id]; !ok {
			return 0, fmt.Errorf("%w: option %q does not belong to this product", ErrInvalidModifierSelection, id)
		}
	}

	return extraPrice, nil
}
