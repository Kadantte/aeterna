package services

import (
	"strings"
	"testing"

	"github.com/alpyxn/aeterna/backend/internal/config/common"
)

func TestDatabaseModule_Metadata(t *testing.T) {
	m := DatabaseModule{}
	if got := m.Name(); got != "DatabaseModule" {
		t.Fatalf("Name() = %q, want %q", got, "DatabaseModule")
	}
	if got := m.Section(); got != "database" {
		t.Fatalf("Section() = %q, want %q", got, "database")
	}
}

func TestDatabaseModule_LoadAndValidate(t *testing.T) {
	t.Run("defaults in development mode", func(t *testing.T) {
		t.Setenv("ENV", "")
		t.Setenv("DATABASE_PATH", "")
		t.Setenv("DB_ENCRYPTION_ENABLED", "")
		t.Setenv("DB_ENCRYPTION_AUTO_MIGRATE", "")
		t.Setenv("DB_ENCRYPTION_KDF_CONTEXT_FILE", "")
		section, err := DatabaseModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.Path != common.DefaultDatabasePath {
			t.Fatalf("Path = %q, want %q", section.Path, common.DefaultDatabasePath)
		}
		if section.PathIsSet {
			t.Fatal("PathIsSet should be false when DATABASE_PATH is empty")
		}
		if section.EncryptionEnabled {
			t.Fatal("EncryptionEnabled should default to false")
		}
		if !section.EncryptionAutoMigrate {
			t.Fatal("EncryptionAutoMigrate should default to true")
		}
		if section.EncryptionKDFContextFile != common.DefaultDBEncryptionKDFContextFile {
			t.Fatalf("EncryptionKDFContextFile = %q, want %q", section.EncryptionKDFContextFile, common.DefaultDBEncryptionKDFContextFile)
		}
	})

	t.Run("custom DATABASE_PATH", func(t *testing.T) {
		t.Setenv("ENV", "")
		t.Setenv("DATABASE_PATH", "/data/custom.db")
		section, err := DatabaseModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.Path != "/data/custom.db" {
			t.Fatalf("Path = %q, want %q", section.Path, "/data/custom.db")
		}
		if !section.PathIsSet {
			t.Fatal("PathIsSet should be true when DATABASE_PATH is set")
		}
	})

	t.Run("production requires DATABASE_PATH", func(t *testing.T) {
		t.Setenv("ENV", "production")
		t.Setenv("DATABASE_PATH", "")
		_, err := DatabaseModule{}.LoadAndValidate()
		if err == nil {
			t.Fatal("expected error when DATABASE_PATH is unset in production")
		}
		if !strings.Contains(err.Error(), "DATABASE_PATH") {
			t.Fatalf("error message should mention DATABASE_PATH, got: %v", err)
		}
	})

	t.Run("production with DATABASE_PATH succeeds", func(t *testing.T) {
		t.Setenv("ENV", "production")
		t.Setenv("DATABASE_PATH", "/prod/aeterna.db")
		section, err := DatabaseModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.Path != "/prod/aeterna.db" {
			t.Fatalf("Path = %q, want %q", section.Path, "/prod/aeterna.db")
		}
		if !section.PathIsSet {
			t.Fatal("PathIsSet should be true")
		}
	})

	t.Run("postgres env vars are captured but not used as path", func(t *testing.T) {
		t.Setenv("ENV", "")
		t.Setenv("DATABASE_PATH", "")
		t.Setenv("DB_HOST", "pg.host")
		t.Setenv("POSTGRES_HOST", "pg2.host")
		t.Setenv("DATABASE_URL", "postgres://user:pass@host/db")
		section, err := DatabaseModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.DBHost != "pg.host" {
			t.Fatalf("DBHost = %q, want %q", section.DBHost, "pg.host")
		}
		if section.PostgresHost != "pg2.host" {
			t.Fatalf("PostgresHost = %q, want %q", section.PostgresHost, "pg2.host")
		}
		if section.DatabaseURL != "postgres://user:pass@host/db" {
			t.Fatalf("DatabaseURL = %q, want %q", section.DatabaseURL, "postgres://user:pass@host/db")
		}
		if section.Path != common.DefaultDatabasePath {
			t.Fatalf("Path should remain default, got %q", section.Path)
		}
	})

	t.Run("DATABASE_PATH whitespace is trimmed", func(t *testing.T) {
		t.Setenv("ENV", "")
		t.Setenv("DATABASE_PATH", "  /data/trimmed.db  ")
		section, err := DatabaseModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if section.Path != "/data/trimmed.db" {
			t.Fatalf("Path = %q, want trimmed %q", section.Path, "/data/trimmed.db")
		}
	})

	t.Run("database encryption env vars are parsed and kdf context path is custom", func(t *testing.T) {
		t.Setenv("ENV", "")
		t.Setenv("DATABASE_PATH", "")
		t.Setenv("DB_ENCRYPTION_ENABLED", "true")
		t.Setenv("DB_ENCRYPTION_AUTO_MIGRATE", "false")
		t.Setenv("DB_ENCRYPTION_KDF_CONTEXT_FILE", "/tmp/custom-kdf-context")

		section, err := DatabaseModule{}.LoadAndValidate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !section.EncryptionEnabled {
			t.Fatal("EncryptionEnabled should be true")
		}
		if section.EncryptionAutoMigrate {
			t.Fatal("EncryptionAutoMigrate should be false")
		}
		if section.EncryptionKDFContextFile != "/tmp/custom-kdf-context" {
			t.Fatalf("EncryptionKDFContextFile = %q, want %q", section.EncryptionKDFContextFile, "/tmp/custom-kdf-context")
		}
	})
}
