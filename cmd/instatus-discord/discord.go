package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type DiscordEmbed struct {
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	URL         string         `json:"url,omitempty"`
	Color       int            `json:"color,omitempty"`
	Fields      []DiscordField `json:"fields,omitempty"`
	Footer      *DiscordFooter `json:"footer,omitempty"`
	Timestamp   string         `json:"timestamp,omitempty"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type DiscordFooter struct {
	Text string `json:"text"`
}

type DiscordPayload struct {
	Username         string         `json:"username"`
	Embeds           []DiscordEmbed `json:"embeds"`
	AllowedMentions  AllowedMentions `json:"allowed_mentions"`
}

type AllowedMentions struct {
	Parse []string `json:"parse"`
}

func sendToDiscord(ctx context.Context, client *http.Client, webhookURL string, payload DiscordPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal Discord payload: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		webhookURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("create Discord request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Discord request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Discord returned HTTP %d", resp.StatusCode)
	}

	return nil
}

func discordColor(eventType, status string) int {
	status = strings.ToUpper(status)

	switch eventType {
	case "incident":
		switch status {
		case "RESOLVED":
			return 0x57F287
		case "MONITORING":
			return 0xFEE75C
		case "IDENTIFIED":
			return 0xFAA61A
		default:
			return 0xED4245
		}

	case "maintenance":
		switch status {
		case "COMPLETED":
			return 0x57F287
		case "INPROGRESS":
			return 0xFEE75C
		default:
			return 0x5865F2
		}

	case "component":
		switch status {
		case "OPERATIONAL":
			return 0x57F287
		case "DEGRADEDPERFORMANCE":
			return 0xFEE75C
		case "PARTIALOUTAGE":
			return 0xFAA61A
		case "MAJOROUTAGE":
			return 0xED4245
		case "UNDERMAINTENANCE":
			return 0x5865F2
		default:
			return 0x99AAB5
		}
	}

	return 0x5865F2
}
