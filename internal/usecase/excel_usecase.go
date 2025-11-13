package usecase

import (
	"bytes"
	"context"
	"fmt"

	"github.com/filehash/internal/domain/entity"
	"github.com/filehash/internal/domain/repository"
	"github.com/filehash/internal/domain/service"
	excelService "github.com/filehash/internal/infrastructure/service"
	"go.uber.org/zap"
)

type ExcelUseCase struct {
	excelRepo  repository.ExcelRepository
	storageSvc service.StorageService
	log        *zap.Logger
}

func NewExcelUseCase(
	excelRepo repository.ExcelRepository,
	storageSvc service.StorageService,
	log *zap.Logger,
) *ExcelUseCase {
	return &ExcelUseCase{
		excelRepo:  excelRepo,
		storageSvc: storageSvc,
		log:        log,
	}
}

type GenerateExcelRequest struct {
	Data map[string][]any
}

type GenerateExcelResponse struct {
	ExcelID string
	Path    string
	Rows    int
}

func (uc *ExcelUseCase) GenerateExcel(ctx context.Context, req GenerateExcelRequest) (*GenerateExcelResponse, error) {
	buf, rows, err := excelService.GenerateExcel(req.Data)
	if err != nil {
		return nil, fmt.Errorf("generate excel: %w", err)
	}

	path, err := uc.storageSvc.SaveExcel(ctx, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("save excel: %w", err)
	}

	export := &entity.ExcelExport{
		StoredPath: path,
	}

	if err := uc.excelRepo.Create(ctx, export); err != nil {
		_ = uc.storageSvc.Delete(ctx, path)
		return nil, fmt.Errorf("create record: %w", err)
	}

	return &GenerateExcelResponse{
		ExcelID: export.ID,
		Path:    path,
		Rows:    rows,
	}, nil
}
