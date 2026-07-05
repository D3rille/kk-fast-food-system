package models

// CreateCategoryRequest represents the HTTP body for creating a category.
type CreateCategoryRequest struct {
	StoreID   string `json:"store_id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

// UpdateCategoryRequest represents the HTTP body for updating a category.
type UpdateCategoryRequest struct {
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

// CategoryResponse represents the public category data structure.
type CategoryResponse struct {
	ID        string `json:"id"`
	StoreID   string `json:"store_id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}
