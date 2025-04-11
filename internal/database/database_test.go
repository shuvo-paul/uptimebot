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
			name: "empty host",
			config: config.DatabaseConfig{
				Port:     "5432",
				User:     "test",
				Password: "test",
				DBName:   "test",
				SSLMode:  "disable",
			},
			wantErr:     true,
			expectedErr: "failed to connect to database",
		},
		{
			name: "invalid port",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     "invalid",
				User:     "test",
				Password: "test",
				DBName:   "test",
				SSLMode:  "disable",
			},
			wantErr:     true,
			expectedErr: "failed to connect to database",
		},
		{
			name: "invalid credentials",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "invalid",
				Password: "invalid",
				DBName:   "test",
				SSLMode:  "disable",
			},
			wantErr:     true,
			expectedErr: "failed to connect to database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := InitDatabase(tt.config)
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
