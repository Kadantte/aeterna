package services

import (
	"testing"
	"time"

	"github.com/alpyxn/aeterna/backend/internal/database"
	"github.com/alpyxn/aeterna/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared&_foreign_keys=1"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&models.Message{},
		&models.MessageReminder{},
		&models.Attachment{},
		&models.FarewellLetter{},
		&models.FarewellAttachment{},
	); err != nil {
		t.Fatal(err)
	}
	prev := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = prev })
	return db
}

func TestMessageDelete_NoFarewellNoAttachments(t *testing.T) {
	db := setupTestDB(t)
	if err := db.Create(&models.Message{
		ID: "m1", UserID: "u1", Content: "x", KeyFragment: "v1",
		ManagementToken: "tok", RecipientEmail: "a@a.com",
		TriggerDuration: 60, LastSeen: time.Now(), Status: models.StatusActive,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := (MessageService{}).Delete("u1", "m1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestMessageDelete_WithFarewellLetter(t *testing.T) {
	db := setupTestDB(t)
	if err := db.Create(&models.Message{
		ID: "m1", UserID: "u1", Content: "x", KeyFragment: "v1",
		ManagementToken: "tok", RecipientEmail: "a@a.com",
		TriggerDuration: 60, LastSeen: time.Now(), Status: models.StatusActive,
	}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.FarewellLetter{
		ID: "l1", UserID: "u1", MessageID: "m1",
		RecipientEmail: "b@b.com", Subject: "bye", Content: "x",
		DelayMinutes: 60, Status: models.FarewellStatusPending,
	}).Error; err != nil {
		t.Fatal(err)
	}

	if err := (MessageService{}).Delete("u1", "m1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	var count int64
	db.Unscoped().Model(&models.FarewellLetter{}).Where("message_id = ?", "m1").Count(&count)
	if count != 0 {
		t.Fatalf("expected 0 farewell letters after delete, got %d", count)
	}
}
