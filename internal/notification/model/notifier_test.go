package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotifier_GetSlackConfig(t *testing.T) {
	tests := []struct {
		name     string
		notifier *Notifier
		want     *SlackConfig
		wantErr  bool
	}{
		{
			name: "valid slack config",
			notifier: &Notifier{
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
			notifier: &Notifier{
				Type:   NotifierTypeEmail,
				Config: json.RawMessage(`{"webhook_url": "https://hooks.slack.com/test"}`),
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "invalid json",
			notifier: &Notifier{
				Type:   NotifierTypeSlack,
				Config: json.RawMessage(`invalid json`),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.notifier.GetSlackConfig()
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

func TestNotifier_GetEmailConfig(t *testing.T) {
	tests := []struct {
		name     string
		notifier *Notifier
		want     *EmailConfig
		wantErr  bool
	}{
		{
			name: "valid email config",
			notifier: &Notifier{
				Type: NotifierTypeEmail,
				Config: json.RawMessage(`{
					"recipients": ["user1@example.com", "user2@example.com"]
				}`),
			},
			want: &EmailConfig{
				Recipients: []string{"user1@example.com", "user2@example.com"},
			},
			wantErr: false,
		},
		{
			name: "wrong notifier type",
			notifier: &Notifier{
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
			notifier: &Notifier{
				Type:   NotifierTypeEmail,
				Config: json.RawMessage(`invalid json`),
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.notifier.GetEmailConfig()
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
