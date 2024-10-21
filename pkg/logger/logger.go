package logger

import (
	"context"
	"log/slog"
	"os"
)

type ContextKey string

var (
	logger *slog.Logger
)

const (
	CTXKey ContextKey = "logger"
)

func Get() *slog.Logger {
	if logger == nil {
		Init(false, false)
	}

	return logger
}

func Init(json bool, verbose bool) *slog.Logger {
	opts := &slog.HandlerOptions{}
	if verbose {
		opts.Level = slog.LevelDebug
	}

	logger = slog.New(slog.NewTextHandler(os.Stdout, opts))

	if json {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	return logger
}

func IntoContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, CTXKey, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(CTXKey).(*slog.Logger)
	if !ok {
		return Get()
	}

	return l
}
