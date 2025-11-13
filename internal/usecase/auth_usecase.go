package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/filehash/internal/domain/entity"
	"github.com/filehash/internal/domain/repository"
	"github.com/filehash/internal/domain/service"
	"github.com/filehash/pkg/utils"
	"go.uber.org/zap"
)

type AuthUseCase struct {
	userRepo  repository.UserRepository
	authSvc   service.AuthService
	log       *zap.Logger
}

func NewAuthUseCase(
	userRepo repository.UserRepository,
	authSvc service.AuthService,
	log *zap.Logger,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo: userRepo,
		authSvc:  authSvc,
		log:      log,
	}
}

type RegisterRequest struct {
	Email    string
	Password string
}

type RegisterResponse struct {
	UserID string
	Token  string
}

func (uc *AuthUseCase) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return nil, fmt.Errorf("invalid email format")
	}
	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}

	_, err := uc.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return nil, fmt.Errorf("email already exists")
	}
	if err != utils.ErrRecordNotFound {
		return nil, fmt.Errorf("check email: %w", err)
	}

	hashedPassword, err := uc.authSvc.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &entity.User{
		Email:    email,
		Password: hashedPassword,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	token, err := uc.authSvc.GenerateAuthToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &RegisterResponse{
		UserID: user.ID,
		Token:  token,
	}, nil
}

type LoginRequest struct {
	Email    string
	Password string
}

type LoginResponse struct {
	UserID string
	Token  string
}

func (uc *AuthUseCase) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return nil, fmt.Errorf("invalid email format")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if err == utils.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if err := uc.authSvc.ComparePassword(user.Password, req.Password); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := uc.authSvc.GenerateAuthToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResponse{
		UserID: user.ID,
		Token:  token,
	}, nil
}

