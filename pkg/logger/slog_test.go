package logger_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/iden3/go-service-template/pkg/logger"
	"github.com/pkg/errors"
)

func TestLogDevelopment(t *testing.T) {
	logger.SetDefaultLogger(logger.EnvDevelopment, slog.LevelDebug)
	logger.Debug("debug", slog.String("x", "1"))
	logger.Info("info", slog.String("x", "2"))
	logger.Warn("warn", slog.String("x", "3"))
	logger.Error("error", slog.String("x", "4"))

	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "123")
	logger.WithContext(ctx).Debug("debug", slog.String("y", "1"))
	logger.WithContext(ctx).Info("debug", slog.String("y", "2"))
	logger.WithContext(ctx).Warn("debug", slog.String("y", "3"))
	logger.WithContext(ctx).Error("debug", slog.String("y", "4"))
}

func TestLogProduction(t *testing.T) {
	logger.SetDefaultLogger(logger.EnvProduction, slog.LevelDebug)
	logger.Debug("debug", slog.String("x", "1"))
	logger.Info("info", slog.String("x", "2"))
	logger.Warn("warn", slog.String("x", "3"))
	logger.Error("error", slog.String("x", "4"))
}

func TestWithParamsContextError(t *testing.T) {
	logger.SetDefaultLogger(logger.EnvDevelopment, slog.LevelDebug)
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "123")
	logger.WithContext(ctx).WithError(errors.New("common error1")).Debug("debug", slog.String("y", "1"))
	logger.WithContext(ctx).WithError(errors.New("common error2")).Info("debug", slog.String("y", "2"))
	logger.WithContext(ctx).WithError(errors.New("common error3")).Warn("debug", slog.String("y", "3"))
	logger.WithContext(ctx).WithError(errors.New("common error4")).Error("debug", slog.String("y", "4"))
}

func TestWithParamsErrorContext(t *testing.T) {
	logger.SetDefaultLogger(logger.EnvDevelopment, slog.LevelDebug)
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "123")
	logger.WithError(errors.New("common error1")).WithContext(ctx).Debug("debug", slog.String("y", "1"))
	logger.WithError(errors.New("common error2")).WithContext(ctx).Info("debug", slog.String("y", "2"))
	logger.WithError(errors.New("common error3")).WithContext(ctx).Warn("debug", slog.String("y", "3"))
	logger.WithError(errors.New("common error4")).WithContext(ctx).Error("debug", slog.String("y", "4"))
}
