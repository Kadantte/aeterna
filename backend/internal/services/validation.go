package services

import (
	"html"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

type ValidationService struct{}

const (
	MaxContentLength     = 50000
	MinContentLength     = 1
	MaxEmailLength       = 254
	MaxRecipientEmails   = 20
	MaxFileSize          = 10 * 1024 * 1024 // 10 MB
	MaxTotalAttachSize   = 25 * 1024 * 1024 // 25 MB
	MaxAttachmentsPerMsg = 5

	// Farewell letter attachments use email-provider limits as the practical ceiling.
	MaxFarewellFileSize    = 20 * 1024 * 1024 // 20 MB per file
	MaxFarewellTotalSize   = 50 * 1024 * 1024 // 50 MB total
	MaxFarewellAttachments = 10
)

var AllowedExtensions = map[string]bool{
	".pdf":  true,
	".txt":  true,
	".doc":  true,
	".docx": true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".zip":  true,
}

// AllowedFarewellExtensions composes AllowedExtensions and adds audio/video formats.
var AllowedFarewellExtensions = func() map[string]bool {
	m := make(map[string]bool, len(AllowedExtensions)+9)
	for k, v := range AllowedExtensions {
		m[k] = v
	}
	for _, ext := range []string{".mp3", ".wav", ".ogg", ".m4a", ".aac", ".mp4", ".mov", ".webm", ".avi"} {
		m[ext] = true
	}
	return m
}()

var AllowedMIMEPrefixes = []string{
	"application/pdf",
	"text/plain",
	"application/msword",
	"application/vnd.openxmlformats",
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
	"application/zip",
	"application/octet-stream", // fallback for unknown binary types like .docx
}

// AllowedFarewellMIMEPrefixes composes AllowedMIMEPrefixes and adds audio/video MIME types.
// application/ogg is explicit because http.DetectContentType returns it for .ogg files,
// not the audio/ogg variant.
var AllowedFarewellMIMEPrefixes = append(
	append([]string{}, AllowedMIMEPrefixes...),
	"audio/", "video/", "application/ogg",
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

func (s ValidationService) ValidateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return BadRequest("Email is required", nil)
	}

	if len(email) > MaxEmailLength {
		return BadRequest("Email address is too long", nil)
	}

	if !emailRegex.MatchString(email) {
		return BadRequest("Invalid email format", nil)
	}

	// Check for common dangerous patterns
	lowerEmail := strings.ToLower(email)
	dangerousPatterns := []string{"<script", "javascript:", "data:", "vbscript:"}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lowerEmail, pattern) {
			return BadRequest("Invalid email format", nil)
		}
	}

	return nil
}

func (s ValidationService) ValidateEmailListLength(count int) error {
	if count < 1 {
		return BadRequest("At least one recipient email is required", nil)
	}
	if count > MaxRecipientEmails {
		return BadRequest("Too many recipient emails (max 20)", nil)
	}
	return nil
}

func (s ValidationService) ValidateContent(content string) error {
	if len(content) < MinContentLength {
		return BadRequest("Content is required", nil)
	}

	if len(content) > MaxContentLength {
		return BadRequest("Content exceeds maximum length of 50000 characters", nil)
	}

	return nil
}

func (s ValidationService) SanitizeContent(content string) string {
	// HTML escape to prevent XSS
	sanitized := html.EscapeString(content)
	return sanitized
}

func (s ValidationService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return BadRequest("Password must be at least 8 characters", nil)
	}

	if len(password) > 128 {
		return BadRequest("Password exceeds maximum length", nil)
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return BadRequest("Password must contain at least one uppercase letter", nil)
	}
	if !hasLower {
		return BadRequest("Password must contain at least one lowercase letter", nil)
	}
	if !hasNumber {
		return BadRequest("Password must contain at least one number", nil)
	}
	if !hasSpecial {
		return BadRequest("Password must contain at least one special character (!@#$%^&* etc.)", nil)
	}

	return nil
}

// ValidateTriggerDuration validates the trigger duration in minutes
func (s ValidationService) ValidateTriggerDuration(duration int) error {
	if duration < 1 {
		return BadRequest("Duration must be at least 1 minute", nil)
	}
	if duration > 525600 {
		return BadRequest("Duration cannot exceed 1 year (525600 minutes)", nil)
	}
	return nil
}

type fileValidationOptions struct {
	maxSize      int64
	sizeErrMsg   string
	extensions   map[string]bool
	extErrMsg    string
	mimePrefixes []string
}

func (s ValidationService) validateFileWith(filename string, size int64, data []byte, opts fileValidationOptions) error {
	if size == 0 {
		return BadRequest("File is empty", nil)
	}
	if size > opts.maxSize {
		return BadRequest(opts.sizeErrMsg, nil)
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return BadRequest("File must have an extension", nil)
	}
	if !opts.extensions[ext] {
		return BadRequest(opts.extErrMsg, nil)
	}
	detectedMIME := http.DetectContentType(data)
	for _, prefix := range opts.mimePrefixes {
		if strings.HasPrefix(detectedMIME, prefix) {
			return nil
		}
	}
	return BadRequest("File content type not allowed", nil)
}

// ValidateFile validates a switch attachment: extension, MIME type, and size.
func (s ValidationService) ValidateFile(filename string, size int64, data []byte) error {
	return s.validateFileWith(filename, size, data, fileValidationOptions{
		maxSize:      MaxFileSize,
		sizeErrMsg:   "File exceeds maximum size of 10 MB",
		extensions:   AllowedExtensions,
		extErrMsg:    "File type not allowed. Allowed: PDF, TXT, DOC, DOCX, JPG, PNG, GIF, WEBP, ZIP",
		mimePrefixes: AllowedMIMEPrefixes,
	})
}

// ValidateFarewellFile validates a farewell letter attachment using provider-ceiling limits.
func (s ValidationService) ValidateFarewellFile(filename string, size int64, data []byte) error {
	return s.validateFileWith(filename, size, data, fileValidationOptions{
		maxSize:      MaxFarewellFileSize,
		sizeErrMsg:   "File exceeds maximum size of 20 MB",
		extensions:   AllowedFarewellExtensions,
		extErrMsg:    "File type not allowed. Allowed: PDF, TXT, DOC, DOCX, JPG, PNG, GIF, WEBP, ZIP, MP3, WAV, OGG, M4A, AAC, MP4, MOV, WEBM, AVI",
		mimePrefixes: AllowedFarewellMIMEPrefixes,
	})
}

// SanitizeFilename cleans a filename to prevent path traversal and other attacks
func (s ValidationService) SanitizeFilename(filename string) string {
	// Extract just the base name (no directory path)
	filename = filepath.Base(filename)

	// Remove null bytes and control characters
	cleaned := strings.Map(func(r rune) rune {
		if r == 0 || unicode.IsControl(r) {
			return -1
		}
		return r
	}, filename)

	// Replace path separators
	cleaned = strings.ReplaceAll(cleaned, "/", "_")
	cleaned = strings.ReplaceAll(cleaned, "\\", "_")

	// Remove leading dots (hidden files)
	cleaned = strings.TrimLeft(cleaned, ".")

	if cleaned == "" {
		cleaned = "unnamed_file"
	}

	// Limit length
	if len(cleaned) > 255 {
		ext := filepath.Ext(cleaned)
		name := cleaned[:255-len(ext)]
		cleaned = name + ext
	}

	return cleaned
}
