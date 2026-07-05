package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var ErrImageNotFound = errors.New("image file not found")

// ImageStorage defines local disk access for uploaded image files.
type ImageStorage interface {
	Save(ctx context.Context, filename string, r io.Reader) error
	Delete(ctx context.Context, filename string) error
}

type localImageStorage struct {
	dir string
}

// NewLocalImageStorage creates a disk-backed ImageStorage rooted at dir, creating it if missing.
func NewLocalImageStorage(dir string) (ImageStorage, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create images directory: %w", err)
	}
	return &localImageStorage{dir: dir}, nil
}

func (s *localImageStorage) Save(_ context.Context, filename string, r io.Reader) error {
	path := filepath.Join(s.dir, filepath.Base(filename))

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to write image file: %w", err)
	}
	return nil
}

func (s *localImageStorage) Delete(_ context.Context, filename string) error {
	path := filepath.Join(s.dir, filepath.Base(filename))

	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrImageNotFound
		}
		return fmt.Errorf("failed to delete image file: %w", err)
	}
	return nil
}
