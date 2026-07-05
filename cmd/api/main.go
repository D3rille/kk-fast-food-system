package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/D3rille/kk-fast-food-system/internal/config"
	"github.com/D3rille/kk-fast-food-system/internal/database"
	"github.com/D3rille/kk-fast-food-system/internal/handlers"
	"github.com/D3rille/kk-fast-food-system/internal/logger"
	"github.com/D3rille/kk-fast-food-system/internal/middleware"
	"github.com/D3rille/kk-fast-food-system/internal/models"
	"github.com/D3rille/kk-fast-food-system/internal/repository"
	"github.com/D3rille/kk-fast-food-system/internal/service"
	"github.com/D3rille/kk-fast-food-system/internal/storage"
	"github.com/D3rille/kk-fast-food-system/internal/ws"
)

func main() {
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize structured logger
	log := logger.New(cfg.App.Env, cfg.App.LogLevel)
	log.Info("starting NextGen Kiosk-to-Kitchen Fast Food API Server",
		"env", cfg.App.Env,
		"log_level", cfg.App.LogLevel,
	)

	// Create a context that listens for signals to stop
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Connect to database
	db, err := database.New(ctx, cfg.Database.URL, cfg.Database.MaxConns, cfg.Database.MinConns, log)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// 4. Initialize Chi Router & middleware
	r := chi.NewRouter()
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.Recovery(log))
	r.Use(middleware.Logging(log))
	r.Use(middleware.CORS())

	// 5. Initialize Repositories
	storeRepo := repository.NewStoreRepository(db)
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	orderItemRepo := repository.NewOrderItemRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productRepo := repository.NewProductRepository(db)
	modifierRepo := repository.NewModifierRepository(db)

	// 6. Initialize WebSocket hub and start event fan-out goroutine
	wsHub := ws.NewHub(log)
	go wsHub.Run()

	// 6b. Initialize local image storage for menu item photos
	imageStorage, err := storage.NewLocalImageStorage(cfg.Storage.ImagesDir)
	if err != nil {
		log.Error("failed to initialize image storage", "error", err)
		os.Exit(1)
	}

	// 7. Initialize Services
	authService := service.NewAuthService(userRepo, storeRepo, cfg.JWT.Secret, cfg.JWT.Expiration, cfg.JWT.RefreshExpiration)
	orderService := service.NewOrderServiceWithItems(orderRepo, orderItemRepo, paymentRepo, productRepo, modifierRepo, wsHub)
	categoryService := service.NewCategoryService(categoryRepo)
	imageService := service.NewImageService(imageStorage, cfg.Storage.MaxUploadSizeMB*1024*1024)
	productService := service.NewProductService(productRepo, imageService, log)
	modifierService := service.NewModifierService(modifierRepo)

	// Payment providers registry
	paymentProviders := map[string]service.PaymentProvider{
		string(models.ProviderCash): service.NewCashProvider(),
	}

	// 8. Initialize Handlers
	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(authService)
	orderHandler := handlers.NewOrderHandler(orderService, paymentProviders)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	productHandler := handlers.NewProductHandler(productService, cfg.Storage.MaxUploadSizeMB*1024*1024)
	modifierHandler := handlers.NewModifierHandler(modifierService)

	// 9. Mount health handlers
	r.Get("/healthz", healthHandler.Healthz)
	r.Get("/readyz", healthHandler.Readyz)

	// WebSocket endpoint for kitchen display clients (local-first LAN — no auth required)
	r.Get("/ws/kitchen", ws.ServeKitchen(wsHub, log))

	// Static file serving for uploaded menu item images (public — kiosk/cashier/admin/KDS all load these unauthenticated)
	r.Handle("/files/images/*", http.StripPrefix("/files/images/", http.FileServer(http.Dir(cfg.Storage.ImagesDir))))

	// API Routes Group
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"message":"Welcome to NextGen Kiosk-to-Kitchen API v1"}`))
		})

		// Public Auth Routes
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/refresh", authHandler.Refresh)

		// Public Kiosk Menu Routes
		r.Route("/menu", func(r chi.Router) {
			r.Get("/categories", categoryHandler.ListActive)
			r.Get("/items", productHandler.ListAvailable)
			r.Get("/items/{id}/modifiers", modifierHandler.GetForProduct)
		})

		// Protected Admin Routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			r.Use(middleware.RequireRole(models.RoleAdmin, models.RoleManager))

			r.Post("/admin/register", authHandler.Register)

			r.Route("/admin/menu", func(r chi.Router) {
				r.Route("/categories", func(r chi.Router) {
					r.Post("/", categoryHandler.Create)
					r.Get("/", categoryHandler.ListAll)
					r.Get("/{id}", categoryHandler.GetByID)
					r.Put("/{id}", categoryHandler.Update)
					r.Delete("/{id}", categoryHandler.Delete)
				})
				r.Route("/items", func(r chi.Router) {
					r.Post("/", productHandler.Create)
					r.Get("/", productHandler.List)
					r.Get("/{id}", productHandler.GetByID)
					r.Put("/{id}", productHandler.Update)
					r.Delete("/{id}", productHandler.Delete)
					r.Patch("/{id}/availability", productHandler.ToggleAvailability)
					r.Post("/{id}/modifier-groups", modifierHandler.AttachToProduct)
					r.Delete("/{id}/modifier-groups/{groupId}", modifierHandler.DetachFromProduct)
				})
			})

			r.Route("/admin/modifiers", func(r chi.Router) {
				r.Route("/groups", func(r chi.Router) {
					r.Post("/", modifierHandler.CreateGroup)
					r.Get("/", modifierHandler.ListGroups)
					r.Get("/{id}", modifierHandler.GetGroup)
					r.Put("/{id}", modifierHandler.UpdateGroup)
					r.Delete("/{id}", modifierHandler.DeleteGroup)
					r.Post("/{id}/options", modifierHandler.CreateOption)
					r.Get("/{id}/options", modifierHandler.ListOptions)
				})
				r.Route("/options", func(r chi.Router) {
					r.Put("/{optionId}", modifierHandler.UpdateOption)
					r.Delete("/{optionId}", modifierHandler.DeleteOption)
				})
			})
		})

		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(cfg.JWT.Secret))
			// r.Use(middleware.RequireRole(models.RoleAdmin, models.RoleManager))

			r.Route("/orders", func(r chi.Router) {
				r.Post("/", orderHandler.Create)
				r.Get("/", orderHandler.List)
				r.Get("/{id}", orderHandler.GetByID)
				r.Put("/{id}", orderHandler.Update)
				r.Delete("/{id}", orderHandler.Delete)
				r.Post("/{id}/checkout", orderHandler.Checkout)
				r.Post("/{id}/pay", orderHandler.Pay)
			})
		})
	})

	// 6. Build HTTP Server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 7. Start HTTP Server in goroutine
	srvErr := make(chan error, 1)
	go func() {
		log.Info("http server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			srvErr <- err
		}
	}()

	// 8. Wait for interruption or server error
	select {
	case err := <-srvErr:
		log.Error("http server failed to start or serve", "error", err)
	case <-ctx.Done():
		log.Info("shutting down HTTP server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("graceful shutdown failed, forcing close", "error", err)
			_ = srv.Close()
		} else {
			log.Info("HTTP server stopped cleanly")
		}
	}
}
