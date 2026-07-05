package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/repository"
	"github.com/google/uuid"
)

var ErrProductNotFound = errors.New("product not found")

// ProductService defines business operations for products.
type ProductService interface {
	Create(ctx context.Context, req *models.CreateProductRequest, image *models.ImageUpload) (*models.Product, error)
	GetByID(ctx context.Context, id string) (*models.Product, error)
	List(ctx context.Context, categoryID string) ([]*models.Product, error)
	ListAvailable(ctx context.Context, categoryID string) ([]*models.Product, error)
	Update(ctx context.Context, id string, req *models.UpdateProductRequest, image *models.ImageUpload, removeImage bool) (*models.Product, error)
	Delete(ctx context.Context, id string) error
	ToggleAvailability(ctx context.Context, id string) (*models.Product, error)
}

type productService struct {
	repo     repository.ProductRepository
	imageSvc ImageService
	log      *slog.Logger
}

// NewProductService instantiates a new ProductService.
func NewProductService(repo repository.ProductRepository, imageSvc ImageService, log *slog.Logger) ProductService {
	return &productService{repo: repo, imageSvc: imageSvc, log: log}
}

func (s *productService) Create(ctx context.Context, req *models.CreateProductRequest, image *models.ImageUpload) (*models.Product, error) {
	if req.CategoryID == "" {
		return nil, fmt.Errorf("category_id is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.BasePrice <= 0 {
		return nil, fmt.Errorf("base_price must be greater than zero")
	}

	if image != nil {
		imageURL, err := s.imageSvc.Upload(ctx, image)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image: %w", err)
		}
		req.ImageURL = imageURL
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ID: %w", err)
	}

	now := time.Now()
	p := &models.Product{
		ID:          id.String(),
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		BasePrice:   req.BasePrice,
		ImageURL:    req.ImageURL,
		IsAvailable: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}
	return p, nil
}

func (s *productService) GetByID(ctx context.Context, id string) (*models.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return p, nil
}

// List returns all products when categoryID is empty, or filters by category when provided.
func (s *productService) List(ctx context.Context, categoryID string) ([]*models.Product, error) {
	if categoryID == "" {
		items, err := s.repo.ListAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list products: %w", err)
		}
		return items, nil
	}
	items, err := s.repo.ListByCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list products by category: %w", err)
	}
	return items, nil
}

func (s *productService) ListAvailable(ctx context.Context, categoryID string) ([]*models.Product, error) {
	items, err := s.repo.ListAvailableByCategory(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list available products: %w", err)
	}
	return items, nil
}

func (s *productService) Update(ctx context.Context, id string, req *models.UpdateProductRequest, image *models.ImageUpload, removeImage bool) (*models.Product, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.BasePrice <= 0 {
		return nil, fmt.Errorf("base_price must be greater than zero")
	}

	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	oldImageURL := p.ImageURL
	newImageURL := oldImageURL
	switch {
	case image != nil:
		newImageURL, err = s.imageSvc.Upload(ctx, image)
		if err != nil {
			return nil, fmt.Errorf("failed to upload image: %w", err)
		}
	case removeImage:
		newImageURL = ""
	}

	p.CategoryID = req.CategoryID
	p.Name = req.Name
	p.Description = req.Description
	p.BasePrice = req.BasePrice
	p.ImageURL = newImageURL
	p.IsAvailable = req.IsAvailable
	p.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	if newImageURL != oldImageURL && oldImageURL != "" {
		if err := s.imageSvc.Delete(ctx, oldImageURL); err != nil {
			s.log.Warn("failed to delete replaced product image", "product_id", id, "image_url", oldImageURL, "error", err)
		}
	}

	return p, nil
}

func (s *productService) Delete(ctx context.Context, id string) error {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return ErrProductNotFound
		}
		return fmt.Errorf("failed to find product: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return ErrProductNotFound
		}
		return fmt.Errorf("failed to delete product: %w", err)
	}

	if p.ImageURL != "" {
		if err := s.imageSvc.Delete(ctx, p.ImageURL); err != nil {
			s.log.Warn("failed to delete product image", "product_id", id, "image_url", p.ImageURL, "error", err)
		}
	}

	return nil
}

func (s *productService) ToggleAvailability(ctx context.Context, id string) (*models.Product, error) {
	p, err := s.repo.ToggleAvailability(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to toggle product availability: %w", err)
	}
	return p, nil
}
