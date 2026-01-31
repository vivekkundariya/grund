package tunnel

import (
	"regexp"
	"testing"
)

func TestCloudflaredURLParsing(t *testing.T) {
	// Sample cloudflared output lines
	testCases := []struct {
		line     string
		expected string
	}{
		{
			line:     "2024-01-15T10:00:00Z INF |  https://random-words-here.trycloudflare.com",
			expected: "https://random-words-here.trycloudflare.com",
		},
		{
			line:     "INF +--------------------------------------------------------------------------------------------+",
			expected: "",
		},
		{
			line:     "https://abc-def-ghi.trycloudflare.com",
			expected: "https://abc-def-ghi.trycloudflare.com",
		},
	}

	pattern := regexp.MustCompile(`(https://[a-z0-9-]+\.trycloudflare\.com)`)

	for _, tc := range testCases {
		matches := pattern.FindStringSubmatch(tc.line)
		var result string
		if len(matches) > 1 {
			result = matches[1]
		}
		if result != tc.expected {
			t.Errorf("line %q: expected %q, got %q", tc.line, tc.expected, result)
		}
	}
}

func TestCloudflaredProviderName(t *testing.T) {
	provider := NewCloudflaredProvider()
	if provider.Name() != ProviderCloudflared {
		t.Errorf("expected %s, got %s", ProviderCloudflared, provider.Name())
	}
}
