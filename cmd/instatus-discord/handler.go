package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	maxBodySize = 1 << 20 // 1 MiB
)

func newHandler(cfg Config) http.Handler {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !requirePOST(w, r) {
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, maxBodySize+1))
		if err != nil {
			http.Error(w, "Unable to read request body", http.StatusBadRequest)
			return
		}

		if len(body) == 0 {
			http.Error(w, "Empty request body", http.StatusBadRequest)
			return
		}

		if len(body) > maxBodySize {
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}

		signature := r.Header.Get("X-Instatus-Webhook-Signature")
		if signature == "" {
			http.Error(w, "Missing Instatus signature", http.StatusBadRequest)
			return
		}

		if !verifyInstatusSignature(body, signature, cfg.InstatusWebhookSecret) {
			http.Error(w, "Invalid Instatus signature", http.StatusUnauthorized)
			return
		}

		payload, err := decodePayload(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		embed, ok := buildEmbed(payload, cfg.StatusPageName)
		if !ok {
			http.Error(w, "Unknown Instatus payload", http.StatusBadRequest)
			return
		}

		discordPayload := DiscordPayload{
			Username: cfg.StatusPageName,
			Embeds:   []DiscordEmbed{embed},
			AllowedMentions: AllowedMentions{
				Parse: []string{},
			},
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		if err := sendToDiscord(ctx, client, cfg.DiscordWebhookURL, discordPayload); err != nil {
			log.Printf("Discord delivery failed: %v", err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
}

func buildEmbed(payload Payload, footerName string) (DiscordEmbed, bool) {
	pageURL := ""
	if payload.Page != nil {
		pageURL = payload.Page.URL
	}

	// Incident
	if payload.Incident != nil {
		incident := payload.Incident

		status := incident.Status
		description := "There is an update to this incident."

		if len(incident.IncidentUpdates) > 0 {
			if body := incident.IncidentUpdates[0].Body; body != "" {
				description = truncate(body, 2000)
			}
		}

		fields := make([]DiscordField, 0, 2)

		if status != "" {
			fields = append(fields, DiscordField{
				Name:   "Status",
				Value:  statusEmoji(status) + " " + prettyStatus(status),
				Inline: true,
			})
		}

		if incident.Impact != "" {
			fields = append(fields, DiscordField{
				Name:   "Impact",
				Value:  prettyImpact(incident.Impact),
				Inline: true,
			})
		}

		embed := DiscordEmbed{
			Title:       statusEmoji(status) + " " + truncate(defaultString(incident.Name, "Service Incident"), 256),
			Description: description,
			Color:       discordColor("incident", status),
			Fields:      fields,
			Footer: &DiscordFooter{
				Text: footerName,
			},
			Timestamp: firstNonEmpty(
				incident.UpdatedAt,
				incident.CreatedAt,
				time.Now().UTC().Format(time.RFC3339),
			),
		}

		embed.URL = firstNonEmpty(incident.URL, pageURL)

		return embed, true
	}

	// Maintenance
	if payload.Maintenance != nil {
		maintenance := payload.Maintenance

		status := maintenance.Status
		description := "There is an update to scheduled maintenance."

		if len(maintenance.MaintenanceUpdates) > 0 {
			if body := maintenance.MaintenanceUpdates[0].Body; body != "" {
				description = truncate(body, 2000)
			}
		}

		fields := make([]DiscordField, 0, 1)

		if status != "" {
			fields = append(fields, DiscordField{
				Name:   "Status",
				Value:  statusEmoji(status) + " " + prettyStatus(status),
				Inline: true,
			})
		}

		embed := DiscordEmbed{
			Title:       "🔧 " + truncate(defaultString(maintenance.Name, "Scheduled Maintenance"), 256),
			Description: description,
			Color:       discordColor("maintenance", status),
			Fields:      fields,
			Footer: &DiscordFooter{
				Text: footerName,
			},
			Timestamp: firstNonEmpty(
				maintenance.UpdatedAt,
				maintenance.CreatedAt,
				time.Now().UTC().Format(time.RFC3339),
			),
		}

		embed.URL = firstNonEmpty(maintenance.URL, pageURL)

		return embed, true
	}

	// Component update
	if payload.Component != nil && payload.ComponentUpdate != nil {
		component := payload.Component
		update := payload.ComponentUpdate

		status := firstNonEmpty(update.NewStatus, component.Status)

		embed := DiscordEmbed{
			Title:       statusEmoji(status) + " " + truncate(defaultString(component.Name, "Component"), 256),
			Description: "Component status changed to **" + prettyStatus(status) + "**.",
			Color:       discordColor("component", status),
			Fields: []DiscordField{
				{
					Name:   "Status",
					Value:  statusEmoji(status) + " " + prettyStatus(status),
					Inline: true,
				},
			},
			Footer: &DiscordFooter{
				Text: footerName,
			},
			Timestamp: firstNonEmpty(
				update.CreatedAt,
				component.CreatedAt,
				time.Now().UTC().Format(time.RFC3339),
			),
			URL: pageURL,
		}

		return embed, true
	}

	return DiscordEmbed{}, false
}

func statusEmoji(status string) string {
	switch strings.ToUpper(status) {
	case "INVESTIGATING":
		return "🔴"
	case "IDENTIFIED":
		return "🟠"
	case "MONITORING":
		return "🟡"
	case "RESOLVED":
		return "🟢"
	case "NOTSTARTEDYET":
		return "🔵"
	case "INPROGRESS":
		return "🟡"
	case "COMPLETED":
		return "🟢"
	case "OPERATIONAL":
		return "🟢"
	case "DEGRADEDPERFORMANCE":
		return "🟡"
	case "PARTIALOUTAGE":
		return "🟠"
	case "MAJOROUTAGE":
		return "🔴"
	case "UNDERMAINTENANCE":
		return "🔵"
	default:
		return "ℹ️"
	}
}

func prettyStatus(status string) string {
	switch strings.ToUpper(status) {
	case "INVESTIGATING":
		return "Investigating"
	case "IDENTIFIED":
		return "Identified"
	case "MONITORING":
		return "Monitoring"
	case "RESOLVED":
		return "Resolved"
	case "NOTSTARTEDYET":
		return "Not started"
	case "INPROGRESS":
		return "In progress"
	case "COMPLETED":
		return "Completed"
	case "OPERATIONAL":
		return "Operational"
	case "DEGRADEDPERFORMANCE":
		return "Degraded performance"
	case "PARTIALOUTAGE":
		return "Partial outage"
	case "MAJOROUTAGE":
		return "Major outage"
	case "UNDERMAINTENANCE":
		return "Under maintenance"
	default:
		s := strings.ToLower(strings.ReplaceAll(status, "_", " "))
		return titleWords(s)
	}
}

func prettyImpact(impact string) string {
	switch strings.ToUpper(impact) {
	case "NONE":
		return "None"
	case "MINOR":
		return "Minor"
	case "MAJOR":
		return "Major"
	case "CRITICAL":
		return "Critical"
	default:
		return titleWords(strings.ToLower(impact))
	}
}

func titleWords(s string) string {
	parts := strings.Fields(s)

	for i, part := range parts {
		if len(part) == 0 {
			continue
		}

		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}

	return strings.Join(parts, " ")
}

func truncate(s string, max int) string {
	runes := []rune(s)

	if len(runes) <= max {
		return s
	}

	if max <= 3 {
		return string(runes[:max])
	}

	return string(runes[:max-3]) + "..."
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

