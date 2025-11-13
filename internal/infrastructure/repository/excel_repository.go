package repository

import (
	"context"
	"errors"

	"github.com/filehash/internal/domain/entity"
	"github.com/filehash/internal/domain/repository"
	"github.com/filehash/pkg/utils"
	"gorm.io/gorm"
)

type excelRepository struct {
	db *gorm.DB
}

func NewExcelRepository(db *gorm.DB) repository.ExcelRepository {
	return &excelRepository{db: db}
}

func (r *excelRepository) Create(ctx context.Context, export *entity.ExcelExport) error {
	return r.db.WithContext(ctx).Create(export).Error
}

func (r *excelRepository) FindByID(ctx context.Context, id string) (*entity.ExcelExport, error) {
	var export entity.ExcelExport
	if err := r.db.WithContext(ctx).First(&export, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}
	return &export, nil
}

func (r *excelRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.ExcelExport{}, "id = ?", id).Error
}

