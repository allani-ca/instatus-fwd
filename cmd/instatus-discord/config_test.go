package main

import (
	"strings"
	"testing"
	"time"
)

func setRequiredConfig(t *testing.T) {
	t.Helper()
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/api/webhooks/123456/token")
	t.Setenv("INSTATUS_WEBHOOK_SECRET", "test-secret")
	t.Setenv("LISTEN_ADDR", "")
	t.Setenv("STATUS_PAGE_NAME", "")
	t.Setenv("DISCORD_TIMEOUT", "")
	t.Setenv("READ_HEADER_TIMEOUT", "")
	t.Setenv("READ_TIMEOUT", "")
	t.Setenv("WRITE_TIMEOUT", "")
	t.Setenv("IDLE_TIMEOUT", "")
}

func TestLoadConfigDefaultsAndTimeoutOverrides(t *testing.T) {
	setRequiredConfig(t)
	t.Setenv("READ_TIMEOUT", "3s")
	t.Setenv("DISCORD_TIMEOUT", "2s")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if cfg.ListenAddr != "127.0.0.1:8080" || cfg.StatusPageName != "BSky Status" {
		t.Fatalf("unexpected defaults: listen=%q name=%q", cfg.ListenAddr, cfg.StatusPageName)
	}
	if cfg.DiscordTimeout != 2*time.Second || cfg.ReadTimeout != 3*time.Second {
		t.Fatalf("unexpected timeout values: discord=%s read=%s", cfg.DiscordTimeout, cfg.ReadTimeout)
	}
	if cfg.ReadHeaderTimeout != 5*time.Second || cfg.WriteTimeout != 15*time.Second || cfg.IdleTimeout != 60*time.Second {
		t.Fatalf("unexpected timeout defaults: %+v", cfg)
	}
}

func TestLoadConfigRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name      string
		set       func(*testing.T)
		wantError string
	}{
		{
			name: "missing URL",
			set: func(t *testing.T) {
				t.Setenv("DISCORD_WEBHOOK_URL", "")
			},
			wantError: "DISCORD_WEBHOOK_URL is not configured",
		},
		{
			name: "invalid webhook endpoint",
			set: func(t *testing.T) {
				t.Setenv("DISCORD_WEBHOOK_URL", "https://example.com/api/webhooks/123/token")
			},
			wantError: "must use discord.com or discordapp.com",
		},
		{
			name: "missing secret",
			set: func(t *testing.T) {
				t.Setenv("INSTATUS_WEBHOOK_SECRET", " ")
			},
			wantError: "INSTATUS_WEBHOOK_SECRET is not configured",
		},
		{
			name: "invalid timeout",
			set: func(t *testing.T) {
				t.Setenv("READ_TIMEOUT", "0s")
			},
			wantError: "READ_TIMEOUT must be a positive duration",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setRequiredConfig(t)
			test.set(t)

			_, err := loadConfig()
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("loadConfig() error = %v, want containing %q", err, test.wantError)
			}
		})
	}
}

func TestValidateDiscordWebhookURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"https://discord.com/api/webhooks/123/token", false},
		{"https://discordapp.com/api/webhooks/123/token", false},
		{"http://discord.com/api/webhooks/123/token", true},
		{"https://discord.com/api/webhooks/123/", true},
		{"https://discord.com/api/webhooks/123/token/extra", true},
		{"https://discord.com.evil.test/api/webhooks/123/token", true},
	}

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			err := validateDiscordWebhookURL(test.url)
			if (err != nil) != test.wantErr {
				t.Fatalf("validateDiscordWebhookURL() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}
