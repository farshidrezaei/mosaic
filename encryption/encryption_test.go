package encryption

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	k1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if len(k1) != 16 {
		t.Errorf("expected 16 bytes key, got %d", len(k1))
	}

	k2, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey failed: %v", err)
	}
	if string(k1) == string(k2) {
		t.Errorf("expected unique random keys, got identical keys")
	}
}

func TestGenerateIV(t *testing.T) {
	iv, err := GenerateIV()
	if err != nil {
		t.Fatalf("GenerateIV failed: %v", err)
	}
	if len(iv) != 32 {
		t.Errorf("expected 32 hex chars, got %d (%s)", len(iv), iv)
	}
}

func TestSetupKeyInfo(t *testing.T) {
	tmpDir := t.TempDir()

	keyInfoPath, err := SetupKeyInfo(tmpDir, Config{
		KeyURI: "https://example.com/keys/video.key",
	})
	if err != nil {
		t.Fatalf("SetupKeyInfo failed: %v", err)
	}

	keyPath := filepath.Join(tmpDir, "enc.key")
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("failed to read key file: %v", err)
	}
	if len(keyData) != 16 {
		t.Errorf("expected 16 bytes key file, got %d", len(keyData))
	}

	infoData, err := os.ReadFile(keyInfoPath)
	if err != nil {
		t.Fatalf("failed to read keyinfo file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(infoData)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines in keyinfo, got %d", len(lines))
	}
	if lines[0] != "https://example.com/keys/video.key" {
		t.Errorf("expected key URI in line 0, got %s", lines[0])
	}
	if lines[1] != keyPath {
		t.Errorf("expected key path in line 1, got %s", lines[1])
	}
}

func TestSetupKeyInfo_CustomKeyAndIV(t *testing.T) {
	tmpDir := t.TempDir()
	customKey := []byte("0123456789abcdef")
	customIV := "000102030405060708090a0b0c0d0e0f"

	keyInfoPath, err := SetupKeyInfo(tmpDir, Config{
		Key:    customKey,
		KeyURI: "enc.key",
		IV:     customIV,
	})
	if err != nil {
		t.Fatalf("SetupKeyInfo with custom key failed: %v", err)
	}

	infoData, err := os.ReadFile(keyInfoPath)
	if err != nil {
		t.Fatalf("failed to read keyinfo: %v", err)
	}

	if !strings.Contains(string(infoData), customIV) {
		t.Errorf("expected custom IV in keyinfo:\n%s", string(infoData))
	}
}
