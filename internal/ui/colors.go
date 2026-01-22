package ui

// Terminal color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

// Colorize wraps text with a color code
func Colorize(text string, color string) string {
	return color + text + Reset
}

// Success returns green text
func Success(text string) string {
	return Colorize(text, Green)
}

// Error returns red text
func Error(text string) string {
	return Colorize(text, Red)
}

// Warning returns yellow text
func Warning(text string) string {
	return Colorize(text, Yellow)
}

// Info returns blue text
func Info(text string) string {
	return Colorize(text, Blue)
}
