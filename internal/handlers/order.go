package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/service"
	"github.com/go-chi/chi/v5"
)

// OrderHandler coordinates HTTP endpoints for order.
type OrderHandler struct {
	srv       service.OrderService
	providers map[string]service.PaymentProvider
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(srv service.OrderService, providers map[string]service.PaymentProvider) *OrderHandler {
	return &OrderHandler{srv: srv, providers: providers}
}

// Create handles POST /api/v1/orders
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req models.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request payload"}`))
		return
	}

	item, err := h.srv.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidModifierSelection) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toOrderResponse(item))
}

// GetByID handles GET /api/v1/orders/{id}
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	detail, err := h.srv.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "order not found"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(detail)
}

// List handles GET /api/v1/orders
// Accepts an optional ?status= query parameter to filter by order status.
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := r.URL.Query().Get("status")
	items, err := h.srv.List(r.Context(), status)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	resp := make([]*models.OrderResponse, len(items))
	for i, item := range items {
		resp[i] = toOrderResponse(item)
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Update handles PUT /api/v1/orders/{id}
func (h *OrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	var req models.UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request payload"}`))
		return
	}

	item, err := h.srv.Update(r.Context(), id, &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toOrderResponse(item))
}

// Delete handles DELETE /api/v1/orders/{id}
func (h *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	if err := h.srv.Delete(r.Context(), id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Checkout handles POST /api/v1/orders/{id}/checkout
// Transitions the order from draft → pending_payment.
func (h *OrderHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	order, err := h.srv.Checkout(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "order not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidStateTransition) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "order is not in draft state"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toOrderResponse(order))
}

// Pay handles POST /api/v1/orders/{id}/pay
// Charges the requested payment provider and transitions the order to paid on success.
func (h *OrderHandler) Pay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	var req models.ProcessPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid request payload"}`))
		return
	}
	if req.Provider == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"provider is required"}`))
		return
	}

	provider, ok := h.providers[req.Provider]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "unsupported payment provider: " + req.Provider})
		return
	}

	payment, err := h.srv.ProcessPayment(r.Context(), id, provider)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "order not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidStateTransition) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "order is not in pending_payment state"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(toPaymentResponse(payment))
}

func toOrderResponse(item *models.Order) *models.OrderResponse {
	return &models.OrderResponse{
		ID:            item.ID,
		StoreID:       item.StoreID,
		OrderNumber:   item.OrderNumber,
		Source:        item.Source,
		Status:        item.Status,
		PaymentStatus: item.PaymentStatus,
		TotalAmount:   item.TotalAmount,
		CreatedAt:     item.CreatedAt,
		UpdatedAt:     item.UpdatedAt,
	}
}

func toPaymentResponse(p *models.Payment) *models.PaymentResponse {
	return &models.PaymentResponse{
		ID:             p.ID,
		OrderID:        p.OrderID,
		Provider:       p.Provider,
		Amount:         p.Amount,
		Status:         p.Status,
		TransactionRef: p.TransactionRef,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
