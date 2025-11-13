package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/filehash/internal/domain/service"
	"github.com/filehash/pkg/crypto"
	"github.com/google/uuid"
)

const (
	encryptedDir = "encrypted"
	excelDir     = "excel"
)

type storageService struct {
	baseDir string
}

func NewStorageService(baseDir string) (service.StorageService, error) {
	if baseDir == "" {
		return nil, errors.New("baseDir is required")
	}
	return &storageService{baseDir: baseDir}, nil
}

var _ service.StorageService = (*storageService)(nil)

func (s *storageService) SaveEncrypted(ctx context.Context, originalName string, nonce, data []byte) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if len(nonce) == 0 {
		return "", errors.New("nonce cannot be empty")
	}
	if len(data) == 0 {
		return "", errors.New("ciphertext cannot be empty")
	}

	now := time.Now().UTC()
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext == "" {
		ext = ".bin"
	}

	dir := filepath.Join(s.baseDir, encryptedDir, now.Format("2006"), now.Format("01"), now.Format("02"))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}
	filename := fmt.Sprintf("%s%s.enc", uuid.NewString(), ext)
	fullPath := filepath.Join(dir, filename)

	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(nonce); err != nil {
		return "", fmt.Errorf("write nonce: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		return "", fmt.Errorf("write ciphertext: %w", err)
	}

	relPath, err := filepath.Rel(s.baseDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("relative path: %w", err)
	}
	return filepath.ToSlash(relPath), nil
}

func (s *storageService) LoadEncrypted(ctx context.Context, relativePath string) (nonce, data []byte, err error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	default:
	}

	fullPath, err := s.safeJoin(relativePath)
	if err != nil {
		return nil, nil, err
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, fmt.Errorf("open: %w", err)
	}
	defer file.Close()

	nonce = make([]byte, crypto.GCMNonceSize)
	if _, err := io.ReadFull(file, nonce); err != nil {
		return nil, nil, fmt.Errorf("read nonce: %w", err)
	}
	data, err = io.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("read ciphertext: %w", err)
	}
	return nonce, data, nil
}

func (s *storageService) SaveExcel(ctx context.Context, reader io.Reader) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	now := time.Now().UTC()
	dir := filepath.Join(s.baseDir, excelDir, now.Format("2006"), now.Format("01"), now.Format("02"))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}

	filename := fmt.Sprintf("excel_%d.xlsx", now.UnixNano())
	fullPath := filepath.Join(dir, filename)
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return "", fmt.Errorf("write excel: %w", err)
	}

	relPath, err := filepath.Rel(s.baseDir, fullPath)
	if err != nil {
		return "", fmt.Errorf("relative path: %w", err)
	}
	return filepath.ToSlash(relPath), nil
}

func (s *storageService) Delete(ctx context.Context, relativePath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fullPath, err := s.safeJoin(relativePath)
	if err != nil {
		return err
	}
	return os.Remove(fullPath)
}

func (s *storageService) safeJoin(relativePath string) (string, error) {
	if relativePath == "" {
		return "", errors.New("path cannot be empty")
	}
	clean := filepath.Clean(relativePath)
	full := filepath.Join(s.baseDir, clean)
	if !strings.HasPrefix(full, filepath.Clean(s.baseDir)+string(filepath.Separator)) {
		return "", errors.New("invalid path traversal attempt")
	}
	return full, nil
}

