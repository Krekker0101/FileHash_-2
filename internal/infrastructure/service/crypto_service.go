package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/filehash/internal/domain/service"
	"github.com/filehash/pkg/crypto"
)

type cryptoService struct{}

func NewCryptoService() service.CryptoService {
	return &cryptoService{}
}

var _ service.CryptoService = (*cryptoService)(nil)

func (c *cryptoService) GenerateAESKey() ([]byte, error) {
	key := make([]byte, crypto.AESKeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return key, nil
}

func (c *cryptoService) EncryptAESGCM(key, plaintext []byte) ([]byte, []byte, error) {
	if len(key) != crypto.AESKeySize {
		return nil, nil, errors.New("key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("new gcm: %w", err)
	}
	nonce := make([]byte, crypto.GCMNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("nonce: %w", err)
	}
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func (c *cryptoService) DecryptAESGCM(key, nonce, ciphertext []byte) ([]byte, error) {
	if len(key) != crypto.AESKeySize {
		return nil, errors.New("key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size: %d", len(nonce))
	}
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plaintext, nil
}

