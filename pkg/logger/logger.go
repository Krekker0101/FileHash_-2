package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(env string) *zap.Logger {
	var cfg zap.Config
	switch env {
	case "production":
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	default:
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.TimeKey = "time"
	}
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	l, err := cfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return zap.NewNop()
	}
	_ = zap.ReplaceGlobals(l)
	return l
}

func NewFromConfig(cfg zap.Config) (*zap.Logger, error) {
	return cfg.Build(zap.AddStacktrace(zapcore.ErrorLevel))
}

func NewNop() *zap.Logger {
	return zap.NewNop()
}

func SetGlobal(logger *zap.Logger) {
	zap.ReplaceGlobals(logger)
}

func GetGlobal() *zap.Logger {
	return zap.L()
}
