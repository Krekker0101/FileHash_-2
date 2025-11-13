package service

import (
	"context"
	"io"
)

type StorageService interface {
	SaveEncrypted(ctx context.Context, originalName string, nonce, data []byte) (string, error)
	LoadEncrypted(ctx context.Context, relativePath string) (nonce, data []byte, err error)
	SaveExcel(ctx context.Context, reader io.Reader) (string, error)
	Delete(ctx context.Context, relativePath string) error
}

