package logging

import (
	"log/slog"
	"os"
	"strings"

	"github.com/alpyxn/aeterna/backend/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Init(cfg config.Config) *slog.Logger {
	level := parseLevel(cfg.Logging.Level)
	format := strings.ToLower(cfg.Logging.Format)

	var output *lumberjack.Logger
	if cfg.Logging.File != "" {
		output = &lumberjack.Logger{
			Filename:   cfg.Logging.File,
			MaxSize:    cfg.Logging.MaxSize,
			MaxBackups: cfg.Logging.MaxBackups,
			MaxAge:     cfg.Logging.MaxAge,
			Compress:   cfg.Logging.Compress,
		}
	}

	handlerOpts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if format == "text" {
		if output != nil {
			handler = slog.NewTextHandler(output, handlerOpts)
		} else {
			handler = slog.NewTextHandler(os.Stdout, handlerOpts)
		}
	} else {
		if output != nil {
			handler = slog.NewJSONHandler(output, handlerOpts)
		} else {
			handler = slog.NewJSONHandler(os.Stdout, handlerOpts)
		}
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func parseLevel(value string) slog.Level {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
