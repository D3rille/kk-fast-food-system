package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func main() {
	force := flag.Bool("force", false, "Overwrite existing files if they exist")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Error: Missing feature name.")
		fmt.Println("Usage: go run scripts/scaffold.go [-force] <FeatureName>")
		fmt.Println("Example: go run scripts/scaffold.go Product")
		os.Exit(1)
	}

	rawFeatureName := args[0]
	if !isValidFeatureName(rawFeatureName) {
		fmt.Printf("Error: Invalid feature name '%s'. Use alpha-numeric names starting with a letter.\n", rawFeatureName)
		os.Exit(1)
	}

	// Calculate name casings
	featurePascal := toPascalCase(rawFeatureName)
	featureCamel := toCamelCase(featurePascal)
	featureSnake := toSnakeCase(featurePascal)
	featurePlural := pluralize(featureSnake)

	fmt.Printf("Generating scaffolding for feature: %s\n", featurePascal)
	fmt.Printf("  PascalCase: %s\n", featurePascal)
	fmt.Printf("  camelCase:  %s\n", featureCamel)
	fmt.Printf("  snake_case: %s\n", featureSnake)
	fmt.Printf("  Pluralized: %s\n\n", featurePlural)

	// Ensure we are in project root (looking for go.mod)
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		fmt.Println("Error: This script must be run from the project root directory containing go.mod.")
		os.Exit(1)
	}

	// Define files to create
	files := []struct {
		path     string
		content  string
		category string
	}{
		{
			path:     filepath.Join("internal", "models", featureSnake+"_dto.go"),
			content:  generateDTO(featurePascal),
			category: "DTO model file",
		},
		{
			path:     filepath.Join("internal", "repository", featureSnake+".go"),
			content:  generateRepository(featurePascal, featureCamel),
			category: "Repository file",
		},
		{
			path:     filepath.Join("internal", "service", featureSnake+".go"),
			content:  generateService(featurePascal, featureCamel),
			category: "Service file",
		},
		{
			path:     filepath.Join("internal", "handlers", featureSnake+".go"),
			content:  generateHandler(featurePascal, featureCamel, featureSnake, featurePlural),
			category: "Handler file",
		},
		{
			path:     filepath.Join("internal", "service", featureSnake+"_test.go"),
			content:  generateTest(featurePascal, featureCamel),
			category: "Unit test file",
		},
	}

	// Process files
	for _, f := range files {
		// Ensure parent directory exists
		dir := filepath.Dir(f.path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}

		// Check if file exists
		if _, err := os.Stat(f.path); err == nil && !*force {
			fmt.Printf("Skipping existing file (use -force to overwrite): %s\n", f.path)
			continue
		}

		// Write file
		err := os.WriteFile(f.path, []byte(f.content), 0644)
		if err != nil {
			fmt.Printf("Error writing file %s: %v\n", f.path, err)
			os.Exit(1)
		}
		fmt.Printf("✔ Created %s -> %s\n", f.category, f.path)
	}

	// Print wiring instructions
	fmt.Println("\n==========================================================================")
	fmt.Println("Scaffolding Complete! To wire up this new feature:")
	fmt.Println("==========================================================================")
	fmt.Println("1. Ensure your database migration is created and run:")
	fmt.Printf("   make migrate-create name=create_%s_table\n", featurePlural)
	fmt.Println("\n2. In internal/models/models.go, add your domain model struct:")
	fmt.Printf("   type %s struct {\n", featurePascal)
	fmt.Println("       ID        string    `json:\"id\" db:\"id\"`\n       // Add other database fields here...")
	fmt.Println("   }")
	fmt.Println("\n3. In cmd/api/main.go, wire up the dependencies:")
	fmt.Printf("   // Repositories\n   %sRepo := repository.New%sRepository(db)\n", featureCamel, featurePascal)
	fmt.Printf("   // Services\n   %sSrv := service.New%sService(%sRepo)\n", featureCamel, featurePascal, featureCamel)
	fmt.Printf("   // Handlers\n   %sHandler := handlers.New%sHandler(%sSrv)\n", featureCamel, featurePascal, featureCamel)
	fmt.Println("\n4. Register the HTTP routes in cmd/api/main.go:")
	fmt.Printf("   r.Route(\"/api/v1/%s\", func(r chi.Router) {\n", featurePlural)
	fmt.Printf("       r.Post(\"/\", %sHandler.Create)\n", featureCamel)
	fmt.Printf("       r.Get(\"/{id}\", %sHandler.GetByID)\n", featureCamel)
	fmt.Printf("       r.Get(\"/\", %sHandler.List)\n", featureCamel)
	fmt.Printf("       r.Put(\"/{id}\", %sHandler.Update)\n", featureCamel)
	fmt.Printf("       r.Delete(\"/{id}\", %sHandler.Delete)\n", featureCamel)
	fmt.Println("   })")
	fmt.Println("==========================================================================\n")
}

