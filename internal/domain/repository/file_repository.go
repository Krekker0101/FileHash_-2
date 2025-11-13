package repository

import (
	"context"

	"github.com/filehash/internal/domain/entity"
)

type FileRepository interface {
	Create(ctx context.Context, asset *entity.FileAsset) error
	FindByID(ctx context.Context, id string) (*entity.FileAsset, error)
	FindByUserID(ctx context.Context, userID string) ([]entity.FileAsset, error)
	Delete(ctx context.Context, id string) error
}

