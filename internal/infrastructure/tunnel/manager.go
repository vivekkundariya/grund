package tunnel

import (
	"context"
	"fmt"
	"sync"

	"github.com/vivekkundariya/grund/internal/config"
)

// Manager handles tunnel lifecycle
type Manager struct {
	tunnels map[string]*Tunnel
	mu      sync.Mutex
}

// NewManager creates a new tunnel manager
func NewManager() *Manager {
	return &Manager{
		tunnels: make(map[string]*Tunnel),
	}
}

// GetProvider returns the appropriate provider for the given name
func (m *Manager) GetProvider(name string) (Provider, error) {
	switch name {
	case ProviderCloudflared:
		return NewCloudflaredProvider(), nil
	case ProviderNgrok:
		return NewNgrokProvider(), nil
	default:
		return nil, fmt.Errorf("unknown tunnel provider: %s (supported: cloudflared, ngrok)", name)
	}
}

// ValidateConfig validates tunnel configuration
func (m *Manager) ValidateConfig(cfg *config.TunnelConfig) error {
	if cfg == nil {
		return nil
	}

	// Validate provider
	if cfg.Provider != ProviderCloudflared && cfg.Provider != ProviderNgrok {
		return fmt.Errorf("invalid tunnel provider: %s (must be 'cloudflared' or 'ngrok')", cfg.Provider)
	}

	// Validate targets
	names := make(map[string]bool)
	for _, target := range cfg.Targets {
		if target.Name == "" {
			return fmt.Errorf("tunnel target missing name")
		}
		if target.Host == "" {
			return fmt.Errorf("tunnel target %s missing host", target.Name)
		}
		if target.Port == "" {
			return fmt.Errorf("tunnel target %s missing port", target.Name)
		}
		if names[target.Name] {
			return fmt.Errorf("duplicate tunnel target name: %s", target.Name)
		}
		names[target.Name] = true
	}

	return nil
}

// ResolvedTarget represents a target with placeholders resolved
type ResolvedTarget struct {
	Name string
	Host string
	Port string
}

// StartAll starts tunnels for all targets in the config
func (m *Manager) StartAll(ctx context.Context, cfg *config.TunnelConfig, resolvedTargets []ResolvedTarget) ([]*Tunnel, error) {
	if cfg == nil || len(cfg.Targets) == 0 {
		return nil, nil
	}

	provider, err := m.GetProvider(cfg.Provider)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	var tunnels []*Tunnel
	for _, target := range resolvedTargets {
		localAddr := fmt.Sprintf("%s:%s", target.Host, target.Port)
		tunnel, err := provider.Start(ctx, target.Name, localAddr)
		if err != nil {
			// Cleanup any started tunnels
			for _, t := range tunnels {
				_ = provider.Stop(t)
			}
			return nil, fmt.Errorf("failed to start tunnel %s: %w", target.Name, err)
		}
		tunnels = append(tunnels, tunnel)
		m.tunnels[target.Name] = tunnel
	}

	return tunnels, nil
}

// StopAll stops all running tunnels
func (m *Manager) StopAll() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, tunnel := range m.tunnels {
		if tunnel.Process != nil {
			if err := tunnel.Process.Kill(); err != nil {
				lastErr = fmt.Errorf("failed to stop tunnel %s: %w", name, err)
			}
		}
		delete(m.tunnels, name)
	}
	return lastErr
}

// GetTunnels returns all running tunnels
func (m *Manager) GetTunnels() map[string]*Tunnel {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]*Tunnel)
	for k, v := range m.tunnels {
		result[k] = v
	}
	return result
}
