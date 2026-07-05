package config_test

import (
	"os"
	"testing"

	"github.com/D3rille/kk-fast-food-system/internal/config"
)

func TestConfigLoadDefaults(t *testing.T) {
	// Ensure environment is clean of test variables
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("SERVER_PORT")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected Host '0.0.0.0', got %q", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected Port 8080, got %d", cfg.Server.Port)
	}

	if cfg.App.Env != "development" {
		t.Errorf("expected App Env 'development', got %q", cfg.App.Env)
	}
}

func TestConfigLoadEnvOverride(t *testing.T) {
	os.Setenv("SERVER_PORT", "9090")
	defer os.Unsetenv("SERVER_PORT")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected Port 9090, got %d", cfg.Server.Port)
	}
}
