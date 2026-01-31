package tunnel

import (
	"encoding/json"
	"testing"
)

func TestNgrokURLParsing(t *testing.T) {
	// Sample ngrok JSON log line
	logLine := `{"lvl":"info","msg":"started tunnel","obj":"tunnels","name":"command_line","addr":"http://localhost:4566","url":"https://abc123.ngrok-free.app"}`

	var logEntry struct {
		URL string `json:"url"`
	}

	err := json.Unmarshal([]byte(logLine), &logEntry)
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if logEntry.URL != "https://abc123.ngrok-free.app" {
		t.Errorf("expected https://abc123.ngrok-free.app, got %s", logEntry.URL)
	}
}

func TestNgrokProviderName(t *testing.T) {
	provider := NewNgrokProvider()
	if provider.Name() != ProviderNgrok {
		t.Errorf("expected %s, got %s", ProviderNgrok, provider.Name())
	}
}
