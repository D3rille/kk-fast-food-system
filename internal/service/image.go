package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/storage"
	"github.com/google/uuid"
)

var (
	ErrUnsupportedImageType = errors.New("unsupported image type: only jpeg, png, webp, and gif are allowed")
	ErrImageTooLarge        = errors.New("image exceeds the maximum allowed size")
)

const imagesPublicPrefix = "/files/images/"

var imageExtensionsByContentType = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// ImageService validates, stores, and removes uploaded image files for menu items.
type ImageService interface {
	Upload(ctx context.Context, u *models.ImageUpload) (imageURL string, err error)
	Delete(ctx context.Context, imageURL string) error
}

type imageService struct {
	storage       storage.ImageStorage
	maxUploadSize int64
}

// NewImageService creates a new ImageService backed by the given ImageStorage.
func NewImageService(imgStorage storage.ImageStorage, maxUploadSize int64) ImageService {
	return &imageService{storage: imgStorage, maxUploadSize: maxUploadSize}
}

func (s *imageService) Upload(ctx context.Context, u *models.ImageUpload) (string, error) {
	if u.Size > s.maxUploadSize {
		return "", ErrImageTooLarge
	}

	sniff := make([]byte, 512)
	n, err := io.ReadFull(u.Reader, sniff)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read image data: %w", err)
	}
	sniff = sniff[:n]

	contentType := http.DetectContentType(sniff)
	ext, ok := imageExtensionsByContentType[contentType]
	if !ok {
		return "", ErrUnsupportedImageType
	}

	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate image filename: %w", err)
	}
	filename := id.String() + ext

	fullReader := io.MultiReader(bytes.NewReader(sniff), u.Reader)
	if err := s.storage.Save(ctx, filename, fullReader); err != nil {
		return "", fmt.Errorf("failed to save image: %w", err)
	}

	return imagesPublicPrefix + filename, nil
}

func (s *imageService) Delete(ctx context.Context, imageURL string) error {
	if imageURL == "" {
		return nil
	}

	filename := path.Base(imageURL)
	if err := s.storage.Delete(ctx, filename); err != nil {
		if errors.Is(err, storage.ErrImageNotFound) {
			return nil
		}
		return fmt.Errorf("failed to delete image: %w", err)
	}
	return nil
}
