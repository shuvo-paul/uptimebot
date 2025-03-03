package database

import (
	"testing"

	"github.com/shuvo-paul/uptimebot/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestInitDatabase(t *testing.T) {
	tests := []struct {
		name        string
		config      config.DatabaseConfig
		wantErr     bool
		expectedErr string
	}{
		{
			name: "empty URL",
			config: config.DatabaseConfig{
				URL:   "",
				Token: "valid-token",
			},
			wantErr:     true,
			expectedErr: "database URL is empty",
		},
		{
			name: "empty token",
			config: config.DatabaseConfig{
				URL:   "libsql://test.turso.io",
				Token: "",
			},
			wantErr:     true,
			expectedErr: "database token is empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for each test
			tempDir, err := NewTempDir()
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer tempDir.Cleanup()

			db, err := InitDatabase(tt.config, tempDir)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, db)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
				if db != nil {
					db.Close()
				}
			}
		})
	}
}
