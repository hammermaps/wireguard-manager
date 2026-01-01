package util

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetPersistedSessionSecret verifies that the session secret is properly persisted
// and retrieved from the database.
func TestGetPersistedSessionSecret(t *testing.T) {
	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "wireguard-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// First call should generate and persist a new secret
	secret1 := GetPersistedSessionSecret(tmpDir)
	if secret1 == "" {
		t.Fatal("Expected non-empty session secret")
	}

	// Second call should retrieve the same persisted secret
	secret2 := GetPersistedSessionSecret(tmpDir)
	if secret1 != secret2 {
		t.Errorf("Expected same session secret on second call. Got %s, want %s", secret2, secret1)
	}

	// Verify the secret is actually stored in the database
	configDir := filepath.Join(tmpDir, "config")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Expected config directory to be created in database")
	}
}

// TestGetPersistedSessionSecretWithEnvVar verifies that environment variable takes precedence
func TestGetPersistedSessionSecretWithEnvVar(t *testing.T) {
	// Set environment variable
	expectedSecret := "test-secret-from-env"
	os.Setenv("SESSION_SECRET", expectedSecret)
	defer os.Unsetenv("SESSION_SECRET")

	// Create a temporary directory for the test database
	tmpDir, err := os.MkdirTemp("", "wireguard-manager-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Should return the environment variable value
	secret := GetPersistedSessionSecret(tmpDir)
	if secret != expectedSecret {
		t.Errorf("Expected secret from environment variable. Got %s, want %s", secret, expectedSecret)
	}

	// Verify no database entry was created since env var is used
	configDir := filepath.Join(tmpDir, "config")
	if _, err := os.Stat(configDir); !os.IsNotExist(err) {
		t.Error("Expected no config directory when using environment variable")
	}
}

// TestGetPersistedSessionSecretEmptyPath verifies fallback to default path
func TestGetPersistedSessionSecretEmptyPath(t *testing.T) {
	// Unset environment variable first
	os.Unsetenv("SESSION_SECRET")

	// Test with empty path - should use environment variable or default
	secret := GetPersistedSessionSecret("")
	if secret == "" {
		t.Fatal("Expected non-empty session secret with empty path")
	}
}
