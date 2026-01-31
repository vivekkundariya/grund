package tunnel

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"time"
)

// Compile-time interface check
var _ Provider = (*CloudflaredProvider)(nil)

var cloudflaredURLPattern = regexp.MustCompile(`(https://[a-z0-9-]+\.trycloudflare\.com)`)

// CloudflaredProvider implements Provider for cloudflared tunnels
type CloudflaredProvider struct{}

// NewCloudflaredProvider creates a new cloudflared provider
func NewCloudflaredProvider() *CloudflaredProvider {
	return &CloudflaredProvider{}
}

// Name returns the provider name
func (p *CloudflaredProvider) Name() string {
	return ProviderCloudflared
}

// Start creates a tunnel using cloudflared
func (p *CloudflaredProvider) Start(ctx context.Context, name string, localAddr string) (*Tunnel, error) {
	// Check if cloudflared is installed
	if _, err := exec.LookPath("cloudflared"); err != nil {
		return nil, fmt.Errorf("cloudflared not found in PATH: install with 'brew install cloudflared': %w", err)
	}

	// Start cloudflared tunnel
	cmd := exec.CommandContext(ctx, "cloudflared", "tunnel", "--url", fmt.Sprintf("http://%s", localAddr))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start cloudflared: %w", err)
	}

	// Wait for URL with timeout
	urlChan := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if matches := cloudflaredURLPattern.FindStringSubmatch(line); len(matches) > 1 {
				urlChan <- matches[1]
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
		return nil, fmt.Errorf("timeout waiting for cloudflared URL after 30 seconds")
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		return nil, fmt.Errorf("context cancelled while waiting for cloudflared URL: %w", ctx.Err())
	}
}

// Stop terminates the tunnel
func (p *CloudflaredProvider) Stop(tunnel *Tunnel) error {
	if tunnel == nil || tunnel.Process == nil {
		return nil
	}
	if err := tunnel.Process.Kill(); err != nil {
		return fmt.Errorf("failed to stop cloudflared tunnel %s: %w", tunnel.Name, err)
	}
	return nil
}
