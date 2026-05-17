package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alpyxn/aeterna/backend/internal/models"
)

func initTestKeyManager(t *testing.T) {
	t.Helper()
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "enc.key")
	if err := os.WriteFile(keyPath, []byte(key), 0600); err != nil {
		t.Fatalf("failed to write test key: %v", err)
	}
	if err := os.Chmod(keyPath, 0600); err != nil {
		t.Fatalf("failed to chmod test key: %v", err)
	}

	InitKeyManager(keyPath)
}

func TestFarewellCreate_PersistsZeroDelay(t *testing.T) {
	db := setupTestDB(t)
	initTestKeyManager(t)

	msg := models.Message{
		ID:              "m-delay-zero",
		UserID:          "u-delay-zero",
		Content:         "encrypted",
		KeyFragment:     "v1",
		ManagementToken: "tok",
		RecipientEmail:  "owner@example.com",
		TriggerDuration: 60,
		LastSeen:        time.Now(),
		Status:          models.StatusActive,
	}
	if err := db.Create(&msg).Error; err != nil {
		t.Fatal(err)
	}

	letter, err := (FarewellService{}).Create(
		msg.UserID,
		msg.ID,
		"recipient@example.com",
		"Subject",
		"Content",
		0,
	)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if letter.DelayMinutes != 0 {
		t.Fatalf("expected returned delay 0, got %d", letter.DelayMinutes)
	}

	var stored models.FarewellLetter
	if err := db.First(&stored, "id = ?", letter.ID).Error; err != nil {
		t.Fatalf("failed to load stored farewell letter: %v", err)
	}
	if stored.DelayMinutes != 0 {
		t.Fatalf("expected stored delay 0, got %d", stored.DelayMinutes)
	}
}
