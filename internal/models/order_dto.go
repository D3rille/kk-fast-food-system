package models

import "time"

// CreateOrderItemRequest is a single line item within a CreateOrderRequest.
type CreateOrderItemRequest struct {
	ProductID         string   `json:"product_id" validate:"required"`
	ModifierOptionIDs []string `json:"modifier_option_ids"`
	Quantity          int      `json:"quantity"   validate:"required"`
	UnitPrice         int64    `json:"unit_price" validate:"required"`
}

// CreateOrderRequest defines the API payload for order creation.
// Status, PaymentStatus, and OrderNumber are not accepted from the client — the service always
// creates orders as draft/pending and generates the order number server-side.
type CreateOrderRequest struct {
	StoreID     string                   `json:"store_id"     validate:"required"`
	Source      OrderSource              `json:"source"       validate:"required"`
	Items       []CreateOrderItemRequest `json:"items"`
	TotalAmount int64                    `json:"total_amount" validate:"required"`
}

// UpdateOrderRequest defines the API payload for updating an order.
type UpdateOrderRequest struct {
	ID            string        `json:"id"  validate:"required"`
	StoreID       string        `json:"store_id"`
	Source        OrderSource   `json:"source"`
	Status        OrderStatus   `json:"status"`
	PaymentStatus PaymentStatus `json:"payment_status"`
	OrderNumber   int           `json:"order_number"`
	TotalAmount   int64         `json:"total_amount"`
}

// ProcessPaymentRequest defines the API payload for initiating payment on an order.
type ProcessPaymentRequest struct {
	Provider string `json:"provider" validate:"required"`
}

// OrderResponse defines the serialized public API response structure for an order.
type OrderResponse struct {
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	ID            string        `json:"id"`
	StoreID       string        `json:"store_id"`
	Source        OrderSource   `json:"source"`
	Status        OrderStatus   `json:"status"`
	PaymentStatus PaymentStatus `json:"payment_status"`
	OrderNumber   int           `json:"order_number"`
	TotalAmount   int64         `json:"total_amount"`
}

// OrderItemModifierResponse is a single selected modifier option attached to an order item.
type OrderItemModifierResponse struct {
	ModifierOptionID string `json:"modifier_option_id"`
	GroupName        string `json:"group_name"`
	Name             string `json:"name"`
	ExtraPrice       int64  `json:"extra_price"`
}

// OrderItemResponse is a single line item within an OrderDetailResponse.
type OrderItemResponse struct {
	CreatedAt   time.Time                   `json:"created_at"`
	ID          string                      `json:"id"`
	ProductID   string                      `json:"product_id"`
	ProductName string                      `json:"product_name"`
	Modifiers   []OrderItemModifierResponse `json:"modifiers"`
	Quantity    int                         `json:"quantity"`
	UnitPrice   int64                       `json:"unit_price"`
	Subtotal    int64                       `json:"subtotal"`
}

// OrderDetailResponse extends OrderResponse with the order's line items.
type OrderDetailResponse struct {
	Items []OrderItemResponse `json:"items"`
	OrderResponse
}

// PaymentResponse defines the serialized payment record returned after a payment attempt.
type PaymentResponse struct {
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	ID             string          `json:"id"`
	OrderID        string          `json:"order_id"`
	Provider       PaymentProvider `json:"provider"`
	Status         PaymentStatus   `json:"status"`
	TransactionRef string          `json:"transaction_ref"`
	Amount         int64           `json:"amount"`
}
