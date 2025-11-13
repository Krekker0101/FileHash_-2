package database

import (
	"fmt"
	"time"

	"github.com/filehash/internal/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(cfg config.Config, log *zap.Logger) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		FullSaveAssociations: false,
		NowFunc:              time.Now().UTC,
	}

	if cfg.Env == "production" {
		gormCfg.Logger = logger.Default.LogMode(logger.Silent)
	} else {
		gormCfg.Logger = logger.Default.LogMode(logger.Info)
	}

	var db *gorm.DB
	var err error

	switch cfg.DatabaseType {
	case config.DBTypePostgres:
		dsn := cfg.GetDatabaseDSN()
		db, err = gorm.Open(postgres.Open(dsn), gormCfg)
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}
		log.Info("connected to PostgreSQL database")
	case config.DBTypeSQLite:
		fallthrough
	default:
		dsn := cfg.GetDatabaseDSN()
		db, err = gorm.Open(sqlite.Open(dsn), gormCfg)
		if err != nil {
			return nil, fmt.Errorf("open sqlite: %w", err)
		}
		log.Info("connected to SQLite database", zap.String("path", cfg.DatabasePath))
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db handle: %w", err)
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxIdleTime(15 * time.Minute)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

