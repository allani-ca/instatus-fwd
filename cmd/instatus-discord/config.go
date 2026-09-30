package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

// Config contains the service settings loaded from the environment.
type Config struct {
	DiscordWebhookURL     string
	InstatusWebhookSecret string
	ListenAddr            string
	StatusPageName        string
	DiscordTimeout        time.Duration
	ReadHeaderTimeout     time.Duration
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
}

// loadConfig reads required settings and applies validated defaults.
func loadConfig() (Config, error) {
	cfg := Config{
		DiscordWebhookURL:     os.Getenv("DISCORD_WEBHOOK_URL"),
		InstatusWebhookSecret: os.Getenv("INSTATUS_WEBHOOK_SECRET"),
		ListenAddr:            os.Getenv("LISTEN_ADDR"),
		StatusPageName:        os.Getenv("STATUS_PAGE_NAME"),
	}

	if err := validateDiscordWebhookURL(cfg.DiscordWebhookURL); err != nil {
		return Config{}, err
	}
	if strings.TrimSpace(cfg.InstatusWebhookSecret) == "" {
		return Config{}, errors.New("INSTATUS_WEBHOOK_SECRET is not configured")
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "127.0.0.1:8080"
	}

	if cfg.StatusPageName == "" {
		cfg.StatusPageName = "BSky Status"
	}

	durations := []struct {
		name     string
		fallback time.Duration
		target   *time.Duration
	}{
		{"DISCORD_TIMEOUT", 10 * time.Second, &cfg.DiscordTimeout},
		{"READ_HEADER_TIMEOUT", 5 * time.Second, &cfg.ReadHeaderTimeout},
		{"READ_TIMEOUT", 10 * time.Second, &cfg.ReadTimeout},
		{"WRITE_TIMEOUT", 15 * time.Second, &cfg.WriteTimeout},
		{"IDLE_TIMEOUT", 60 * time.Second, &cfg.IdleTimeout},
	}
	for _, setting := range durations {
		value := os.Getenv(setting.name)
		if value == "" {
			*setting.target = setting.fallback
			continue
		}

		duration, err := time.ParseDuration(value)
		if err != nil || duration <= 0 {
			return Config{}, fmt.Errorf("%s must be a positive duration: %q", setting.name, value)
		}
		*setting.target = duration
	}

	return cfg, nil
}

// validateDiscordWebhookURL requires an HTTPS URL for a Discord webhook endpoint.
func validateDiscordWebhookURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return errors.New("DISCORD_WEBHOOK_URL is not configured")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New("DISCORD_WEBHOOK_URL must be a valid HTTPS Discord webhook URL")
	}

	host := strings.ToLower(parsed.Hostname())
	if host != "discord.com" && host != "discordapp.com" {
		return errors.New("DISCORD_WEBHOOK_URL must use discord.com or discordapp.com")
	}
	pathParts := strings.Split(parsed.Path, "/")
	if parsed.Port() != "" || len(pathParts) != 5 || pathParts[1] != "api" || pathParts[2] != "webhooks" || pathParts[3] == "" || pathParts[4] == "" {
		return errors.New("DISCORD_WEBHOOK_URL must contain a Discord webhook endpoint")
	}

	return nil
}
