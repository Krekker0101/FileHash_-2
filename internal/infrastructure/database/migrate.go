package database

import (
	"fmt"

	"github.com/filehash/internal/domain/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, log *zap.Logger) error {
	if err := db.AutoMigrate(
		&entity.User{},
		&entity.FileAsset{},
		&entity.ExcelExport{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	log.Info("database migrations applied")
	return nil
}

