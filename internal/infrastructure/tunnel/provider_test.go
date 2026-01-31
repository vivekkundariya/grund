package tunnel

import (
	"testing"
)

func TestTunnelStructure(t *testing.T) {
	tunnel := &Tunnel{
		Name:      "localstack",
		PublicURL: "https://abc.trycloudflare.com",
		LocalAddr: "localhost:4566",
	}

	if tunnel.Name != "localstack" {
		t.Errorf("expected name localstack, got %s", tunnel.Name)
	}
	if tunnel.PublicURL != "https://abc.trycloudflare.com" {
		t.Errorf("expected URL https://abc.trycloudflare.com, got %s", tunnel.PublicURL)
	}
}

func TestProviderConstants(t *testing.T) {
	if ProviderCloudflared != "cloudflared" {
		t.Errorf("expected cloudflared constant")
	}
	if ProviderNgrok != "ngrok" {
		t.Errorf("expected ngrok constant")
	}
}
