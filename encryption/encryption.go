// Package encryption provides AES-128 key generation and HLS key_info management for segment encryption.
package encryption

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// Config defines the HLS AES-128 encryption configuration.
type Config struct {
	KeyURI string
	IV     string
	Key    []byte
}

// GenerateKey generates 16 cryptographically secure random bytes for AES-128 encryption.
func GenerateKey() ([]byte, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate AES-128 key: %w", err)
	}
	return key, nil
}

// GenerateIV generates a 16-byte random IV as a 32-character hex string.
func GenerateIV() (string, error) {
	ivBytes := make([]byte, 16)
	if _, err := rand.Read(ivBytes); err != nil {
		return "", fmt.Errorf("generate IV: %w", err)
	}
	return hex.EncodeToString(ivBytes), nil
}

// SetupKeyInfo writes the key and keyinfo file into outDir and returns the keyinfo file path.
func SetupKeyInfo(outDir string, cfg Config) (string, error) {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("create encryption directory: %w", err)
	}

	key := cfg.Key
	if len(key) == 0 {
		var err error
		key, err = GenerateKey()
		if err != nil {
			return "", err
		}
	} else if len(key) != 16 {
		return "", fmt.Errorf("invalid AES-128 key length: expected 16 bytes, got %d", len(key))
	}

	keyURI := cfg.KeyURI
	if keyURI == "" {
		keyURI = "enc.key"
	}

	keyPath := filepath.Join(outDir, "enc.key")
	if err := os.WriteFile(keyPath, key, 0o600); err != nil {
		return "", fmt.Errorf("write key file %s: %w", keyPath, err)
	}

	keyInfoPath := filepath.Join(outDir, "enc.keyinfo")
	var keyInfoContent string
	if cfg.IV != "" {
		keyInfoContent = fmt.Sprintf("%s\n%s\n%s\n", keyURI, keyPath, cfg.IV)
	} else {
		keyInfoContent = fmt.Sprintf("%s\n%s\n", keyURI, keyPath)
	}

	if err := os.WriteFile(keyInfoPath, []byte(keyInfoContent), 0o600); err != nil {
		return "", fmt.Errorf("write keyinfo file %s: %w", keyInfoPath, err)
	}

	return keyInfoPath, nil
}
