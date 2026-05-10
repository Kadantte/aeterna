package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageStatus string

const (
	StatusActive    MessageStatus = "active"
	StatusTriggered MessageStatus = "triggered"
)

type Message struct {
	ID              string            `gorm:"type:text;primaryKey" json:"id"`
	UserID          string            `gorm:"type:text;index" json:"-"`
	Content         string            `gorm:"column:encrypted_content;not null" json:"content"`
	KeyFragment     string            `gorm:"column:key_fragment;not null" json:"-"`
	ManagementToken string            `gorm:"column:management_token;not null" json:"-"`
	RecipientEmail  string            `gorm:"not null" json:"recipient_email"`
	TriggerDuration int               `gorm:"not null" json:"trigger_duration"`
	LastSeen        time.Time         `gorm:"not null;default:CURRENT_TIMESTAMP" json:"last_seen"`
	Status          MessageStatus     `gorm:"default:'active'" json:"status"`
	TriggeredAt     *time.Time        `json:"triggered_at,omitempty"`
	Reminders       []MessageReminder `gorm:"foreignKey:MessageID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"reminders"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	DeletedAt       gorm.DeletedAt    `gorm:"index" json:"-"`
	AttachmentCount int64             `gorm:"-" json:"attachment_count"`
}

// BeforeCreate hook to generate UUID before creating
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	if m.ManagementToken == "" {
		m.ManagementToken = uuid.NewString()
	}
	return nil
}

// BeforeDelete cascades the delete to associated FarewellLetters and their attachments,
// mirroring the soft/hard mode of the parent operation.
func (m *Message) BeforeDelete(tx *gorm.DB) error {
	if m.ID == "" {
		return nil
	}
	db := tx
	if tx.Statement.Unscoped {
		db = tx.Unscoped()
	}
	var letterIDs []string
	if err := db.Model(&FarewellLetter{}).Select("id").Where("message_id = ?", m.ID).Find(&letterIDs).Error; err != nil {
		return err
	}
	if len(letterIDs) == 0 {
		return nil
	}
	if err := db.Where("letter_id IN ?", letterIDs).Delete(&FarewellAttachment{}).Error; err != nil {
		return err
	}
	return db.Where("id IN ?", letterIDs).Delete(&FarewellLetter{}).Error
}
