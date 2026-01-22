package executor

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HealthChecker checks if a service is healthy
type HealthChecker struct {
	Endpoint string
	Interval time.Duration
	Timeout  time.Duration
	Retries  int
}

// CheckHealth polls a health endpoint until it's healthy or retries are exhausted
func (hc *HealthChecker) CheckHealth(ctx context.Context) error {
	client := &http.Client{
		Timeout: hc.Timeout,
	}

	for i := 0; i < hc.Retries; i++ {
		req, err := http.NewRequestWithContext(ctx, "GET", hc.Endpoint, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(hc.Interval):
		}
	}

	return fmt.Errorf("health check failed after %d retries", hc.Retries)
}

// WaitForHealthy waits for a service to become healthy
func WaitForHealthy(endpoint string, interval, timeout time.Duration, retries int) error {
	checker := &HealthChecker{
		Endpoint: endpoint,
		Interval: interval,
		Timeout:  timeout,
		Retries:  retries,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(retries)*interval+timeout)
	defer cancel()

	return checker.CheckHealth(ctx)
}
