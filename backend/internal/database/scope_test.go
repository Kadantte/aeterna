package database

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alpyxn/aeterna/backend/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func testSQLiteDSN(t *testing.T) string {
	t.Helper()
	replacer := strings.NewReplacer("/", "_", " ", "_")
	return fmt.Sprintf("file:%s_%d?mode=memory&cache=shared&_foreign_keys=1", replacer.Replace(t.Name()), time.Now().UnixNano())
}

func TestScopeForUser_IsolatesRows(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(testSQLiteDSN(t)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Message{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Message{
		ID:              "msg-a",
		UserID:          "user-1",
		Content:         "enc",
		KeyFragment:     "v1",
		ManagementToken: "tok",
		RecipientEmail:  "a@a.com",
		TriggerDuration: 60,
		LastSeen:        time.Now(),
		Status:          models.StatusActive,
	}).Error; err != nil {
		t.Fatal(err)
	}

	var got models.Message
	err = db.Scopes(ScopeForUser("user-2")).First(&got, "id = ?", "msg-a").Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected no row for wrong tenant, got err=%v", err)
	}

	err = db.Scopes(ScopeForUser("user-1")).First(&got, "id = ?", "msg-a").Error
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "msg-a" {
		t.Fatalf("expected message for correct tenant")
	}
}

func TestForTenant_EmptyUserID_MatchesNothing(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(testSQLiteDSN(t)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Message{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Message{
		ID:              "msg-x",
		UserID:          "user-1",
		Content:         "enc",
		KeyFragment:     "v1",
		ManagementToken: "tok",
		RecipientEmail:  "a@a.com",
		TriggerDuration: 60,
		LastSeen:        time.Now(),
		Status:          models.StatusActive,
	}).Error; err != nil {
		t.Fatal(err)
	}

	prev := DB
	DB = db
	t.Cleanup(func() { DB = prev })

	var n int64
	if err := ForTenant("").Model(&models.Message{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("ForTenant(\"\") must match no rows, got count=%d", n)
	}
}

func TestForTenant_WrapsScopeForUser(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(testSQLiteDSN(t)), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Message{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.Message{
		ID:              "m1",
		UserID:          "u1",
		Content:         "e",
		KeyFragment:     "v1",
		ManagementToken: "t",
		RecipientEmail:  "a@a.com",
		TriggerDuration: 1,
		LastSeen:        time.Now(),
		Status:          models.StatusActive,
	}).Error; err != nil {
		t.Fatal(err)
	}

	prev := DB
	DB = db
	t.Cleanup(func() { DB = prev })

	var got models.Message
	if err := ForTenant("u1").First(&got, "id = ?", "m1").Error; err != nil {
		t.Fatal(err)
	}
	if got.ID != "m1" {
		t.Fatalf("ForTenant should behave like ScopeForUser")
	}
}
