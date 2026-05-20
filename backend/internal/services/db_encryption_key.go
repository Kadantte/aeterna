package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/hkdf"
)

const sqliteDBKeyKDFSalt = "aeterna-sqlite-db-key-v1"

// PrepareSQLiteEncryptionPassphrase derives a stable SQLCipher passphrase from:
// 1) the main encryption key material, and
// 2) a persisted KDF context file generated once.
func PrepareSQLiteEncryptionPassphrase(contextFile string) (string, error) {
	if strings.TrimSpace(contextFile) == "" {
		return "", fmt.Errorf("db encryption context file path is empty")
	}

	masterKeyB64, err := (CryptoService{}).getOrCreateKey()
	if err != nil {
		return "", err
	}

	masterKeyBytes, err := base64.StdEncoding.DecodeString(masterKeyB64)
	if err != nil {
		return "", fmt.Errorf("invalid master key format: %w", err)
	}

	contextValue, err := ensureKDFContextFile(contextFile)
	if err != nil {
		return "", err
	}

	reader := hkdf.New(sha256.New, masterKeyBytes, []byte(sqliteDBKeyKDFSalt), []byte(contextValue))
	derived := make([]byte, 32)
	if _, err := io.ReadFull(reader, derived); err != nil {
		return "", fmt.Errorf("failed to derive sqlite encryption key: %w", err)
	}

	return hex.EncodeToString(derived), nil
}

func ensureKDFContextFile(contextFile string) (string, error) {
	if data, err := os.ReadFile(contextFile); err == nil {
		value := strings.TrimSpace(string(data))
		if value == "" {
			return "", fmt.Errorf("db encryption context file is empty: %s", contextFile)
		}
		return value, nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to read db encryption context file: %w", err)
	}

	dir := filepath.Dir(contextFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create db encryption context directory: %w", err)
	}

	// Generate a stable random context once (64 hex chars / 32 bytes entropy).
	random := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, random); err != nil {
		return "", fmt.Errorf("failed to generate db encryption context: %w", err)
	}
	contextValue := hex.EncodeToString(random)

	tmp := contextFile + ".tmp"
	if err := os.WriteFile(tmp, []byte(contextValue+"\n"), 0600); err != nil {
		return "", fmt.Errorf("failed to write db encryption context file: %w", err)
	}
	if err := os.Rename(tmp, contextFile); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("failed to persist db encryption context file: %w", err)
	}

	return contextValue, nil
}
