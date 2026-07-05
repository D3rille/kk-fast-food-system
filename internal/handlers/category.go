package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/service"
	"github.com/go-chi/chi/v5"
)

// CategoryHandler coordinates HTTP endpoints for categories.
type CategoryHandler struct {
	srv service.CategoryService
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(srv service.CategoryService) *CategoryHandler {
	return &CategoryHandler{srv: srv}
}

// Create handles POST /api/v1/admin/menu/categories
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request payload"}`))
		return
	}

	if req.StoreID == "" || req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"store_id and name are required"}`))
		return
	}

	cat, err := h.srv.Create(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toCategoryResponse(cat))
}

// GetByID handles GET /api/v1/admin/menu/categories/{id}
func (h *CategoryHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	cat, err := h.srv.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "category not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toCategoryResponse(cat))
}

// ListAll handles GET /api/v1/admin/menu/categories?store_id=<id>
func (h *CategoryHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"store_id query parameter is required"}`))
		return
	}

	cats, err := h.srv.ListAll(r.Context(), storeID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := make([]*models.CategoryResponse, len(cats))
	for i, c := range cats {
		resp[i] = toCategoryResponse(c)
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// ListActive handles GET /api/v1/menu/categories?store_id=<id> (public kiosk route)
func (h *CategoryHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	storeID := r.URL.Query().Get("store_id")
	if storeID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"store_id query parameter is required"}`))
		return
	}

	cats, err := h.srv.ListActive(r.Context(), storeID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := make([]*models.CategoryResponse, len(cats))
	for i, c := range cats {
		resp[i] = toCategoryResponse(c)
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Update handles PUT /api/v1/admin/menu/categories/{id}
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	var req models.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request payload"}`))
		return
	}

	if req.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"name is required"}`))
		return
	}

	cat, err := h.srv.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "category not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toCategoryResponse(cat))
}

// Delete handles DELETE /api/v1/admin/menu/categories/{id}
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	if err := h.srv.Delete(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "category not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toCategoryResponse(cat *models.Category) *models.CategoryResponse {
	return &models.CategoryResponse{
		ID:        cat.ID,
		StoreID:   cat.StoreID,
		Name:      cat.Name,
		SortOrder: cat.SortOrder,
		IsActive:  cat.IsActive,
	}
}
