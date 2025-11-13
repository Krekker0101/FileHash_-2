package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/filehash/internal/config"
	"github.com/filehash/internal/infrastructure/database"
	infrahttp "github.com/filehash/internal/infrastructure/http"
	infrarepo "github.com/filehash/internal/infrastructure/repository"
	infraservice "github.com/filehash/internal/infrastructure/service"
	"github.com/filehash/internal/usecase"
	"github.com/filehash/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	cfg     config.Config
	log     *zap.Logger
	db      *gorm.DB
	router  http.Handler
	server  *http.Server
}

func New(cfg config.Config) (*App, error) {
	log := logger.New(cfg.Env)
	defer func() { _ = log.Sync() }()

	db, err := database.Open(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := database.Migrate(db, log); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	userRepo := infrarepo.NewUserRepository(db)
	fileRepo := infrarepo.NewFileRepository(db)
	excelRepo := infrarepo.NewExcelRepository(db)

	storageSvc, err := infraservice.NewStorageService(cfg.UploadsDir)
	if err != nil {
		return nil, fmt.Errorf("new storage service: %w", err)
	}

	tokenSvc, err := infraservice.NewTokenService(cfg.JWTSecret, cfg.TokenTTL)
	if err != nil {
		return nil, fmt.Errorf("new token service: %w", err)
	}

	authSvc, err := infraservice.NewAuthService(cfg.JWTSecret, cfg.TokenTTL)
	if err != nil {
		return nil, fmt.Errorf("new auth service: %w", err)
	}

	cryptoSvc := infraservice.NewCryptoService()

	authUseCase := usecase.NewAuthUseCase(userRepo, authSvc, log)
	fileUseCase := usecase.NewFileUseCase(fileRepo, storageSvc, cryptoSvc, tokenSvc, log)
	excelUseCase := usecase.NewExcelUseCase(excelRepo, storageSvc, log)

	handlers := infrahttp.NewHandlers(cfg, log, authUseCase, fileUseCase, excelUseCase)

	router := infrahttp.NewRouter(cfg, log, handlers)

	server := &http.Server{
		Addr:              cfg.HTTPAddr(),
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{
		cfg:    cfg,
		log:    log,
		db:     db,
		router: router,
		server: server,
	}, nil
}

func (a *App) Run() error {
	go func() {
		a.log.Sugar().Infow("server starting", "addr", a.cfg.HTTPAddr())
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Fatal("server error", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	a.log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	sqlDB, err := a.db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}

	a.log.Info("server stopped")
	return nil
}

func (a *App) Close() error {
	sqlDB, err := a.db.DB()
	if err == nil {
		return sqlDB.Close()
	}
	return nil
}

