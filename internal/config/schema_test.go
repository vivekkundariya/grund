package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTunnelConfigParsing(t *testing.T) {
	yamlContent := `
version: "1"
service:
  name: test-service
  type: go
  port: 8080
  health:
    endpoint: /health
    interval: 5s
    timeout: 3s
    retries: 10
requires:
  infrastructure:
    tunnel:
      provider: cloudflared
      targets:
        - name: localstack
          host: localhost
          port: "4566"
        - name: api
          host: "${self.host}"
          port: "${self.port}"
`
	var config ServiceConfig
	err := yaml.Unmarshal([]byte(yamlContent), &config)
	if err != nil {
		t.Fatalf("failed to parse yaml: %v", err)
	}

	if config.Requires.Infrastructure.Tunnel == nil {
		t.Fatal("expected tunnel config to be parsed")
	}
	if config.Requires.Infrastructure.Tunnel.Provider != "cloudflared" {
		t.Errorf("expected provider cloudflared, got %s", config.Requires.Infrastructure.Tunnel.Provider)
	}
	if len(config.Requires.Infrastructure.Tunnel.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(config.Requires.Infrastructure.Tunnel.Targets))
	}
	if config.Requires.Infrastructure.Tunnel.Targets[0].Name != "localstack" {
		t.Errorf("expected target name localstack, got %s", config.Requires.Infrastructure.Tunnel.Targets[0].Name)
	}
	if config.Requires.Infrastructure.Tunnel.Targets[0].Port != "4566" {
		t.Errorf("expected port 4566, got %s", config.Requires.Infrastructure.Tunnel.Targets[0].Port)
	}
}
