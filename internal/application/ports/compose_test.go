package ports

import "testing"

func TestEnvironmentContextHasTunnel(t *testing.T) {
	ctx := NewDefaultEnvironmentContext()

	// Tunnel map should exist
	if ctx.Tunnel == nil {
		t.Fatal("expected Tunnel map to be initialized")
	}

	// Should be able to add tunnel context
	ctx.Tunnel["localstack"] = TunnelContext{
		Name:      "localstack",
		PublicURL: "https://abc.trycloudflare.com",
	}

	if ctx.Tunnel["localstack"].PublicURL != "https://abc.trycloudflare.com" {
		t.Error("expected tunnel context to be stored")
	}
}
