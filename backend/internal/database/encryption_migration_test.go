package database

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMigratePlainToEncryptedAndBack(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "aeterna.db")
	passphrase := "test-passphrase-for-sqlcipher"

	plain, err := openSQLite(buildSQLiteDSN(dbPath, ""))
	if err != nil {
		t.Fatalf("failed to open initial plain db: %v", err)
	}
	if err := plain.Exec("CREATE TABLE test_rows (id INTEGER PRIMARY KEY, value TEXT)").Error; err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}
	if err := plain.Exec("INSERT INTO test_rows(value) VALUES ('alpha')").Error; err != nil {
		t.Fatalf("failed to insert seed row: %v", err)
	}
	if err := closeSQLite(plain); err != nil {
		t.Fatalf("failed to close initial plain db: %v", err)
	}

	if _, err := migratePlainToEncrypted(dbPath, passphrase); err != nil {
		t.Fatalf("failed to migrate plain->encrypted: %v", err)
	}

	encrypted, err := openSQLite(buildSQLiteDSN(dbPath, passphrase))
	if err != nil {
		t.Fatalf("failed to open encrypted db with key: %v", err)
	}
	var count int
	if err := encrypted.Raw("SELECT count(*) FROM test_rows").Scan(&count).Error; err != nil {
		t.Fatalf("failed to query encrypted db: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row in encrypted db, got %d", count)
	}
	if err := closeSQLite(encrypted); err != nil {
		t.Fatalf("failed to close encrypted db: %v", err)
	}

	if _, err := openSQLite(buildSQLiteDSN(dbPath, "")); err == nil {
		t.Fatal("expected plain open to fail after encryption migration")
	}

	if _, err := migrateEncryptedToPlain(dbPath, passphrase); err != nil {
		t.Fatalf("failed to migrate encrypted->plain: %v", err)
	}

	plainAgain, err := openSQLite(buildSQLiteDSN(dbPath, ""))
	if err != nil {
		t.Fatalf("failed to reopen plain db after reverse migration: %v", err)
	}
	if err := plainAgain.Raw("SELECT count(*) FROM test_rows").Scan(&count).Error; err != nil {
		t.Fatalf("failed to query plain db after reverse migration: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after reverse migration, got %d", count)
	}
	if err := closeSQLite(plainAgain); err != nil {
		t.Fatalf("failed to close plain db after reverse migration: %v", err)
	}
}

func TestConnectEncrypted_RemovesBackupAfterSuccessfulAutoMigration(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "aeterna.db")
	passphrase := "test-passphrase-for-sqlcipher"

	plain, err := openSQLite(buildSQLiteDSN(dbPath, ""))
	if err != nil {
		t.Fatalf("failed to open initial plain db: %v", err)
	}
	if err := plain.Exec("CREATE TABLE test_rows (id INTEGER PRIMARY KEY, value TEXT)").Error; err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}
	if err := closeSQLite(plain); err != nil {
		t.Fatalf("failed to close initial plain db: %v", err)
	}

	db, err := connectEncrypted(dbPath, SQLiteEncryptionConfig{
		Enabled:     true,
		AutoMigrate: true,
		Passphrase:  passphrase,
	})
	if err != nil {
		t.Fatalf("connectEncrypted failed: %v", err)
	}
	if err := closeSQLite(db); err != nil {
		t.Fatalf("failed to close encrypted db: %v", err)
	}

	checkNoBackupArtifacts(t, tempDir, "plain-to-encrypted")
}

func TestConnectPlain_RemovesBackupAfterSuccessfulAutoMigration(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "aeterna.db")
	passphrase := "test-passphrase-for-sqlcipher"

	plain, err := openSQLite(buildSQLiteDSN(dbPath, ""))
	if err != nil {
		t.Fatalf("failed to open initial plain db: %v", err)
	}
	if err := plain.Exec("CREATE TABLE test_rows (id INTEGER PRIMARY KEY, value TEXT)").Error; err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}
	if err := closeSQLite(plain); err != nil {
		t.Fatalf("failed to close initial plain db: %v", err)
	}

	if _, err := migratePlainToEncrypted(dbPath, passphrase); err != nil {
		t.Fatalf("failed to prepare encrypted db: %v", err)
	}

	db, err := connectPlain(dbPath, SQLiteEncryptionConfig{
		Enabled:     false,
		AutoMigrate: true,
		Passphrase:  passphrase,
	})
	if err != nil {
		t.Fatalf("connectPlain failed: %v", err)
	}
	if err := closeSQLite(db); err != nil {
		t.Fatalf("failed to close plain db: %v", err)
	}

	checkNoBackupArtifacts(t, tempDir, "encrypted-to-plain")
}

func checkNoBackupArtifacts(t *testing.T, dir, migrationType string) {
	t.Helper()
	entries, err := filepath.Glob(filepath.Join(dir, "aeterna.db."+migrationType+".*.bak*"))
	if err != nil {
		t.Fatalf("failed to scan backup artifacts: %v", err)
	}
	if len(entries) > 0 {
		t.Fatalf("expected no backup artifact after successful migration, found: %s", strings.Join(entries, ", "))
	}
}
