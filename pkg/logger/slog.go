package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

const (
	EnvDevelopment = "development"
	EnvProduction  = "production"

	LevelNotice = slog.Level(2)
	LevelFatal  = slog.Level(12)
)

type logger struct {
	provider *slog.Logger
	context  context.Context
}

func (l *logger) Debug(msg string, attrs ...slog.Attr) {
	l.provider.LogAttrs(l.context, slog.LevelDebug, msg, attrs...)
}

func (l *logger) Info(msg string, attrs ...slog.Attr) {
	l.provider.LogAttrs(l.context, slog.LevelInfo, msg, attrs...)
}

func (l *logger) Notice(msg string, attrs ...slog.Attr) {
	l.provider.LogAttrs(l.context, LevelNotice, msg, attrs...)
}

func (l *logger) Warn(msg string, attrs ...slog.Attr) {
	l.provider.LogAttrs(l.context, slog.LevelWarn, msg, attrs...)
}

func (l *logger) Error(msg string, attrs ...slog.Attr) {
	l.provider.LogAttrs(l.context, slog.LevelError, msg, attrs...)
}

func (l *logger) Fatal(msg string, attrs ...slog.Attr) {
	l.provider.LogAttrs(l.context, LevelFatal, msg, attrs...)
	os.Exit(1)
}

func (l *logger) WithContext(ctx context.Context) *logger {
	reqID := middleware.GetReqID(ctx)
	if reqID == "" {
		return l
	}
	childlog := l.provider.With(
		slog.String("request_id", reqID),
	)
	return &logger{childlog, ctx}
}

func (l *logger) WithError(err error) *logger {
	var stackArr []string
	if err, ok := err.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			stackArr = append(stackArr, fmt.Sprintf("%+s:%d", f, f))
		}
	}

	parentlog := l.provider
	if len(stackArr) > 0 {
		// stacktrace as set of strings works only with json format
		parentlog = l.provider.With(slog.Any("stacktrace", stackArr))
	}

	childlog := parentlog.With(
		slog.String("error", err.Error()),
	)
	return &logger{childlog, l.context}
}

var defaultLogger *logger

func mustGetDefaultLogger() *logger {
	if defaultLogger == nil {
		panic("default logger is not set")
	}
	return defaultLogger
}

func SetDefaultLogger(env string, loglevel slog.Level) error {
	var (
		providerOpts *slog.HandlerOptions
		handler      slog.Handler
	)

	switch loglevel {
	case slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
		LevelFatal,
		LevelNotice:
	default:
		return errors.Errorf("unknown log level '%s'", loglevel)
	}

	providerOpts = &slog.HandlerOptions{
		Level: loglevel,
	}
	switch env {
	case EnvDevelopment:
		handler = slog.NewTextHandler(os.Stdout, providerOpts)
	case EnvProduction:
		handler = slog.NewJSONHandler(os.Stdout, providerOpts)
	default:
		return errors.Errorf("unknown log environment '%s'", env)
	}

	logProvider := slog.New(handler)
	slog.SetDefault(logProvider) // we should have common logformat if somebody call log/slog directly
	defaultLogger = &logger{logProvider, context.Background()}
	return nil
}

func WithContext(ctx context.Context) *logger {
	return mustGetDefaultLogger().WithContext(ctx)
}

func WithError(err error) *logger {
	return mustGetDefaultLogger().WithError(err)
}

func Debug(msg string, attrs ...slog.Attr) {
	mustGetDefaultLogger().Debug(msg, attrs...)
}

func Info(msg string, attrs ...slog.Attr) {
	mustGetDefaultLogger().Info(msg, attrs...)
}

func Notice(msg string, attrs ...slog.Attr) {
	mustGetDefaultLogger().Notice(msg, attrs...)
}

func Warn(msg string, attrs ...slog.Attr) {
	mustGetDefaultLogger().Warn(msg, attrs...)
}

func Error(msg string, attrs ...slog.Attr) {
	mustGetDefaultLogger().Error(msg, attrs...)
}

func Fatal(msg string, attrs ...slog.Attr) {
	mustGetDefaultLogger().Fatal(msg, attrs...)
}
