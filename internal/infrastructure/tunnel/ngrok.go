package tunnel

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ngrokLogEntry represents a JSON log entry from ngrok
type ngrokLogEntry struct {
	URL string `json:"url"`
	Msg string `json:"msg"`
}

// NgrokProvider implements Provider for ngrok tunnels
type NgrokProvider struct{}

// NewNgrokProvider creates a new ngrok provider
func NewNgrokProvider() *NgrokProvider {
	return &NgrokProvider{}
}

// Name returns the provider name
func (p *NgrokProvider) Name() string {
	return ProviderNgrok
}

// Start creates a tunnel using ngrok
func (p *NgrokProvider) Start(ctx context.Context, name string, localAddr string) (*Tunnel, error) {
	// Check if ngrok is installed
	if _, err := exec.LookPath("ngrok"); err != nil {
		return nil, fmt.Errorf("ngrok not found in PATH: install from https://ngrok.com/download: %w", err)
	}

	// Extract port from localAddr
	parts := strings.Split(localAddr, ":")
	port := parts[len(parts)-1]

	// Start ngrok with JSON logging
	cmd := exec.CommandContext(ctx, "ngrok", "http", port, "--log", "stdout", "--log-format", "json")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start ngrok: %w", err)
	}

	// Wait for URL with timeout
	urlChan := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			var entry ngrokLogEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				continue
			}
			if entry.URL != "" && strings.HasPrefix(entry.URL, "https://") {
				urlChan <- entry.URL
				return
			}
		}
	}()

	select {
	case url := <-urlChan:
		return &Tunnel{
			Name:      name,
			PublicURL: url,
			LocalAddr: localAddr,
			Process:   cmd.Process,
		}, nil
	case <-time.After(30 * time.Second):
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("timeout waiting for ngrok URL after 30 seconds")
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("context cancelled while waiting for ngrok URL: %w", ctx.Err())
	}
}

// Stop terminates the tunnel
func (p *NgrokProvider) Stop(tunnel *Tunnel) error {
	if tunnel == nil || tunnel.Process == nil {
		return nil
	}
	if err := tunnel.Process.Kill(); err != nil {
		return fmt.Errorf("failed to stop ngrok tunnel %s: %w", tunnel.Name, err)
	}
	return nil
}
