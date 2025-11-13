package repository

import (
	"context"
	"errors"

	"github.com/filehash/internal/domain/entity"
	"github.com/filehash/internal/domain/repository"
	"github.com/filehash/pkg/utils"
	"gorm.io/gorm"
)

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) repository.FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(ctx context.Context, asset *entity.FileAsset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

func (r *fileRepository) FindByID(ctx context.Context, id string) (*entity.FileAsset, error) {
	var asset entity.FileAsset
	if err := r.db.WithContext(ctx).First(&asset, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.ErrRecordNotFound
		}
		return nil, err
	}
	return &asset, nil
}

func (r *fileRepository) FindByUserID(ctx context.Context, userID string) ([]entity.FileAsset, error) {
	var assets []entity.FileAsset
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (r *fileRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entity.FileAsset{}, "id = ?", id).Error
}

