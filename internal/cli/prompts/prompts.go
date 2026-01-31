package prompts

import (
	"strconv"

	"github.com/charmbracelet/huh"
)

// Text prompts for text input with an optional default value
func Text(title string, defaultVal string) (string, error) {
	var value string
	if defaultVal != "" {
		value = defaultVal
	}

	err := huh.NewInput().
		Title(title).
		Value(&value).
		Run()

	if err != nil {
		return defaultVal, err
	}

	if value == "" {
		return defaultVal, nil
	}

	return value, nil
}

// Int prompts for integer input with a default value
func Int(title string, defaultVal int) (int, error) {
	value := strconv.Itoa(defaultVal)

	err := huh.NewInput().
		Title(title).
		Value(&value).
		Validate(func(s string) error {
			_, err := strconv.Atoi(s)
			return err
		}).
		Run()

	if err != nil {
		return defaultVal, err
	}

	if value == "" {
		return defaultVal, nil
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultVal, nil
	}

	return result, nil
}

// Confirm prompts for yes/no confirmation
func Confirm(title string, defaultVal bool) (bool, error) {
	value := defaultVal

	err := huh.NewConfirm().
		Title(title).
		Value(&value).
		Run()

	if err != nil {
		return defaultVal, err
	}

	return value, nil
}

// Select prompts user to select from a list of options
func Select(title string, options []string, defaultVal string) (string, error) {
	var value string

	// Build options
	opts := make([]huh.Option[string], len(options))
	for i, opt := range options {
		opts[i] = huh.NewOption(opt, opt)
		if opt == defaultVal {
			value = opt
		}
	}

	// If no match found, use first option
	if value == "" && len(options) > 0 {
		value = options[0]
	}

	err := huh.NewSelect[string]().
		Title(title).
		Options(opts...).
		Value(&value).
		Run()

	if err != nil {
		return defaultVal, err
	}

	return value, nil
}

// MultiSelect prompts user to select multiple options
// Use Space to toggle, Enter to confirm
func MultiSelect(title string, options []string) ([]string, error) {
	var values []string

	// Build options
	opts := make([]huh.Option[string], len(options))
	for i, opt := range options {
		opts[i] = huh.NewOption(opt, opt)
	}

	err := huh.NewMultiSelect[string]().
		Title(title).
		Description("Use ↑/↓ to navigate, Space to select, Enter to confirm").
		Options(opts...).
		Value(&values).
		Run()

	if err != nil {
		return nil, err
	}

	return values, nil
}
