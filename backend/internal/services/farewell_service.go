package services

import (
	"errors"

	"github.com/alpyxn/aeterna/backend/internal/database"
	"github.com/alpyxn/aeterna/backend/internal/models"
	"gorm.io/gorm"
)

type FarewellService struct{}

var farewellCrypto = CryptoService{}
var farewellValidation = ValidationService{}
var farewellFileService = FileService{}

func (s FarewellService) Create(userID, messageID, recipientEmail, subject, content string, delayMinutes int) (models.FarewellLetter, error) {
	var msg models.Message
	if err := database.ForTenant(userID).First(&msg, "id = ?", messageID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.FarewellLetter{}, NotFound("Message not found", err)
		}
		return models.FarewellLetter{}, Internal("Failed to fetch message", err)
	}

	if err := farewellValidation.ValidateEmail(recipientEmail); err != nil {
		return models.FarewellLetter{}, err
	}

	if subject == "" {
		return models.FarewellLetter{}, BadRequest("Subject is required", nil)
	}

	if err := farewellValidation.ValidateContent(content); err != nil {
		return models.FarewellLetter{}, err
	}

	if delayMinutes < 0 {
		return models.FarewellLetter{}, BadRequest("Delay must be zero or positive", nil)
	}

	encrypted, err := farewellCrypto.Encrypt(content)
	if err != nil {
		return models.FarewellLetter{}, err
	}

	letter := models.FarewellLetter{
		UserID:         userID,
		MessageID:      messageID,
		RecipientEmail: recipientEmail,
		Subject:        subject,
		Content:        encrypted,
		DelayMinutes:   delayMinutes,
		Status:         models.FarewellStatusPending,
	}

	if err := database.DB.Create(&letter).Error; err != nil {
		return models.FarewellLetter{}, Internal("Failed to create farewell letter", err)
	}

	letter.Content = content
	return letter, nil
}

func (s FarewellService) List(userID, messageID string) ([]models.FarewellLetter, error) {
	var msg models.Message
	if err := database.ForTenant(userID).First(&msg, "id = ?", messageID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, NotFound("Message not found", err)
		}
		return nil, Internal("Failed to fetch message", err)
	}

	var letters []models.FarewellLetter
	if err := database.ForTenant(userID).Where("message_id = ?", messageID).Order("created_at ASC").Find(&letters).Error; err != nil {
		return nil, Internal("Failed to fetch farewell letters", err)
	}

	for i := range letters {
		decrypted, err := farewellCrypto.Decrypt(letters[i].Content)
		if err != nil {
			return nil, err
		}
		letters[i].Content = decrypted

		count, _ := farewellFileService.CountFarewellAttachmentsByLetterID(userID, letters[i].ID)
		letters[i].AttachmentCount = count
	}

	return letters, nil
}

func (s FarewellService) Update(userID, messageID, id, recipientEmail, subject, content string, delayMinutes int) (models.FarewellLetter, error) {
	var letter models.FarewellLetter
	if err := database.ForTenant(userID).Where("message_id = ? AND id = ?", messageID, id).First(&letter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.FarewellLetter{}, NotFound("Farewell letter not found", err)
		}
		return models.FarewellLetter{}, Internal("Failed to fetch farewell letter", err)
	}

	if letter.Status == models.FarewellStatusSent {
		return models.FarewellLetter{}, BadRequest("Cannot edit an already-sent farewell letter", nil)
	}

	if err := farewellValidation.ValidateEmail(recipientEmail); err != nil {
		return models.FarewellLetter{}, err
	}

	if subject == "" {
		return models.FarewellLetter{}, BadRequest("Subject is required", nil)
	}

	if err := farewellValidation.ValidateContent(content); err != nil {
		return models.FarewellLetter{}, err
	}

	if delayMinutes < 0 {
		return models.FarewellLetter{}, BadRequest("Delay must be zero or positive", nil)
	}

	encrypted, err := farewellCrypto.Encrypt(content)
	if err != nil {
		return models.FarewellLetter{}, err
	}

	letter.RecipientEmail = recipientEmail
	letter.Subject = subject
	letter.Content = encrypted
	letter.DelayMinutes = delayMinutes

	if err := database.ForTenant(userID).Save(&letter).Error; err != nil {
		return models.FarewellLetter{}, Internal("Failed to update farewell letter", err)
	}

	letter.Content = content
	return letter, nil
}

func (s FarewellService) Delete(userID, messageID, id string) error {
	var letter models.FarewellLetter
	if err := database.ForTenant(userID).Where("message_id = ? AND id = ?", messageID, id).First(&letter).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFound("Farewell letter not found", err)
		}
		return Internal("Failed to fetch farewell letter", err)
	}

	if err := farewellFileService.DeleteFarewellAttachmentsByLetterID(userID, id); err != nil {
		return Internal("Failed to delete farewell attachments", err)
	}

	if err := database.ForTenant(userID).Unscoped().Delete(&letter).Error; err != nil {
		return Internal("Failed to delete farewell letter", err)
	}

	return nil
}
