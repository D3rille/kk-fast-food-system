package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Storage  StorageConfig  `mapstructure:"storage"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Server   ServerConfig   `mapstructure:"server"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	URL      string `mapstructure:"url"`
	MaxConns int32  `mapstructure:"max_conns"`
	MinConns int32  `mapstructure:"min_conns"`
}

type AppConfig struct {
	Env      string `mapstructure:"env"`
	LogLevel string `mapstructure:"log_level"`
}

type JWTConfig struct {
	Secret            string        `mapstructure:"secret"`
	Expiration        time.Duration `mapstructure:"expiration"`
	RefreshExpiration time.Duration `mapstructure:"refresh_expiration"`
}

type StorageConfig struct {
	ImagesDir       string `mapstructure:"images_dir"`
	MaxUploadSizeMB int64  `mapstructure:"max_upload_size_mb"`
}

// Load loads config from environment variables and optionally configs/.env file.
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "5s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.idle_timeout", "30s")
	v.SetDefault("database.max_conns", int32(25))
	v.SetDefault("database.min_conns", int32(5))
	v.SetDefault("app.env", "development")
	v.SetDefault("app.log_level", "debug")
	v.SetDefault("jwt.secret", "default-jwt-secret-key-change-me-in-production")
	v.SetDefault("jwt.expiration", "15m")
	v.SetDefault("jwt.refresh_expiration", "168h") // 7 days
	v.SetDefault("storage.images_dir", "files/images")
	v.SetDefault("storage.max_upload_size_mb", int64(5))

	// Bind Environment Variables
	_ = v.BindEnv("server.host", "SERVER_HOST")
	_ = v.BindEnv("server.port", "SERVER_PORT")
	_ = v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	_ = v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	_ = v.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")
	_ = v.BindEnv("database.url", "DATABASE_URL")
	_ = v.BindEnv("database.max_conns", "DATABASE_MAX_CONNS")
	_ = v.BindEnv("database.min_conns", "DATABASE_MIN_CONNS")
	_ = v.BindEnv("app.env", "APP_ENV")
	_ = v.BindEnv("app.log_level", "APP_LOG_LEVEL")
	_ = v.BindEnv("jwt.secret", "JWT_SECRET")
	_ = v.BindEnv("jwt.expiration", "JWT_EXPIRATION")
	_ = v.BindEnv("jwt.refresh_expiration", "JWT_REFRESH_EXPIRATION")
	_ = v.BindEnv("storage.images_dir", "STORAGE_IMAGES_DIR")
	_ = v.BindEnv("storage.max_upload_size_mb", "STORAGE_MAX_UPLOAD_SIZE_MB")

	// Read from config file configs/.env if exists
	v.AddConfigPath("./configs")
	v.SetConfigName(".env")
	v.SetConfigType("env")

	if err := v.ReadInConfig(); err != nil {
		// If config file is not found, we don't return error because environment variables are fine.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &cfg, nil
}
