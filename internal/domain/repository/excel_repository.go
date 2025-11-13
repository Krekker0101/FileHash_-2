package repository

import (
	"context"

	"github.com/filehash/internal/domain/entity"
)

type ExcelRepository interface {
	Create(ctx context.Context, export *entity.ExcelExport) error
	FindByID(ctx context.Context, id string) (*entity.ExcelExport, error)
	Delete(ctx context.Context, id string) error
}

