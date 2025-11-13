package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/filehash/internal/domain/service"
	"github.com/filehash/pkg/crypto"
	"github.com/golang-jwt/jwt/v5"
)

type tokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewTokenService(secret string, ttl time.Duration) (service.TokenService, error) {
	if secret == "" {
		return nil, errors.New("jwt secret required")
	}
	if ttl <= 0 {
		return nil, errors.New("ttl must be positive")
	}
	return &tokenService{
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

var _ service.TokenService = (*tokenService)(nil)

type jwtClaims struct {
	service.FileTokenClaims
	jwt.RegisteredClaims
}

func (t *tokenService) Generate(fileID string, aesKey []byte, userID *string) (string, error) {
	if fileID == "" {
		return "", errors.New("fileID is required")
	}
	if len(aesKey) != crypto.AESKeySize {
		return "", errors.New("aes key must be 32 bytes")
	}

	now := time.Now().UTC()
	claims := jwtClaims{
		FileTokenClaims: service.FileTokenClaims{
			FileID: fileID,
			Key:    base64.StdEncoding.EncodeToString(aesKey),
			UserID: userID,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(t.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (t *tokenService) Validate(tokenStr string) (*service.FileTokenClaims, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	token, err := parser.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return t.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return &claims.FileTokenClaims, nil
}

