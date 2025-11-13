package usecase

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/filehash/internal/domain/entity"
	"github.com/filehash/internal/domain/repository"
	"github.com/filehash/internal/domain/service"
	"github.com/filehash/pkg/utils"
	"go.uber.org/zap"
)

type FileUseCase struct {
	fileRepo   repository.FileRepository
	storageSvc service.StorageService
	cryptoSvc  service.CryptoService
	tokenSvc   service.TokenService
	log        *zap.Logger
}

func NewFileUseCase(
	fileRepo repository.FileRepository,
	storageSvc service.StorageService,
	cryptoSvc service.CryptoService,
	tokenSvc service.TokenService,
	log *zap.Logger,
) *FileUseCase {
	return &FileUseCase{
		fileRepo:   fileRepo,
		storageSvc: storageSvc,
		cryptoSvc:  cryptoSvc,
		tokenSvc:   tokenSvc,
		log:        log,
	}
}

type UploadFileRequest struct {
	Filename    string
	Content     []byte
	ContentType string
	UserID      *string
}

type UploadFileResponse struct {
	FileID    string
	Token     string
	ExpiresIn int
}

func (uc *FileUseCase) UploadFile(ctx context.Context, req UploadFileRequest) (*UploadFileResponse, error) {
	aesKey, err := uc.cryptoSvc.GenerateAESKey()
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	ciphertext, nonce, err := uc.cryptoSvc.EncryptAESGCM(aesKey, req.Content)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	storagePath, err := uc.storageSvc.SaveEncrypted(ctx, req.Filename, nonce, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("save encrypted: %w", err)
	}

	asset := &entity.FileAsset{
		OriginalName:      req.Filename,
		StoredPath:        storagePath,
		UserID:            req.UserID,
		ContentType:       req.ContentType,
		SizeBytes:         int64(len(req.Content)),
		EncryptionAlg:     "AES-256",
		AuthenticationAlg: "GCM",
	}

	if err := uc.fileRepo.Create(ctx, asset); err != nil {
		_ = uc.storageSvc.Delete(ctx, storagePath)
		return nil, fmt.Errorf("create record: %w", err)
	}

	token, err := uc.tokenSvc.Generate(asset.ID, aesKey, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &UploadFileResponse{
		FileID:    asset.ID,
		Token:     token,
		ExpiresIn: 900, // 15 minutes in seconds
	}, nil
}

type GetFileRequest struct {
	FileID string
	Token  string
}

type GetFileResponse struct {
	Content     []byte
	ContentType string
	Filename    string
}

func (uc *FileUseCase) GetFile(ctx context.Context, req GetFileRequest) (*GetFileResponse, error) {
	claims, err := uc.tokenSvc.Validate(req.Token)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}

	if claims.FileID != req.FileID {
		return nil, fmt.Errorf("token file_id mismatch")
	}

	asset, err := uc.fileRepo.FindByID(ctx, req.FileID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("find file: %w", err)
	}

	key, err := decodeKey(claims.Key)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}

	nonce, ciphertext, err := uc.storageSvc.LoadEncrypted(ctx, asset.StoredPath)
	if err != nil {
		return nil, fmt.Errorf("load encrypted: %w", err)
	}

	plaintext, err := uc.cryptoSvc.DecryptAESGCM(key, nonce, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}

	return &GetFileResponse{
		Content:     plaintext,
		ContentType: asset.ContentType,
		Filename:    asset.OriginalName,
	}, nil
}

type GetFileMetadataRequest struct {
	FileID string
	Token  string
}

func (uc *FileUseCase) GetFileMetadata(ctx context.Context, req GetFileMetadataRequest) (*entity.FileAsset, error) {
	claims, err := uc.tokenSvc.Validate(req.Token)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}

	if claims.FileID != req.FileID {
		return nil, fmt.Errorf("token file_id mismatch")
	}

	asset, err := uc.fileRepo.FindByID(ctx, req.FileID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found")
		}
		return nil, fmt.Errorf("find file: %w", err)
	}

	return asset, nil
}

type DeleteFileRequest struct {
	FileID string
	Token  string
}

func (uc *FileUseCase) DeleteFile(ctx context.Context, req DeleteFileRequest) error {
	claims, err := uc.tokenSvc.Validate(req.Token)
	if err != nil {
		return fmt.Errorf("validate token: %w", err)
	}

	if claims.FileID != req.FileID {
		return fmt.Errorf("token file_id mismatch")
	}

	asset, err := uc.fileRepo.FindByID(ctx, req.FileID)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return fmt.Errorf("file not found")
		}
		return fmt.Errorf("find file: %w", err)
	}

	if err := uc.storageSvc.Delete(ctx, asset.StoredPath); err != nil {
		uc.log.Warn("storage delete failed", zap.Error(err))
	}

	if err := uc.fileRepo.Delete(ctx, req.FileID); err != nil {
		return fmt.Errorf("delete record: %w", err)
	}

	return nil
}

type ListFilesRequest struct {
	UserID string
}

func (uc *FileUseCase) ListFiles(ctx context.Context, req ListFilesRequest) ([]entity.FileAsset, error) {
	assets, err := uc.fileRepo.FindByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("find files: %w", err)
	}
	return assets, nil
}

func decodeKey(keyStr string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("decode key: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key size")
	}
	return key, nil
}
