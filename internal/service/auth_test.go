package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/service"
	"golang.org/x/crypto/bcrypt"
)

// Mock UserRepository
type mockUserRepository struct {
	usersByID       map[string]*models.User
	usersByUsername map[string]*models.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		usersByID:       make(map[string]*models.User),
		usersByUsername: make(map[string]*models.User),
	}
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	user, exists := m.usersByID[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user, exists := m.usersByUsername[username]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *mockUserRepository) Create(ctx context.Context, user *models.User) error {
	m.usersByID[user.ID] = user
	m.usersByUsername[user.Username] = user
	return nil
}

// Mock StoreRepository
type mockStoreRepository struct {
	stores map[string]*models.Store
}

func newMockStoreRepository() *mockStoreRepository {
	return &mockStoreRepository{
		stores: make(map[string]*models.Store),
	}
}

func (m *mockStoreRepository) GetByID(ctx context.Context, id string) (*models.Store, error) {
	store, exists := m.stores[id]
	if !exists {
		return nil, errors.New("store not found")
	}
	return store, nil
}

func (m *mockStoreRepository) Create(ctx context.Context, store *models.Store) error {
	m.stores[store.ID] = store
	return nil
}

func TestAuthService_Login(t *testing.T) {
	userRepo := newMockUserRepository()
	storeRepo := newMockStoreRepository()
	secret := "test-secret-key-12345"

	// Create service
	authSrv := service.NewAuthService(userRepo, storeRepo, secret, 5*time.Minute, 10*time.Minute)

	ctx := context.Background()

	// Seed test store
	storeID := "store-123"
	_ = storeRepo.Create(ctx, &models.Store{ID: storeID, Name: "Test Store", IsActive: true})

	// Seed test user
	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		ID:        "user-1",
		StoreID:   storeID,
		Username:  "cashier1",
		Password:  string(hash),
		Role:      models.RoleCashier,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = userRepo.Create(ctx, user)

	t.Run("Successful Login", func(t *testing.T) {
		accessToken, refreshToken, loggedInUser, err := authSrv.Login(ctx, "cashier1", "password123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if accessToken == "" {
			t.Error("expected access token to be generated, got empty string")
		}

		if refreshToken == "" {
			t.Error("expected refresh token to be generated, got empty string")
		}

		if loggedInUser.Username != "cashier1" {
			t.Errorf("expected username to be cashier1, got %s", loggedInUser.Username)
		}
	})

	t.Run("Login Failure - Invalid Password", func(t *testing.T) {
		_, _, _, err := authSrv.Login(ctx, "cashier1", "wrongpassword")
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("Login Failure - User Inactive", func(t *testing.T) {
		user.IsActive = false
		defer func() { user.IsActive = true }() // restore active state

		_, _, _, err := authSrv.Login(ctx, "cashier1", "password123")
		if !errors.Is(err, service.ErrUserInactive) {
			t.Errorf("expected ErrUserInactive, got %v", err)
		}
	})

	t.Run("Login Failure - User Not Found", func(t *testing.T) {
		_, _, _, err := authSrv.Login(ctx, "nonexistent", "password123")
		if !errors.Is(err, service.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func TestAuthService_Refresh(t *testing.T) {
	userRepo := newMockUserRepository()
	storeRepo := newMockStoreRepository()
	secret := "test-secret-key-12345"

	authSrv := service.NewAuthService(userRepo, storeRepo, secret, 5*time.Minute, 10*time.Minute)

	ctx := context.Background()
	storeID := "store-123"
	_ = storeRepo.Create(ctx, &models.Store{ID: storeID, Name: "Test Store", IsActive: true})

	password := "password123"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := &models.User{
		ID:       "user-1",
		StoreID:  storeID,
		Username: "cashier1",
		Password: string(hash),
		Role:     models.RoleCashier,
		IsActive: true,
	}
	_ = userRepo.Create(ctx, user)

	_, refreshToken, _, err := authSrv.Login(ctx, "cashier1", "password123")
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	t.Run("Successful Token Refresh", func(t *testing.T) {
		newAccessToken, newRefreshToken, err := authSrv.Refresh(ctx, refreshToken)
		if err != nil {
			t.Fatalf("expected successful refresh, got: %v", err)
		}

		if newAccessToken == "" || newRefreshToken == "" {
			t.Error("expected new tokens, got empty value")
		}

		if newAccessToken == refreshToken {
			t.Error("expected rotated access token, got same as refresh token")
		}
	})

	t.Run("Refresh Failure - Invalid Signature", func(t *testing.T) {
		invalidSecretSrv := service.NewAuthService(userRepo, storeRepo, "different-secret-key", 5*time.Minute, 10*time.Minute)
		_, _, _, _ = invalidSecretSrv.Login(ctx, "cashier1", "password123")

		_, _, err := authSrv.Refresh(ctx, "invalid-token-string")
		if !errors.Is(err, service.ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
}

func TestAuthService_RegisterStaff(t *testing.T) {
	userRepo := newMockUserRepository()
	storeRepo := newMockStoreRepository()
	secret := "test-secret-key-12345"

	authSrv := service.NewAuthService(userRepo, storeRepo, secret, 5*time.Minute, 10*time.Minute)

	ctx := context.Background()
	storeID := "store-123"
	_ = storeRepo.Create(ctx, &models.Store{ID: storeID, Name: "Test Store", IsActive: true})

	t.Run("Successful Registration", func(t *testing.T) {
		registeredUser, err := authSrv.RegisterStaff(ctx, storeID, "kitchen1", "kitchenpassword", models.RoleKitchen)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if registeredUser.ID == "" {
			t.Error("expected assigned UUID, got empty")
		}

		if registeredUser.Username != "kitchen1" {
			t.Errorf("expected username kitchen1, got %s", registeredUser.Username)
		}

		if registeredUser.Role != models.RoleKitchen {
			t.Errorf("expected role kitchen, got %s", registeredUser.Role)
		}

		// Verify hashed password
		err = bcrypt.CompareHashAndPassword([]byte(registeredUser.Password), []byte("kitchenpassword"))
		if err != nil {
			t.Errorf("expected matching password hash, got error: %v", err)
		}
	})

	t.Run("Registration Failure - Duplicate Username", func(t *testing.T) {
		// Register username first time
		_, _ = authSrv.RegisterStaff(ctx, storeID, "kitchen2", "password", models.RoleKitchen)

		// Try registering duplicate username
		_, err := authSrv.RegisterStaff(ctx, storeID, "kitchen2", "differentpassword", models.RoleKitchen)
		if !errors.Is(err, service.ErrDuplicateUsername) {
			t.Errorf("expected ErrDuplicateUsername, got %v", err)
		}
	})

	t.Run("Registration Failure - Invalid Store ID", func(t *testing.T) {
		_, err := authSrv.RegisterStaff(ctx, "invalid-store", "kitchen3", "password", models.RoleKitchen)
		if err == nil {
			t.Error("expected error due to invalid store id, got nil")
		}
	})
}
