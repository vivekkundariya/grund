package ui

import (
	"fmt"
	"time"
)

// Spinner provides a simple spinner for progress indication
type Spinner struct {
	message string
	done    bool
}

// NewSpinner creates a new spinner
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		done:    false,
	}
}

// Start starts the spinner (simplified version)
func (s *Spinner) Start() {
	// TODO: Implement actual spinner using charmbracelet/bubbles
	fmt.Printf("%s...", s.message)
}

// Stop stops the spinner
func (s *Spinner) Stop() {
	if !s.done {
		s.done = true
		fmt.Println(" ✓")
	}
}

// Update updates the spinner message
func (s *Spinner) Update(message string) {
	s.message = message
}

// Simple spinner implementation
func ShowSpinner(message string, fn func() error) error {
	spinner := NewSpinner(message)
	spinner.Start()

	err := fn()

	spinner.Stop()
	return err
}

// ShowProgress shows a progress message
func ShowProgress(message string) {
	fmt.Printf("  %s\n", message)
}

// ShowSuccess shows a success message
func ShowSuccess(message string) {
	fmt.Printf("  %s ✓\n", message)
}

// ShowError shows an error message
func ShowError(message string) {
	fmt.Printf("  %s ✗\n", message)
}
