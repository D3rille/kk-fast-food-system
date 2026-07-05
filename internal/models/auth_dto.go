package models

// LoginRequest defines the payload for authentication.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserResponse represents the public-facing staff user details returned on login.
type UserResponse struct {
	ID       string   `json:"id"`
	StoreID  string   `json:"store_id"`
	Username string   `json:"username"`
	Role     UserRole `json:"role"`
	IsActive bool     `json:"is_active"`
}

// LoginResponse contains the authentication tokens and user info.
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

// RegisterRequest defines the payload for creating a new staff member.
type RegisterRequest struct {
	StoreID  string   `json:"store_id"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Role     UserRole `json:"role"`
}

// RefreshRequest defines the payload to request a new access token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshResponse contains the newly rotated access and refresh tokens.
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
