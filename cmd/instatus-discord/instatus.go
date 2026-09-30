package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
)

// verifyInstatusSignature compares the supplied signature with the body's HMAC.
func verifyInstatusSignature(body []byte, supplied, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)

	expected := hex.EncodeToString(mac.Sum(nil))

	// hmac.Equal prevents ordinary string comparison from becoming a
	// timing side-channel.
	return hmac.Equal([]byte(expected), []byte(supplied))
}

// Payload is an Instatus webhook event and its associated page and event data.
type Payload struct {
	Page *Page `json:"page"`

	Incident *Incident `json:"incident"`

	Maintenance *Maintenance `json:"maintenance"`

	Component *Component `json:"component"`

	ComponentUpdate *ComponentUpdate `json:"component_update"`
}

// Page describes the status page that emitted an event.
type Page struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

// Incident contains the current incident details and updates.
type Incident struct {
	Name            string           `json:"name"`
	Status          string           `json:"status"`
	Impact          string           `json:"impact"`
	URL             string           `json:"url"`
	IncidentUpdates []IncidentUpdate `json:"incident_updates"`
	UpdatedAt       string           `json:"updated_at"`
	CreatedAt       string           `json:"created_at"`
}

// IncidentUpdate contains a message and creation time for an incident update.
type IncidentUpdate struct {
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

// Maintenance contains the current scheduled maintenance details and updates.
type Maintenance struct {
	Name               string              `json:"name"`
	Status             string              `json:"status"`
	URL                string              `json:"url"`
	MaintenanceUpdates []MaintenanceUpdate `json:"maintenance_updates"`
	UpdatedAt          string              `json:"updated_at"`
	CreatedAt          string              `json:"created_at"`
}

// MaintenanceUpdate contains a message and creation time for a maintenance update.
type MaintenanceUpdate struct {
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

// Component contains the name and current status of a page component.
type Component struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// ComponentUpdate describes a change to a component's status.
type ComponentUpdate struct {
	NewStatus string `json:"new_status"`
	CreatedAt string `json:"created_at"`
}

// decodePayload parses a webhook request body as an Instatus event.
func decodePayload(body []byte) (Payload, error) {
	var payload Payload

	if err := json.Unmarshal(body, &payload); err != nil {
		return Payload{}, fmt.Errorf("decode Instatus payload: %w", err)
	}

	return payload, nil
}

// requirePOST allows only POST requests and writes the appropriate error otherwise.
func requirePOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}

	w.Header().Set("Allow", http.MethodPost)
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	return false
}
