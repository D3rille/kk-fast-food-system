package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/service"
	"github.com/go-chi/chi/v5"
)

// ProductHandler coordinates HTTP endpoints for products.
type ProductHandler struct {
	srv           service.ProductService
	maxUploadSize int64
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(srv service.ProductService, maxUploadSize int64) *ProductHandler {
	return &ProductHandler{srv: srv, maxUploadSize: maxUploadSize}
}

// parseProductImage extracts an optional "image" multipart file part from the request.
// It returns (nil, nil) when the caller did not include an image part.
func (h *ProductHandler) parseProductImage(r *http.Request) (*models.ImageUpload, error) {
	file, header, err := r.FormFile("image")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	return &models.ImageUpload{
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		Reader:      file,
	}, nil
}

// Create handles POST /api/v1/admin/menu/items
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadSize)
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid multipart form data or file too large"}`))
		return
	}

	req := models.CreateProductRequest{
		CategoryID:  r.FormValue("category_id"),
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}
	if basePrice, err := strconv.ParseInt(r.FormValue("base_price"), 10, 64); err == nil {
		req.BasePrice = basePrice
	}

	if req.CategoryID == "" || req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"category_id and name are required"}`))
		return
	}
	if req.BasePrice <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"base_price must be greater than zero"}`))
		return
	}

	image, err := h.parseProductImage(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid image upload"}`))
		return
	}

	p, err := h.srv.Create(r.Context(), &req, image)
	if err != nil {
		if errors.Is(err, service.ErrUnsupportedImageType) || errors.Is(err, service.ErrImageTooLarge) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toProductResponse(p))
}

// GetByID handles GET /api/v1/admin/menu/items/{id}
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	p, err := h.srv.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "product not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toProductResponse(p))
}

// List handles GET /api/v1/admin/menu/items[?category_id=<id>]
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	categoryID := r.URL.Query().Get("category_id")

	items, err := h.srv.List(r.Context(), categoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := make([]*models.ProductResponse, len(items))
	for i, p := range items {
		resp[i] = toProductResponse(p)
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// ListAvailable handles GET /api/v1/menu/items?category_id=<id> (public kiosk route)
func (h *ProductHandler) ListAvailable(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	categoryID := r.URL.Query().Get("category_id")
	if categoryID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"category_id query parameter is required"}`))
		return
	}

	items, err := h.srv.ListAvailable(r.Context(), categoryID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := make([]*models.ProductResponse, len(items))
	for i, p := range items {
		resp[i] = toProductResponse(p)
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Update handles PUT /api/v1/admin/menu/items/{id}
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadSize)
	if err := r.ParseMultipartForm(h.maxUploadSize); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid multipart form data or file too large"}`))
		return
	}

	req := models.UpdateProductRequest{
		CategoryID:  r.FormValue("category_id"),
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
	}
	if basePrice, err := strconv.ParseInt(r.FormValue("base_price"), 10, 64); err == nil {
		req.BasePrice = basePrice
	}
	if isAvailable, err := strconv.ParseBool(r.FormValue("is_available")); err == nil {
		req.IsAvailable = isAvailable
	}
	removeImage := r.FormValue("remove_image") == "true"

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"name is required"}`))
		return
	}
	if req.BasePrice <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"base_price must be greater than zero"}`))
		return
	}

	image, err := h.parseProductImage(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid image upload"}`))
		return
	}

	p, err := h.srv.Update(r.Context(), id, &req, image, removeImage)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "product not found"})
			return
		}
		if errors.Is(err, service.ErrUnsupportedImageType) || errors.Is(err, service.ErrImageTooLarge) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toProductResponse(p))
}

// Delete handles DELETE /api/v1/admin/menu/items/{id}
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	if err := h.srv.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "product not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ToggleAvailability handles PATCH /api/v1/admin/menu/items/{id}/availability
func (h *ProductHandler) ToggleAvailability(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	p, err := h.srv.ToggleAvailability(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "product not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(models.ToggleAvailabilityResponse{
		ID:          p.ID,
		IsAvailable: p.IsAvailable,
	})
}

func toProductResponse(p *models.Product) *models.ProductResponse {
	return &models.ProductResponse{
		ID:          p.ID,
		CategoryID:  p.CategoryID,
		Name:        p.Name,
		Description: p.Description,
		BasePrice:   p.BasePrice,
		ImageURL:    p.ImageURL,
		IsAvailable: p.IsAvailable,
	}
}
