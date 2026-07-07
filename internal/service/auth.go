package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserInactive       = errors.New("user account is deactivated")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token has expired")
	ErrDuplicateUsername  = errors.New("username is already taken")
)

// AuthClaims defines custom JWT claims for access and refresh tokens.
type AuthClaims struct {
	UserID    string          `json:"user_id"`
	Role      models.UserRole `json:"role"`
	TokenType string          `json:"token_type"`
	jwt.RegisteredClaims
}

// AuthService defines business operations for authentication and staff registration.
type AuthService interface {
	Login(ctx context.Context, username, password string) (string, string, *models.User, error)
	Refresh(ctx context.Context, refreshTokenStr string) (string, string, error)
	RegisterStaff(ctx context.Context, storeID, username, password string, role models.UserRole) (*models.User, error)
}

type authService struct {
	userRepo        repository.UserRepository
	storeRepo       repository.StoreRepository
	jwtSecret       []byte
	accessTokenDur  time.Duration
	refreshTokenDur time.Duration
}

// NewAuthService creates a concrete implementation of AuthService.
func NewAuthService(
	userRepo repository.UserRepository,
	storeRepo repository.StoreRepository,
	jwtSecret string,
	accessTokenDur time.Duration,
	refreshTokenDur time.Duration,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		storeRepo:       storeRepo,
		jwtSecret:       []byte(jwtSecret),
		accessTokenDur:  accessTokenDur,
		refreshTokenDur: refreshTokenDur,
	}
}

func (s *authService) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, user *models.User, err error) {
	// 1. Fetch user
	user, err = s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	// 2. Check if active
	if !user.IsActive {
		return "", "", nil, ErrUserInactive
	}

	// 3. Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", nil, ErrInvalidCredentials
	}

	// 4. Generate tokens
	accessToken, err = s.generateToken(user.ID, user.Role, "access", s.accessTokenDur)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = s.generateToken(user.ID, user.Role, "refresh", s.refreshTokenDur)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, user, nil
}

func (s *authService) Refresh(ctx context.Context, refreshTokenStr string) (accessToken, refreshToken string, err error) {
	// 1. Parse and validate refresh token
	token, err := jwt.ParseWithClaims(refreshTokenStr, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return "", "", ErrExpiredToken
		}
		return "", "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok || !token.Valid || claims.TokenType != "refresh" {
		return "", "", ErrInvalidToken
	}

	// 2. Verify user is active and exists
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return "", "", ErrInvalidToken
	}
	if !user.IsActive {
		return "", "", ErrUserInactive
	}

	// 3. Generate a new access token and a new refresh token (Token Rotation)
	newAccessToken, err := s.generateToken(user.ID, user.Role, "access", s.accessTokenDur)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	newRefreshToken, err := s.generateToken(user.ID, user.Role, "refresh", s.refreshTokenDur)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	return newAccessToken, newRefreshToken, nil
}

func (s *authService) RegisterStaff(ctx context.Context, storeID, username, password string, role models.UserRole) (*models.User, error) {
	// 1. Verify store exists
	_, err := s.storeRepo.GetByID(ctx, storeID)
	if err != nil {
		return nil, fmt.Errorf("invalid store id: %w", err)
	}

	// 2. Check if username exists
	existingUser, _ := s.userRepo.GetByUsername(ctx, username)
	if existingUser != nil {
		return nil, ErrDuplicateUsername
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Create user
	user := &models.User{
		ID:        generateUUID(), // Generates temporary UUID
		StoreID:   storeID,
		Username:  username,
		Password:  string(hashedPassword),
		Role:      role,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

func (s *authService) generateToken(userID string, role models.UserRole, tokenType string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := AuthClaims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateUUID() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "00000000-0000-0000-0000-000000000000"
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
