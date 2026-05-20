package database

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alpyxn/aeterna/backend/internal/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

type SQLiteEncryptionConfig struct {
	Enabled     bool
	AutoMigrate bool
	Passphrase  string
}

func Connect(cfg config.Config, enc SQLiteEncryptionConfig) {
	dbPath := cfg.Database.Path

	// Warn if PostgreSQL environment variables are set (Aeterna uses SQLite only)
	if cfg.Database.DBHost != "" || cfg.Database.PostgresHost != "" || cfg.Database.DatabaseURL != "" {
		log.Println("WARNING: PostgreSQL environment variables detected, but Aeterna uses SQLite only.")
		log.Println("Ignoring PostgreSQL configuration and using SQLite at:", dbPath)
	}

	// Create data directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatal("Failed to create database directory: ", err)
		}
	}

	var err error
	if enc.Enabled {
		DB, err = connectEncrypted(dbPath, enc)
	} else {
		DB, err = connectPlain(dbPath, enc)
	}
	if err != nil {
		log.Fatal(err)
	}

	mode := "plain"
	if enc.Enabled {
		mode = "encrypted"
	}
	log.Printf("Database connection successfully opened (%s): %s", mode, dbPath)
}

func connectEncrypted(dbPath string, enc SQLiteEncryptionConfig) (*gorm.DB, error) {
	if strings.TrimSpace(enc.Passphrase) == "" {
		return nil, fmt.Errorf("failed to connect to encrypted SQLite database at %s: empty passphrase", dbPath)
	}

	db, err := openSQLite(buildSQLiteDSN(dbPath, enc.Passphrase))
	if err == nil {
		return db, nil
	}

	if !enc.AutoMigrate {
		return nil, fmt.Errorf("failed to open encrypted SQLite database at %s and DB_ENCRYPTION_AUTO_MIGRATE=false: %w", dbPath, err)
	}

	plainProbe, plainErr := openSQLite(buildSQLiteDSN(dbPath, ""))
	if plainErr != nil {
		return nil, fmt.Errorf("failed to open SQLite database at %s as encrypted (%v) and plain (%v)", dbPath, err, plainErr)
	}
	_ = closeSQLite(plainProbe)

	backupPath, err := migratePlainToEncrypted(dbPath, enc.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate SQLite database to encrypted mode: %w", err)
	}

	db, err = openSQLite(buildSQLiteDSN(dbPath, enc.Passphrase))
	if err != nil {
		return nil, fmt.Errorf("migration to encrypted mode completed, but reopening failed: %w", err)
	}
	if err := cleanupMigrationBackup(backupPath); err != nil {
		log.Printf("WARNING: failed to remove migration backup %s: %v", backupPath, err)
	}
	return db, nil
}

func connectPlain(dbPath string, enc SQLiteEncryptionConfig) (*gorm.DB, error) {
	db, err := openSQLite(buildSQLiteDSN(dbPath, ""))
	if err == nil {
		return db, nil
	}

	if !enc.AutoMigrate {
		return nil, fmt.Errorf("failed to open plain SQLite database at %s and DB_ENCRYPTION_AUTO_MIGRATE=false: %w", dbPath, err)
	}

	if strings.TrimSpace(enc.Passphrase) == "" {
		return nil, fmt.Errorf("failed to open plain SQLite database at %s and no passphrase is available for encrypted fallback: %w", dbPath, err)
	}

	encryptedProbe, encErr := openSQLite(buildSQLiteDSN(dbPath, enc.Passphrase))
	if encErr != nil {
		return nil, fmt.Errorf("failed to open SQLite database at %s as plain (%v) and encrypted (%v)", dbPath, err, encErr)
	}
	_ = closeSQLite(encryptedProbe)

	backupPath, err := migrateEncryptedToPlain(dbPath, enc.Passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate SQLite database to plain mode: %w", err)
	}

	db, err = openSQLite(buildSQLiteDSN(dbPath, ""))
	if err != nil {
		return nil, fmt.Errorf("migration to plain mode completed, but reopening failed: %w", err)
	}
	if err := cleanupMigrationBackup(backupPath); err != nil {
		log.Printf("WARNING: failed to remove migration backup %s: %v", backupPath, err)
	}
	return db, nil
}

