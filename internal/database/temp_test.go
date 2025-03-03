package database

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewTempDir(t *testing.T) {
	// Test successful creation
	tempDBDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer tempDBDir.Cleanup()

	// Verify directory exists
	if _, err := os.Stat(tempDBDir.Dir()); os.IsNotExist(err) {
		t.Error("Temporary directory was not created")
	}

	// Verify directory name prefix
	if !strings.HasPrefix(filepath.Base(tempDBDir.Dir()), "libsql-") {
		t.Error("Temporary directory does not have expected prefix")
	}

	// Verify database path
	expectedDBPath := filepath.Join(tempDBDir.Dir(), "local.db")
	if tempDBDir.DBPath() != expectedDBPath {
		t.Errorf("Expected database path %s, got %s", expectedDBPath, tempDBDir.DBPath())
	}
}

func TestTempDirCleanup(t *testing.T) {
	// Create temporary directory
	tempDBDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Store directory path for verification after cleanup
	dirPath := tempDBDir.Dir()

	// Perform cleanup
	tempDBDir.Cleanup()

	// Verify directory no longer exists
	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		t.Error("Temporary directory still exists after cleanup")
	}
}

func TestTempDirMethods(t *testing.T) {
	// Create temporary directory
	tempDBDir, err := NewTempDir()
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer tempDBDir.Cleanup()

	// Test Dir() method
	if tempDBDir.Dir() == "" {
		t.Error("Dir() returned empty string")
	}

	// Test DBPath() method
	if tempDBDir.DBPath() == "" {
		t.Error("DBPath() returned empty string")
	}

	// Verify DBPath is inside Dir
	if !strings.HasPrefix(tempDBDir.DBPath(), tempDBDir.Dir()) {
		t.Error("Database path is not inside temporary directory")
	}
}