func isValidFeatureName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 && !unicode.IsLetter(r) {
			return false
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '-' {
			return false
		}
	}
	return true
}

func toPascalCase(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, "")
}

func toCamelCase(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func toSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

func pluralize(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") {
		return s
	}
	return s + "s"
}

// Templates

func generateDTO(pascal string) string {
	return fmt.Sprintf(`package models

// Create%sRequest defines the API payload for creation.
type Create%sRequest struct {
	// TODO: Add incoming request fields (e.g. Name string `+"`json:\"name\"`"+`)
}

// Update%sRequest defines the API payload for updating.
type Update%sRequest struct {
	// TODO: Add update fields here
}

// %sResponse defines the serialized public API response structure.
type %sResponse struct {
	ID        string `+"`json:\"id\"`"+`
	// TODO: Add response fields
}
`, pascal, pascal, pascal, pascal, pascal, pascal)
}

func generateRepository(pascal, camel string) string {
	return fmt.Sprintf(`package repository

import (
	"context"
	"fmt"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// %sRepository defines all database access methods for %s.
type %sRepository interface {
	Create(ctx context.Context, item *models.%s) error
	GetByID(ctx context.Context, id string) (*models.%s, error)
	List(ctx context.Context) ([]*models.%s, error)
	Update(ctx context.Context, item *models.%s) error
	Delete(ctx context.Context, id string) error
}

type postgres%sRepository struct {
	db *pgxpool.Pool
}

// New%sRepository creates a concrete repository implementation.
func New%sRepository(db *pgxpool.Pool) %sRepository {
	return &postgres%sRepository{db: db}
}

func (r *postgres%sRepository) Create(ctx context.Context, item *models.%s) error {
	query := `+"`"+`
		INSERT INTO %[2]s (id, created_at, updated_at)
		VALUES ($1, $2, $3)
	`+"`"+`
	_, err := r.db.Exec(ctx, query, item.ID, item.CreatedAt, item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create %[2]s: %%w", err)
	}
	return nil
}

func (r *postgres%sRepository) GetByID(ctx context.Context, id string) (*models.%s, error) {
	query := `+"`"+`
		SELECT id, created_at, updated_at
		FROM %[2]s
		WHERE id = $1
	`+"`"+`
	var item models.%s
	err := r.db.QueryRow(ctx, query, id).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %[2]s by id: %%w", err)
	}
	return &item, nil
}

func (r *postgres%sRepository) List(ctx context.Context) ([]*models.%s, error) {
	query := `+"`"+`
		SELECT id, created_at, updated_at
		FROM %[2]s
		ORDER BY created_at DESC
	`+"`"+`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query %[2]s: %%w", err)
	}
	defer rows.Close()

	var items []*models.%s
	for rows.Next() {
		var item models.%s
		if err := rows.Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan %[2]s row: %%w", err)
		}
		items = append(items, &item)
	}
	return items, nil
}

func (r *postgres%sRepository) Update(ctx context.Context, item *models.%s) error {
	query := `+"`"+`
		UPDATE %[2]s
		SET updated_at = $2
		WHERE id = $1
	`+"`"+`
	commandTag, err := r.db.Exec(ctx, query, item.ID, item.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update %[2]s: %%w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no %[2]s rows updated")
	}
	return nil
}

func (r *postgres%sRepository) Delete(ctx context.Context, id string) error {
	query := `+"`"+`
		DELETE FROM %[2]s
		WHERE id = $1
	`+"`"+`
	commandTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete %[2]s: %%w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no %[2]s rows deleted")
	}
	return nil
}
`, pascal, camel, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal)
}

