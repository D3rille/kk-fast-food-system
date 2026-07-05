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

var ErrCategoryNotFound = errors.New("category not found")

// CategoryService defines business operations for categories.
type CategoryService interface {
	Create(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error)
	GetByID(ctx context.Context, id string) (*models.Category, error)
	ListActive(ctx context.Context, storeID string) ([]*models.Category, error)
	ListAll(ctx context.Context, storeID string) ([]*models.Category, error)
	Update(ctx context.Context, id string, req *models.UpdateCategoryRequest) (*models.Category, error)
	Delete(ctx context.Context, id string) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService instantiates a new CategoryService.
func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) Create(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error) {
	if req.StoreID == "" {
		return nil, fmt.Errorf("store_id is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	now := time.Now()
	cat := &models.Category{
		ID:        id.String(),
		StoreID:   req.StoreID,
		Name:      req.Name,
		SortOrder: req.SortOrder,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, cat); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return cat, nil
}

func (s *categoryService) GetByID(ctx context.Context, id string) (*models.Category, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return cat, nil
}

func (s *categoryService) ListActive(ctx context.Context, storeID string) ([]*models.Category, error) {
	cats, err := s.repo.ListActiveByStore(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active categories: %w", err)
	}
	return cats, nil
}

func (s *categoryService) ListAll(ctx context.Context, storeID string) ([]*models.Category, error) {
	cats, err := s.repo.ListAllByStore(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return cats, nil
}

func (s *categoryService) Update(ctx context.Context, id string, req *models.UpdateCategoryRequest) (*models.Category, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, fmt.Errorf("failed to find category: %w", err)
	}

	cat.Name = req.Name
	cat.SortOrder = req.SortOrder
	cat.IsActive = req.IsActive
	cat.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, cat); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return cat, nil
}

func (s *categoryService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("failed to delete category: %w", err)
	}
	return nil
}
