package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelSuccess
	LevelWarn
	LevelError
)

// Logger provides structured, colored logging
type Logger struct {
	mu       sync.Mutex
	out      io.Writer
	verbose  bool
	useColor bool
}

var (
	defaultLogger = &Logger{
		out:      os.Stdout,
		verbose:  false,
		useColor: true,
	}
)

// SetVerbose enables or disables verbose (debug) output
func SetVerbose(v bool) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.verbose = v
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	return defaultLogger.verbose
}

// SetOutput sets the output writer
func SetOutput(w io.Writer) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.out = w
}

// SetColor enables or disables colored output
func SetColor(c bool) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.useColor = c
}

// levelPrefix returns the colored prefix for a log level
func levelPrefix(level LogLevel, useColor bool) string {
	if !useColor {
		switch level {
		case LevelDebug:
			return "[DEBUG]"
		case LevelInfo:
			return "[INFO]"
		case LevelSuccess:
			return "[OK]"
		case LevelWarn:
			return "[WARN]"
		case LevelError:
			return "[ERROR]"
		default:
			return "[LOG]"
		}
	}

	switch level {
	case LevelDebug:
		return Colorize("[DEBUG]", Cyan)
	case LevelInfo:
		return Colorize("[INFO]", Blue)
	case LevelSuccess:
		return Colorize("[OK]", Green)
	case LevelWarn:
		return Colorize("[WARN]", Yellow)
	case LevelError:
		return Colorize("[ERROR]", Red)
	default:
		return "[LOG]"
	}
}

// log writes a log message with the given level
func log(level LogLevel, format string, args ...any) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	// Skip debug messages unless verbose is enabled
	if level == LevelDebug && !defaultLogger.verbose {
		return
	}

	prefix := levelPrefix(level, defaultLogger.useColor)
	timestamp := time.Now().Format("15:04:05")

	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	if defaultLogger.useColor {
		timestamp = Colorize(timestamp, Purple)
	}

	fmt.Fprintf(defaultLogger.out, "%s %s %s\n", timestamp, prefix, msg)
}

// Debug logs a debug message (only shown with --verbose)
func Debug(format string, args ...any) {
	log(LevelDebug, format, args...)
}

// Infof logs an informational message
func Infof(format string, args ...any) {
	log(LevelInfo, format, args...)
}

// Successf logs a success message
func Successf(format string, args ...any) {
	log(LevelSuccess, format, args...)
}

// Warnf logs a warning message
func Warnf(format string, args ...any) {
	log(LevelWarn, format, args...)
}

// Errorf logs an error message
func Errorf(format string, args ...any) {
	log(LevelError, format, args...)
}

// Step logs a step in a process with an arrow prefix
func Step(format string, args ...any) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	arrow := "→"
	if defaultLogger.useColor {
		arrow = Colorize("→", Cyan)
	}

	fmt.Fprintf(defaultLogger.out, "  %s %s\n", arrow, msg)
}

// SubStep logs a sub-step with indentation
func SubStep(format string, args ...any) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	bullet := "•"
	if defaultLogger.useColor {
		bullet = Colorize("•", Purple)
	}

	fmt.Fprintf(defaultLogger.out, "    %s %s\n", bullet, msg)
}

// Header logs a section header
func Header(format string, args ...any) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	if defaultLogger.useColor {
		msg = Colorize(msg, White) + Colorize(strings.Repeat("─", 40), Purple)
	}

	fmt.Fprintf(defaultLogger.out, "\n%s\n", msg)
}

// Progress logs a progress message that can be updated
func Progress(format string, args ...any) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}

	spinner := "⣾"
	if defaultLogger.useColor {
		spinner = Colorize(spinner, Cyan)
	}

	fmt.Fprintf(defaultLogger.out, "  %s %s\n", spinner, msg)
}

// ServiceStatus logs a service status line
func ServiceStatus(name, status string, healthy bool) {
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()

	statusIcon := "●"
	if defaultLogger.useColor {
		if healthy {
			statusIcon = Colorize("●", Green)
			status = Colorize(status, Green)
		} else {
			statusIcon = Colorize("●", Yellow)
			status = Colorize(status, Yellow)
		}
		name = Colorize(name, White)
	}

	fmt.Fprintf(defaultLogger.out, "  %s %-20s %s\n", statusIcon, name, status)
}
