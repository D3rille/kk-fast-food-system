package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/service"
)

// AuthHandler coordinates login and registration HTTP endpoints.
type AuthHandler struct {
	srv service.AuthService
}

// NewAuthHandler instantiates a new AuthHandler.
// func NewAuthHandler(srv service.srv) // Wait, we should type it correctly: service.AuthService
// Let's write the correct constructor.
func NewAuthHandler(srv service.AuthService) *AuthHandler {
	return &AuthHandler{srv: srv}
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid JSON request"}`))
		return
	}

	if req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"username and password are required"}`))
		return
	}

	accessToken, refreshToken, user, err := h.srv.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrUserInactive) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
		return
	}

	resp := models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: models.UserResponse{
			ID:       user.ID,
			StoreID:  user.StoreID,
			Username: user.Username,
			Role:     user.Role,
			IsActive: user.IsActive,
		},
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid JSON request"}`))
		return
	}

	if req.RefreshToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"refresh token is required"}`))
		return
	}

	newAccessToken, newRefreshToken, err := h.srv.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrExpiredToken) || errors.Is(err, service.ErrInvalidToken) || errors.Is(err, service.ErrUserInactive) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
		return
	}

	resp := models.RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Register handles POST /api/v1/admin/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid JSON request"}`))
		return
	}

	if req.StoreID == "" || req.Username == "" || req.Password == "" || req.Role == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"store_id, username, password, and role are required"}`))
		return
	}

	// Validate role enum
	switch req.Role {
	case models.RoleAdmin, models.RoleManager, models.RoleCashier, models.RoleKitchen:
		// valid role
	default:
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid user role"}`))
		return
	}

	createdUser, err := h.srv.RegisterStaff(r.Context(), req.StoreID, req.Username, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateUsername) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := models.UserResponse{
		ID:       createdUser.ID,
		StoreID:  createdUser.StoreID,
		Username: createdUser.Username,
		Role:     createdUser.Role,
		IsActive: createdUser.IsActive,
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}