func openSQLite(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	var tableCount int64
	if err := db.Raw("SELECT count(*) FROM sqlite_master").Scan(&tableCount).Error; err != nil {
		_ = closeSQLite(db)
		return nil, err
	}
	return db, nil
}

func closeSQLite(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func closeSQLiteIfOpen(source **gorm.DB) {
	if source == nil || *source == nil {
		return
	}
	_ = closeSQLite(*source)
	*source = nil
}

func closeSQLiteBeforeSwap(source **gorm.DB) error {
	if source == nil || *source == nil {
		return nil
	}
	if err := closeSQLite(*source); err != nil {
		return err
	}
	*source = nil
	return nil
}

func buildSQLiteDSN(dbPath, passphrase string) string {
	params := "_busy_timeout=10000&_journal_mode=WAL&_foreign_keys=1"
	if strings.TrimSpace(passphrase) != "" {
		params = params + "&_pragma_key=" + url.QueryEscape(passphrase)
	}
	return dbPath + "?" + params
}

func migratePlainToEncrypted(dbPath, passphrase string) (string, error) {
	targetPath := dbPath + ".enc.tmp"
	attach := fmt.Sprintf("ATTACH DATABASE '%s' AS target KEY '%s'", sqliteString(targetPath), sqliteString(passphrase))
	return migrateSQLiteDatabase(dbPath, "", targetPath, "plain-to-encrypted", attach, true)
}

func migrateEncryptedToPlain(dbPath, passphrase string) (string, error) {
	targetPath := dbPath + ".plain.tmp"
	attach := fmt.Sprintf("ATTACH DATABASE '%s' AS target KEY ''", sqliteString(targetPath))
	return migrateSQLiteDatabase(dbPath, passphrase, targetPath, "encrypted-to-plain", attach, false)
}

func migrateSQLiteDatabase(dbPath, sourcePassphrase, targetPath, migrationType, attachStmt string, targetEncrypted bool) (string, error) {
	source, err := openSQLite(buildSQLiteDSN(dbPath, sourcePassphrase))
	if err != nil {
		return "", err
	}
	defer closeSQLiteIfOpen(&source)

	if err := resetTempArtifacts(targetPath); err != nil {
		return "", err
	}
	if err := exportDatabase(source, attachStmt, targetEncrypted); err != nil {
		return "", err
	}
	if err := closeSQLiteBeforeSwap(&source); err != nil {
		return "", fmt.Errorf("failed to close source database before swap: %w", err)
	}

	return swapMigratedFile(dbPath, targetPath, migrationType)
}

func exportDatabase(source *gorm.DB, attachStmt string, targetEncrypted bool) error {
	if err := source.Exec(attachStmt).Error; err != nil {
		return err
	}
	if targetEncrypted {
		if err := source.Exec("PRAGMA target.cipher_page_size = 4096").Error; err != nil {
			return err
		}
	}
	if err := source.Exec("SELECT sqlcipher_export('target')").Error; err != nil {
		return err
	}
	if err := source.Exec("DETACH DATABASE target").Error; err != nil {
		return err
	}
	return nil
}

func swapMigratedFile(originalPath, targetPath, migrationType string) (string, error) {
	backupPath := fmt.Sprintf("%s.%s.%s.bak", originalPath, migrationType, time.Now().UTC().Format("20060102T150405"))

	if err := os.Rename(originalPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup before migration: %w", err)
	}
	if err := os.Rename(targetPath, originalPath); err != nil {
		_ = os.Rename(backupPath, originalPath)
		return "", fmt.Errorf("failed to swap migrated database file: %w", err)
	}

	_ = cleanupSQLiteSidecars(originalPath)
	_ = cleanupSQLiteSidecars(backupPath)
	_ = cleanupSQLiteSidecars(targetPath)
	return backupPath, nil
}

func cleanupMigrationBackup(backupPath string) error {
	if strings.TrimSpace(backupPath) == "" {
		return nil
	}
	if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return cleanupSQLiteSidecars(backupPath)
}

func resetTempArtifacts(path string) error {
	_ = os.Remove(path)
	if err := cleanupSQLiteSidecars(path); err != nil {
		return err
	}
	return nil
}

func cleanupSQLiteSidecars(path string) error {
	if err := os.Remove(path + "-wal"); err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.Remove(path + "-shm"); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func sqliteString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
