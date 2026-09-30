package main

import (
	"errors"
	"os"
)

type Config struct {
	DiscordWebhookURL    string
	InstatusWebhookSecret string
	ListenAddr            string
	StatusPageName       string
}

func loadConfig() (Config, error) {
	cfg := Config{
		DiscordWebhookURL:     os.Getenv("DISCORD_WEBHOOK_URL"),
		InstatusWebhookSecret: os.Getenv("INSTATUS_WEBHOOK_SECRET"),
		ListenAddr:            os.Getenv("LISTEN_ADDR"),
		StatusPageName:        os.Getenv("STATUS_PAGE_NAME"),
	}

	if cfg.ListenAddr == "" {
		cfg.ListenAddr = "127.0.0.1:8080"
	}

	if cfg.StatusPageName == "" {
		cfg.StatusPageName = "BSky Status"
	}

	if cfg.DiscordWebhookURL == "" {
		return Config{}, errors.New("DISCORD_WEBHOOK_URL is not configured")
	}

	if cfg.InstatusWebhookSecret == "" {
		return Config{}, errors.New("INSTATUS_WEBHOOK_SECRET is not configured")
	}

	return cfg, nil
}
