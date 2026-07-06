package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/service"
	"github.com/go-chi/chi/v5"
)

// ModifierHandler coordinates HTTP endpoints for modifier groups and options.
type ModifierHandler struct {
	srv service.ModifierService
}

// NewModifierHandler creates a new ModifierHandler.
func NewModifierHandler(srv service.ModifierService) *ModifierHandler {
	return &ModifierHandler{srv: srv}
}

// GetForProduct handles GET /api/v1/menu/items/{id}/modifiers (public kiosk route)
func (h *ModifierHandler) GetForProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	productID := chi.URLParam(r, "id")

	groups, err := h.srv.GetGroupsForProduct(r.Context(), productID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(groups)
}

// CreateGroup handles POST /api/v1/admin/modifiers/groups
func (h *ModifierHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.CreateModifierGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request payload"}`))
		return
	}

	g, err := h.srv.CreateGroup(r.Context(), &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(g)
}

// ListGroups handles GET /api/v1/admin/modifiers/groups
func (h *ModifierHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	groups, err := h.srv.ListGroups(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(groups)
}

// GetGroup handles GET /api/v1/admin/modifiers/groups/{id}
func (h *ModifierHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	g, err := h.srv.GetGroup(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrModifierGroupNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "modifier group not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(g)
}

// UpdateGroup handles PUT /api/v1/admin/modifiers/groups/{id}
func (h *ModifierHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	var req models.UpdateModifierGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request payload"}`))
		return
	}

	g, err := h.srv.UpdateGroup(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrModifierGroupNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "modifier group not found"})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(g)
}

// DeleteGroup handles DELETE /api/v1/admin/modifiers/groups/{id}
func (h *ModifierHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	if err := h.srv.DeleteGroup(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrModifierGroupNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "modifier group not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateOption handles POST /api/v1/admin/modifiers/groups/{id}/options
func (h *ModifierHandler) CreateOption(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	groupID := chi.URLParam(r, "id")

	var req models.CreateModifierOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request payload"}`))
		return
	}

	o, err := h.srv.CreateOption(r.Context(), groupID, &req)
	if err != nil {
		if errors.Is(err, service.ErrModifierGroupNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "modifier group not found"})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(o)
}

// ListOptions handles GET /api/v1/admin/modifiers/groups/{id}/options
func (h *ModifierHandler) ListOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	groupID := chi.URLParam(r, "id")

	options, err := h.srv.ListOptions(r.Context(), groupID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(options)
}

// UpdateOption handles PUT /api/v1/admin/modifiers/options/{optionId}
func (h *ModifierHandler) UpdateOption(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "optionId")

	var req models.UpdateModifierOptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request payload"}`))
		return
	}

	o, err := h.srv.UpdateOption(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, service.ErrModifierOptionNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "modifier option not found"})
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(o)
}

// DeleteOption handles DELETE /api/v1/admin/modifiers/options/{optionId}
func (h *ModifierHandler) DeleteOption(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "optionId")

	if err := h.srv.DeleteOption(r.Context(), id); err != nil {
		if errors.Is(err, service.ErrModifierOptionNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "modifier option not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AttachToProduct handles POST /api/v1/admin/menu/items/{id}/modifier-groups
func (h *ModifierHandler) AttachToProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	productID := chi.URLParam(r, "id")

	var req models.AttachModifierGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request payload"}`))
		return
	}
	if req.ModifierGroupID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"modifier_group_id is required"}`))
		return
	}

	if err := h.srv.AttachToProduct(r.Context(), productID, req.ModifierGroupID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DetachFromProduct handles DELETE /api/v1/admin/menu/items/{id}/modifier-groups/{groupId}
func (h *ModifierHandler) DetachFromProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	productID := chi.URLParam(r, "id")
	groupID := chi.URLParam(r, "groupId")

	if err := h.srv.DetachFromProduct(r.Context(), productID, groupID); err != nil {
		if errors.Is(err, service.ErrModifierGroupNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "association not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
