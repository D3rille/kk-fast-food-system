package models

// CreateProductRequest represents the payload to add a new product.
type CreateProductRequest struct {
	CategoryID  string `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BasePrice   int64  `json:"base_price"` // Centavos (BIGINT)
	ImageURL    string `json:"image_url"`
}

// UpdateProductRequest represents the payload to modify an existing product.
type UpdateProductRequest struct {
	CategoryID  string `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BasePrice   int64  `json:"base_price"` // Centavos
	ImageURL    string `json:"image_url"`
	IsAvailable bool   `json:"is_available"`
}

// ProductResponse represents the serialized product output.
type ProductResponse struct {
	ID          string `json:"id"`
	CategoryID  string `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	BasePrice   int64  `json:"base_price"` // Centavos
	ImageURL    string `json:"image_url"`
	IsAvailable bool   `json:"is_available"`
}

// ToggleAvailabilityResponse represents the output payload for the manager kill-switch.
type ToggleAvailabilityResponse struct {
	ID          string `json:"id"`
	IsAvailable bool   `json:"is_available"`
}
