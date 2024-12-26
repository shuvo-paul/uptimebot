package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotifierConfig_ScanValue(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    NotifierConfig
		wantErr bool
	}{
		{
			name: "valid slack config",
			input: []byte(`{
				"type": "slack",
				"config": {"webhook_url": "https://hooks.slack.com/test"}
			}`),
			want: NotifierConfig{
				Type:   NotifierTypeSlack,
				Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
			},
			wantErr: false,
		},
		{
			name: "valid email config",
			input: []byte(`{
				"type": "email",
				"config": {
					"host": "smtp.example.com",
					"port": 587,
					"username": "test",
					"password": "pass",
					"from": "test@example.com",
					"recipients": ["user1@example.com"]
				}
			}`),
			want: NotifierConfig{
				Type: NotifierTypeEmail,
				Config: json.RawMessage(`{
					"host": "smtp.example.com",
					"port": 587,
					"username": "test",
					"password": "pass",
					"from": "test@example.com",
					"recipients": ["user1@example.com"]
				}`),
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    NotifierConfig{},
			wantErr: false,
		},
		{
			name:    "invalid type",
			input:   123,
			wantErr: true,
		},
		{
			name:    "invalid json",
			input:   []byte(`invalid json`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got NotifierConfig
			err := got.Scan(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.input != nil {
				assert.Equal(t, tt.want.Type, got.Type)
				if len(tt.want.Config) > 0 {
					assert.JSONEq(t, string(tt.want.Config), string(got.Config))
				}
			}
		})
	}
}

func TestNotifierConfig_GetSlackConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      NotifierConfig
		want        *SlackConfig
		wantErr     bool
		description string
	}{
		{
			name: "valid slack config",
			config: NotifierConfig{
				Type:   NotifierTypeSlack,
				Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
			},
			want: &SlackConfig{
				WebhookURL: "https://hooks.slack.com/test",
			},
			wantErr: false,
		},
		{
			name: "wrong notifier type",
			config: NotifierConfig{
				Type:   NotifierTypeEmail,
				Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "invalid json",
			config: NotifierConfig{
				Type:   NotifierTypeSlack,
				Config: json.RawMessage(`invalid json`),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.GetSlackConfig()
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

func TestNotifierConfig_GetEmailConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  NotifierConfig
		want    *EmailConfig
		wantErr bool
	}{
		{
			name: "valid email config",
			config: NotifierConfig{
				Type: NotifierTypeEmail,
				Config: json.RawMessage(`{
					"host": "smtp.example.com",
					"port": 587,
					"username": "test",
					"password": "pass",
					"from": "test@example.com",
					"recipients": ["user1@example.com", "user2@example.com"]
				}`),
			},
			want: &EmailConfig{
				Host:       "smtp.example.com",
				Port:       587,
				Username:   "test",
				Password:   "pass",
				From:       "test@example.com",
				Recipients: []string{"user1@example.com", "user2@example.com"},
			},
			wantErr: false,
		},
		{
			name: "wrong notifier type",
			config: NotifierConfig{
				Type: NotifierTypeSlack,
				Config: json.RawMessage(`{
					"host": "smtp.example.com",
					"port": 587
				}`),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "invalid json",
			config: NotifierConfig{
				Type:   NotifierTypeEmail,
				Config: json.RawMessage(`invalid json`),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.config.GetEmailConfig()
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