func generateService(pascal, camel string) string {
	return fmt.Sprintf(`package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/repository"
)

var (
	Err%sNotFound = errors.New("%s not found")
)

// %sService defines business operations for %s.
type %sService interface {
	Create(ctx context.Context, req *models.Create%sRequest) (*models.%s, error)
	GetByID(ctx context.Context, id string) (*models.%s, error)
	List(ctx context.Context) ([]*models.%s, error)
	Update(ctx context.Context, id string, req *models.Update%sRequest) (*models.%s, error)
	Delete(ctx context.Context, id string) error
}

type %[2]sService struct {
	repo repository.%sRepository
}

// New%sService instantiates a new %sService.
func New%sService(repo repository.%sRepository) %sService {
	return &%[2]sService{repo: repo}
}

func (s *%[2]sService) Create(ctx context.Context, req *models.Create%sRequest) (*models.%s, error) {
	// TODO: Add input business logic validation

	item := &models.%s{
		ID:        "%[2]s_" + "uniqueuuidplaceholder", // TODO: replace with standard UUID generator
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("service failed to create %[2]s: %%w", err)
	}
	return item, nil
}

func (s *%[2]sService) GetByID(ctx context.Context, id string) (*models.%s, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		// Replace this check with check for pgx.ErrNoRows if necessary
		return nil, fmt.Errorf("service failed to find %[2]s: %%w", err)
	}
	return item, nil
}

func (s *%[2]sService) List(ctx context.Context) ([]*models.%s, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("service failed to list %[2]s: %%w", err)
	}
	return items, nil
}

func (s *%[2]sService) Update(ctx context.Context, id string, req *models.Update%sRequest) (*models.%s, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("service failed to find %[2]s for update: %%w", err)
	}

	// TODO: Apply update fields from req to item
	item.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("service failed to update %[2]s: %%w", err)
	}
	return item, nil
}

func (s *%[2]sService) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("service failed to delete %[2]s: %%w", err)
	}
	return nil
}
`, pascal, camel, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal, pascal)
}

func generateHandler(pascal, camel, snake, plural string) string {
	return fmt.Sprintf(`package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/service"
	"github.com/go-chi/chi/v5"
)

// %sHandler coordinates HTTP endpoints for %s.
type %sHandler struct {
	srv service.%sService
}

// New%sHandler creates a new %sHandler.
func New%sHandler(srv service.%sService) *%sHandler {
	return &%sHandler{srv: srv}
}

// Create handles POST /api/v1/%s
func (h *%sHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.Create%sRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`+"`"+`{"error":"invalid request payload"}`+"`"+`))
		return
	}

	item, err := h.srv.Create(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := to%sResponse(item)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetByID handles GET /api/v1/%s/{id}
func (h *%sHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	item, err := h.srv.GetByID(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := to%sResponse(item)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// List handles GET /api/v1/%s
func (h *%sHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	items, err := h.srv.List(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := make([]*models.%sResponse, len(items))
	for i, item := range items {
		resp[i] = to%sResponse(item)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Update handles PUT /api/v1/%s/{id}
func (h *%sHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	var req models.Update%sRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`+"`"+`{"error":"invalid request payload"}`+"`"+`))
		return
	}

	item, err := h.srv.Update(r.Context(), id, &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := to%sResponse(item)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Delete handles DELETE /api/v1/%s/{id}
func (h *%sHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	if err := h.srv.Delete(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func to%sResponse(item *models.%s) *models.%sResponse {
	return &models.%sResponse{
		ID: item.ID,
		// TODO: Map custom fields
	}
}
`, pascal, camel, pascal, pascal, pascal, pascal, pascal, pascal, plural, pascal, pascal, pascal, plural, pascal, plural, pascal, pascal, plural, pascal, pascal, plural, pascal, pascal, pascal, pascal)
}

func generateTest(pascal, camel string) string {
	return fmt.Sprintf(`package service_test

import (
	"context"
	"testing"
)

func Test%sService_Create(t *testing.T) {
	ctx := context.Background()
	_ = ctx
	// TODO: Implement mock unit tests for %sService
}
`, pascal, pascal)
}
