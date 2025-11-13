package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPHost     = "0.0.0.0"
	defaultHTTPPort     = "8080"
	defaultEnv          = "development"
	defaultUploadsDir   = "uploads"
	defaultDatabasePath = "data/filehash.db"
	defaultTokenTTL     = 15 * time.Minute
	defaultMaxUploadMB  = 10
	defaultDBType       = "sqlite"
)

type DBType string

const (
	DBTypeSQLite    DBType = "sqlite"
	DBTypePostgres  DBType = "postgres"
)

type Config struct {
	Env          string
	Host         string
	Port         string
	UploadsDir   string
	DatabaseType DBType
	DatabasePath string
	DatabaseDSN  string // Full DSN for PostgreSQL
	JWTSecret    string
	TokenTTL     time.Duration
	MaxUpload    int64
	CORSOrigins  []string
}

func (c Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func (c Config) GetDatabaseDSN() string {
	switch c.DatabaseType {
	case DBTypePostgres:
		if c.DatabaseDSN != "" {
			return c.DatabaseDSN
		}
		host := valueOrDefault("DB_HOST", "localhost")
		port := valueOrDefault("DB_PORT", "5432")
		user := valueOrDefault("DB_USER", "postgres")
		password := os.Getenv("DB_PASSWORD")
		dbname := valueOrDefault("DB_NAME", "filehash")
		sslmode := valueOrDefault("DB_SSLMODE", "disable")
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	case DBTypeSQLite:
		fallthrough
	default:
		return fmt.Sprintf("file:%s?_foreign_keys=on&_journal_mode=WAL", c.DatabasePath)
	}
}

func Load() (Config, error) {
	cfg := Config{
		Env:          valueOrDefault("APP_ENV", defaultEnv),
		Host:         valueOrDefault("HTTP_HOST", defaultHTTPHost),
		Port:         valueOrDefault("HTTP_PORT", defaultHTTPPort),
		UploadsDir:   valueOrDefault("UPLOADS_DIR", defaultUploadsDir),
		DatabaseType: DBType(valueOrDefault("DB_TYPE", defaultDBType)),
		DatabasePath: valueOrDefault("DATABASE_PATH", defaultDatabasePath),
		DatabaseDSN:  os.Getenv("DATABASE_URL"), // Full DSN for PostgreSQL
		JWTSecret:    os.Getenv("JWT_SECRET"),
		TokenTTL:     defaultTokenTTL,
		MaxUpload:    defaultMaxUploadMB * 1024 * 1024,
	}

	if cfg.DatabaseType != DBTypeSQLite && cfg.DatabaseType != DBTypePostgres {
		return Config{}, fmt.Errorf("invalid DB_TYPE: %s (must be 'sqlite' or 'postgres')", cfg.DatabaseType)
	}

	if ttlStr := os.Getenv("JWT_TTL_MINUTES"); ttlStr != "" {
		ttlMinutes, err := strconv.Atoi(ttlStr)
		if err != nil || ttlMinutes <= 0 {
			return Config{}, fmt.Errorf("invalid JWT_TTL_MINUTES value: %q", ttlStr)
		}
		cfg.TokenTTL = time.Duration(ttlMinutes) * time.Minute
	}

	if maxStr := os.Getenv("MAX_UPLOAD_MB"); maxStr != "" {
		maxMB, err := strconv.Atoi(maxStr)
		if err != nil || maxMB <= 0 {
			return Config{}, fmt.Errorf("invalid MAX_UPLOAD_MB value: %q", maxStr)
		}
		cfg.MaxUpload = int64(maxMB) * 1024 * 1024
	}

	if cfg.JWTSecret == "" {
		secret, err := randomSecret(32)
		if err != nil {
			return Config{}, fmt.Errorf("generate jwt secret: %w", err)
		}
		cfg.JWTSecret = secret
	}

	// For SQLite, ensure parent directory exists
	if cfg.DatabaseType == DBTypeSQLite {
		if err := ensureParentDir(cfg.DatabasePath); err != nil {
			return Config{}, fmt.Errorf("prepare database path: %w", err)
		}
	}

	if err := os.MkdirAll(cfg.UploadsDir, 0o750); err != nil {
		return Config{}, fmt.Errorf("create uploads dir: %w", err)
	}

	if corsEnv := os.Getenv("CORS_ORIGINS"); corsEnv != "" {
		cfg.CORSOrigins = strings.Split(corsEnv, ",")
		for i := range cfg.CORSOrigins {
			cfg.CORSOrigins[i] = strings.TrimSpace(cfg.CORSOrigins[i])
		}
	}

	return cfg, nil
}

func valueOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func randomSecret(size int) (string, error) {
	if size <= 0 {
		return "", errors.New("size must be positive")
	}
	buff := make([]byte, size)
	if _, err := rand.Read(buff); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buff), nil
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o750)
}
