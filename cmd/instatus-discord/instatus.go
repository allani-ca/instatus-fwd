package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
)

func verifyInstatusSignature(body []byte, supplied, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)

	expected := hex.EncodeToString(mac.Sum(nil))

	// hmac.Equal prevents ordinary string comparison from becoming a
	// timing side-channel.
	return hmac.Equal([]byte(expected), []byte(supplied))
}

type Payload struct {
	Page *Page `json:"page"`

	Incident *Incident `json:"incident"`

	Maintenance *Maintenance `json:"maintenance"`

	Component *Component `json:"component"`

	ComponentUpdate *ComponentUpdate `json:"component_update"`
}

type Page struct {
	URL  string `json:"url"`
	Name string `json:"name"`
}

type Incident struct {
	Name           string          `json:"name"`
	Status         string          `json:"status"`
	Impact         string          `json:"impact"`
	URL            string          `json:"url"`
	IncidentUpdates []IncidentUpdate `json:"incident_updates"`
	UpdatedAt      string          `json:"updated_at"`
	CreatedAt      string          `json:"created_at"`
}

type IncidentUpdate struct {
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type Maintenance struct {
	Name               string              `json:"name"`
	Status             string              `json:"status"`
	URL                string              `json:"url"`
	MaintenanceUpdates []MaintenanceUpdate `json:"maintenance_updates"`
	UpdatedAt          string              `json:"updated_at"`
	CreatedAt          string              `json:"created_at"`
}

type MaintenanceUpdate struct {
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type Component struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type ComponentUpdate struct {
	NewStatus string `json:"new_status"`
	CreatedAt string `json:"created_at"`
}

func decodePayload(body []byte) (Payload, error) {
	var payload Payload

	if err := json.Unmarshal(body, &payload); err != nil {
		return Payload{}, errors.New("invalid JSON")
	}

	return payload, nil
}

func requirePOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}

	w.Header().Set("Allow", http.MethodPost)
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	return false
}
