package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/filehash/internal/domain/service"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	secret []byte
	ttl    time.Duration
}

func NewAuthService(secret string, ttl time.Duration) (service.AuthService, error) {
	if secret == "" {
		return nil, errors.New("jwt secret required")
	}
	if ttl <= 0 {
		return nil, errors.New("ttl must be positive")
	}
	return &authService{
		secret: []byte(secret),
		ttl:    ttl,
	}, nil
}

var _ service.AuthService = (*authService)(nil)

func (a *authService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func (a *authService) ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

type authClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func (a *authService) GenerateAuthToken(userID string) (string, error) {
	if userID == "" {
		return "", errors.New("userID is required")
	}

	now := time.Now().UTC()
	claims := authClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(a.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (a *authService) ValidateAuthToken(tokenStr string) (string, error) {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	token, err := parser.ParseWithClaims(tokenStr, &authClaims{}, func(token *jwt.Token) (interface{}, error) {
		return a.secret, nil
	})
	if err != nil {
		return "", fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*authClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}
	return claims.UserID, nil
}

