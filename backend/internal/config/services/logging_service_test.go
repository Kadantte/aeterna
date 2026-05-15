package services

import (
	"testing"

	"github.com/alpyxn/aeterna/backend/internal/config/common"
)

func TestLoggingModule_Metadata(t *testing.T) {
	m := LoggingModule{}
	if got := m.Name(); got != "LoggingModule" {
		t.Fatalf("Name() = %q, want %q", got, "LoggingModule")
	}
	if got := m.Section(); got != "logging" {
		t.Fatalf("Section() = %q, want %q", got, "logging")
	}
}

func TestLoggingModule_LoadAndValidate(t *testing.T) {
	t.Run("defaults when no env vars set", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "")
		t.Setenv("LOG_FORMAT", "")
		t.Setenv("LOG_FILE", "")
		t.Setenv("LOG_MAX_SIZE", "")
		t.Setenv("LOG_MAX_BACKUPS", "")
		t.Setenv("LOG_MAX_AGE", "")
		t.Setenv("LOG_COMPRESS", "")
		section, err := LoggingModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.Level != "" {
			t.Fatalf("Level = %q, want empty", section.Level)
		}
		if section.Format != "" {
			t.Fatalf("Format = %q, want empty", section.Format)
		}
		if section.File != "" {
			t.Fatalf("File = %q, want empty", section.File)
		}
		if section.MaxSize != common.DefaultLogMaxSize {
			t.Fatalf("MaxSize = %d, want %d", section.MaxSize, common.DefaultLogMaxSize)
		}
		if section.MaxBackups != common.DefaultLogMaxBackups {
			t.Fatalf("MaxBackups = %d, want %d", section.MaxBackups, common.DefaultLogMaxBackups)
		}
		if section.MaxAge != common.DefaultLogMaxAge {
			t.Fatalf("MaxAge = %d, want %d", section.MaxAge, common.DefaultLogMaxAge)
		}
		if section.Compress != common.DefaultLogCompress {
			t.Fatalf("Compress = %v, want %v", section.Compress, common.DefaultLogCompress)
		}
	})

	t.Run("custom log level and format", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("LOG_FORMAT", "json")
		section, err := LoggingModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.Level != "debug" {
			t.Fatalf("Level = %q, want %q", section.Level, "debug")
		}
		if section.Format != "json" {
			t.Fatalf("Format = %q, want %q", section.Format, "json")
		}
	})

	t.Run("custom log file path", func(t *testing.T) {
		t.Setenv("LOG_FILE", "/var/log/aeterna.log")
		section, err := LoggingModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.File != "/var/log/aeterna.log" {
			t.Fatalf("File = %q, want %q", section.File, "/var/log/aeterna.log")
		}
	})

	t.Run("custom rotation settings", func(t *testing.T) {
		t.Setenv("LOG_MAX_SIZE", "100")
		t.Setenv("LOG_MAX_BACKUPS", "10")
		t.Setenv("LOG_MAX_AGE", "30")
		section, err := LoggingModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.MaxSize != 100 {
			t.Fatalf("MaxSize = %d, want 100", section.MaxSize)
		}
		if section.MaxBackups != 10 {
			t.Fatalf("MaxBackups = %d, want 10", section.MaxBackups)
		}
		if section.MaxAge != 30 {
			t.Fatalf("MaxAge = %d, want 30", section.MaxAge)
		}
	})

	t.Run("invalid int values fall back to defaults", func(t *testing.T) {
		t.Setenv("LOG_MAX_SIZE", "not-a-number")
		t.Setenv("LOG_MAX_BACKUPS", "bad")
		t.Setenv("LOG_MAX_AGE", "abc")
		section, err := LoggingModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.MaxSize != common.DefaultLogMaxSize {
			t.Fatalf("MaxSize = %d, want default %d", section.MaxSize, common.DefaultLogMaxSize)
		}
		if section.MaxBackups != common.DefaultLogMaxBackups {
			t.Fatalf("MaxBackups = %d, want default %d", section.MaxBackups, common.DefaultLogMaxBackups)
		}
		if section.MaxAge != common.DefaultLogMaxAge {
			t.Fatalf("MaxAge = %d, want default %d", section.MaxAge, common.DefaultLogMaxAge)
		}
	})

	t.Run("LOG_COMPRESS false disables compression", func(t *testing.T) {
		t.Setenv("LOG_COMPRESS", "false")
		section, err := LoggingModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.Compress {
			t.Fatal("Compress should be false")
		}
	})

	t.Run("LOG_COMPRESS true enables compression", func(t *testing.T) {
		t.Setenv("LOG_COMPRESS", "true")
		section, err := LoggingModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !section.Compress {
			t.Fatal("Compress should be true")
		}
	})
}
