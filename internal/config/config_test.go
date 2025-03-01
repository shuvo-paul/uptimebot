package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			envVars: map[string]string{
				"SMTP_HOST":          "smtp.example.com",
				"SMTP_PORT":          "587",
				"SMTP_USERNAME":      "test@example.com",
				"SMTP_PASSWORD":      "password123",
				"SMTP_EMAIL_FROM":    "sender@example.com",
				"TURSO_DATABASE_URL": "libsql://test.turso.io",
				"TURSO_AUTH_TOKEN":   "valid-token",
				"BASE_URL":          "https://example.com",
			},
			want: &Config{
				Email: EmailConfig{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "test@example.com",
					Password: "password123",
					From:     "sender@example.com",
				},
				Database: DatabaseConfig{
					URL:   "libsql://test.turso.io",
					Token: "valid-token",
				},
				BaseURL: "https://example.com",
				Port:    8080,
			},
			wantErr: false,
		},
		{
			name: "custom port configuration",
			envVars: map[string]string{
				"SMTP_HOST":          "smtp.example.com",
				"SMTP_PORT":          "587",
				"SMTP_USERNAME":      "test@example.com",
				"SMTP_PASSWORD":      "password123",
				"SMTP_EMAIL_FROM":    "sender@example.com",
				"TURSO_DATABASE_URL": "libsql://test.turso.io",
				"TURSO_AUTH_TOKEN":   "valid-token",
				"BASE_URL":          "https://example.com",
				"PORT":              "3000",
			},
			want: &Config{
				Email: EmailConfig{
					Host:     "smtp.example.com",
					Port:     587,
					Username: "test@example.com",
					Password: "password123",
					From:     "sender@example.com",
				},
				Database: DatabaseConfig{
					URL:   "libsql://test.turso.io",
					Token: "valid-token",
				},
				BaseURL: "https://example.com",
				Port:    3000,
			},
			wantErr: false,
		},
		{
			name: "invalid application port",
			envVars: map[string]string{
				"SMTP_HOST":          "smtp.example.com",
				"SMTP_PORT":          "587",
				"SMTP_USERNAME":      "test@example.com",
				"SMTP_PASSWORD":      "password123",
				"SMTP_EMAIL_FROM":    "sender@example.com",
				"TURSO_DATABASE_URL": "libsql://test.turso.io",
				"TURSO_AUTH_TOKEN":   "valid-token",
				"BASE_URL":          "https://example.com",
				"PORT":              "invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing required fields",
			envVars: map[string]string{
				"SMTP_HOST": "smtp.example.com",
				"SMTP_PORT": "587",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing base URL",
			envVars: map[string]string{
				"SMTP_HOST":          "smtp.example.com",
				"SMTP_PORT":          "587",
				"SMTP_USERNAME":      "test@example.com",
				"SMTP_PASSWORD":      "password123",
				"SMTP_EMAIL_FROM":    "sender@example.com",
				"TURSO_DATABASE_URL": "libsql://test.turso.io",
				"TURSO_AUTH_TOKEN":   "valid-token",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid port number",
			envVars: map[string]string{
				"SMTP_HOST":     "smtp.example.com",
				"SMTP_PORT":     "invalid",
				"SMTP_USERNAME": "test@example.com",
				"SMTP_PASSWORD": "password123",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables before each test
			os.Clearenv()

			// Set environment variables for the test
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Run the test
			got, err := Load()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoadEmailConfig(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    EmailConfig
		wantErr bool
	}{
		{
			name: "valid email configuration",
			envVars: map[string]string{
				"SMTP_HOST":       "smtp.example.com",
				"SMTP_PORT":       "587",
				"SMTP_USERNAME":   "test@example.com",
				"SMTP_PASSWORD":   "password123",
				"SMTP_EMAIL_FROM": "sender@example.com",
			},
			want: EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "test@example.com",
				Password: "password123",
				From:     "sender@example.com",
			},
			wantErr: false,
		},
		{
			name: "missing host",
			envVars: map[string]string{
				"SMTP_PORT":     "587",
				"SMTP_USERNAME": "test@example.com",
				"SMTP_PASSWORD": "password123",
			},
			wantErr: true,
		},
		{
			name: "missing port",
			envVars: map[string]string{
				"SMTP_HOST":     "smtp.example.com",
				"SMTP_USERNAME": "test@example.com",
				"SMTP_PASSWORD": "password123",
			},
			wantErr: true,
		},
		{
			name: "missing username",
			envVars: map[string]string{
				"SMTP_HOST":     "smtp.example.com",
				"SMTP_PORT":     "587",
				"SMTP_PASSWORD": "password123",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			envVars: map[string]string{
				"SMTP_HOST":     "smtp.example.com",
				"SMTP_PORT":     "587",
				"SMTP_USERNAME": "test@example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid port format",
			envVars: map[string]string{
				"SMTP_HOST":     "smtp.example.com",
				"SMTP_PORT":     "invalid",
				"SMTP_USERNAME": "test@example.com",
				"SMTP_PASSWORD": "password123",
			},
			wantErr: true,
		},
		{
			name: "missing from field",
			envVars: map[string]string{
				"SMTP_HOST":     "smtp.example.com",
				"SMTP_PORT":     "587",
				"SMTP_USERNAME": "test@example.com",
				"SMTP_PASSWORD": "password123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables before each test
			os.Clearenv()

			// Set environment variables for the test
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Run the test
			got, err := loadEmailConfig()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
