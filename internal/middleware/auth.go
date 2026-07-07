package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/service"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

// GetUserID extracts the authenticated user ID from the context.
func GetUserID(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(UserIDKey).(string)
	return val, ok
}

// GetUserRole extracts the authenticated user role from the context.
func GetUserRole(ctx context.Context) (models.UserRole, bool) {
	val, ok := ctx.Value(UserRoleKey).(models.UserRole)
	return val, ok
}

// Authenticate returns a middleware that validates JWT access tokens in the Authorization header.
func Authenticate(jwtSecret string) func(http.Handler) http.Handler {
	secretBytes := []byte(jwtSecret)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			// 1. Get header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"missing authorization header"}`))
				return
			}

			// 2. Parse Bearer token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"authorization header format must be Bearer <token>"}`))
				return
			}

			tokenStr := parts[1]

			// 3. Parse and validate claims
			token, err := jwt.ParseWithClaims(tokenStr, &service.AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return secretBytes, nil
			})

			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				if errors.Is(err, jwt.ErrTokenExpired) {
					_, _ = w.Write([]byte(`{"error":"access token has expired"}`))
				} else {
					_, _ = w.Write([]byte(`{"error":"invalid access token"}`))
				}
				return
			}

			claims, ok := token.Claims.(*service.AuthClaims)
			if !ok || !token.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"invalid access token claims"}`))
				return
			}

			// 4. Ensure it is an ACCESS token, NOT a REFRESH token!
			if claims.TokenType != "access" {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"invalid token type: access token required"}`))
				return
			}

			// 5. Inject claims into context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns a middleware that checks if the authenticated user has one of the allowed roles.
func RequireRole(allowedRoles ...models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			role, ok := GetUserRole(r.Context())
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthenticated"}`))
				return
			}

			// Check if role is in allowed list
			isAllowed := false
			for _, r := range allowedRoles {
				if r == role {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": fmt.Sprintf("forbidden: insufficient permissions. Role '%s' not authorized", role),
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
