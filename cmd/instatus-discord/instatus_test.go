package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestVerifyInstatusSignature(t *testing.T) {
	tests := []struct {
		name      string
		body      []byte
		supplied  string
		wantValid bool
	}{
		{name: "valid signature", body: []byte(`{"event":"incident"}`), supplied: signatureForTest([]byte(`{"event":"incident"}`), "secret"), wantValid: true},
		{name: "invalid HMAC", body: []byte(`{"event":"incident"}`), supplied: "invalid", wantValid: false},
		{name: "empty body with valid HMAC", body: []byte{}, supplied: signatureForTest(nil, "secret"), wantValid: true},
		{name: "empty body with invalid HMAC", body: []byte{}, supplied: "invalid", wantValid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if valid := verifyInstatusSignature(test.body, test.supplied, "secret"); valid != test.wantValid {
				t.Fatalf("verifyInstatusSignature() = %v, want %v", valid, test.wantValid)
			}
		})
	}
}

func TestDecodePayloadPreservesJSONError(t *testing.T) {
	_, err := decodePayload([]byte(`{"incident":`))
	if err == nil || !strings.Contains(err.Error(), "decode Instatus payload") {
		t.Fatalf("decodePayload() error = %v, want wrapped decode error", err)
	}
	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Fatalf("decodePayload() error %T does not wrap json.SyntaxError", err)
	}
}

func signatureForTest(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
