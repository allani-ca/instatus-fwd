package main

import "testing"

func TestDiscordColor(t *testing.T) {
	tests := []struct {
		eventType string
		status    string
		want      int
	}{
		{"incident", "RESOLVED", 0x57F287},
		{"incident", "MONITORING", 0xFEE75C},
		{"incident", "IDENTIFIED", 0xFAA61A},
		{"incident", "INVESTIGATING", 0xED4245},
		{"maintenance", "COMPLETED", 0x57F287},
		{"maintenance", "INPROGRESS", 0xFEE75C},
		{"maintenance", "NOTSTARTEDYET", 0x5865F2},
		{"component", "OPERATIONAL", 0x57F287},
		{"component", "DEGRADEDPERFORMANCE", 0xFEE75C},
		{"component", "PARTIALOUTAGE", 0xFAA61A},
		{"component", "MAJOROUTAGE", 0xED4245},
		{"component", "UNDERMAINTENANCE", 0x5865F2},
		{"component", "UNKNOWN", 0x99AAB5},
		{"unknown", "UNKNOWN", 0x5865F2},
	}

	for _, test := range tests {
		t.Run(test.eventType+"/"+test.status, func(t *testing.T) {
			if got := discordColor(test.eventType, test.status); got != test.want {
				t.Errorf("discordColor(%q, %q) = %#x, want %#x", test.eventType, test.status, got, test.want)
			}
		})
	}
}
