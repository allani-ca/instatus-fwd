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

// newHandler constructs the HTTP handler that validates and forwards webhooks.
func newHandler(cfg Config) http.Handler {
	client := &http.Client{
		Timeout: cfg.DiscordTimeout,
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
			log.Printf("Malformed Instatus payload: %v", err)
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

		ctx, cancel := context.WithTimeout(r.Context(), cfg.DiscordTimeout)
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

// buildEmbed converts a supported Instatus event into a Discord embed.
func buildEmbed(payload Payload, footerName string) (DiscordEmbed, bool) {
	pageURL := ""
	if payload.Page != nil {
		pageURL = payload.Page.URL
	}

	if payload.Incident != nil {
		incident := payload.Incident
		status := incident.Status
		description := "There is an update to this incident."
		if len(incident.IncidentUpdates) > 0 {
			if body := incident.IncidentUpdates[0].Body; body != "" {
				description = truncate(body, 2000)
			}
		}

		return createEmbed(
			statusEmoji(status)+" "+truncate(defaultString(incident.Name, "Service Incident"), 256),
			description,
			"incident",
			status,
			footerName,
			firstNonEmpty(incident.URL, pageURL),
			createCommonFields(status, incident.Impact),
			incident.UpdatedAt,
			incident.CreatedAt,
		), true
	}

	if payload.Maintenance != nil {
		maintenance := payload.Maintenance
		status := maintenance.Status
		description := "There is an update to scheduled maintenance."
		if len(maintenance.MaintenanceUpdates) > 0 {
			if body := maintenance.MaintenanceUpdates[0].Body; body != "" {
				description = truncate(body, 2000)
			}
		}

		return createEmbed(
			"🔧 "+truncate(defaultString(maintenance.Name, "Scheduled Maintenance"), 256),
			description,
			"maintenance",
			status,
			footerName,
			firstNonEmpty(maintenance.URL, pageURL),
			createCommonFields(status, ""),
			maintenance.UpdatedAt,
			maintenance.CreatedAt,
		), true
	}

	if payload.Component != nil && payload.ComponentUpdate != nil {
		component := payload.Component
		update := payload.ComponentUpdate
		status := firstNonEmpty(update.NewStatus, component.Status)
		return createEmbed(
			statusEmoji(status)+" "+truncate(defaultString(component.Name, "Component"), 256),
			"Component status changed to **"+prettyStatus(status)+"**.",
			"component",
			status,
			footerName,
			pageURL,
			createCommonFields(status, ""),
			update.CreatedAt,
			component.CreatedAt,
		), true
	}

	return DiscordEmbed{}, false
}

// createEmbed fills the fields shared by each supported event embed.
func createEmbed(title, description, eventType, status, footerName, pageURL string, fields []DiscordField, timestamps ...string) DiscordEmbed {
	return DiscordEmbed{
		Title:       title,
		Description: description,
		URL:         pageURL,
		Color:       discordColor(eventType, status),
		Fields:      fields,
		Footer:      &DiscordFooter{Text: footerName},
		Timestamp:   formatTimestamp(timestamps...),
	}
}

// createCommonFields adds status and impact fields when values are available.
func createCommonFields(status, impact string) []DiscordField {
	fields := make([]DiscordField, 0, 2)
	if status != "" {
		fields = append(fields, DiscordField{
			Name:   "Status",
			Value:  statusEmoji(status) + " " + prettyStatus(status),
			Inline: true,
		})
	}
	if impact != "" {
		fields = append(fields, DiscordField{
			Name:   "Impact",
			Value:  prettyImpact(impact),
			Inline: true,
		})
	}
	return fields
}

// formatTimestamp selects the first event timestamp or the current UTC time.
func formatTimestamp(values ...string) string {
	if timestamp := firstNonEmpty(values...); timestamp != "" {
		return timestamp
	}
	return time.Now().UTC().Format(time.RFC3339)
}

// statusEmoji returns the indicator associated with an Instatus status.
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

// prettyStatus converts an Instatus status code into readable text.
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

// prettyImpact converts an impact code into readable text.
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

// titleWords capitalizes the first character of each whitespace-separated word.
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

// truncate limits a string to max runes, adding an ellipsis when needed.
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

// defaultString returns fallback when value is empty.
func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}

	return value
}

// firstNonEmpty returns the first non-empty string, or an empty string.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}
