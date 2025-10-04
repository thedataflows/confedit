package utils

import (
	"os"
	"strings"
)

// Color codes for terminal output
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
	Bold   = "\033[1m"
)

// ColorSupport represents terminal color support
type ColorSupport struct {
	enabled bool
}

// NewColorSupport creates a new ColorSupport instance with terminal detection
func NewColorSupport() *ColorSupport {
	return &ColorSupport{
		enabled: supportsColor(),
	}
}

// supportsColor detects if the terminal supports color output
func supportsColor() bool {
	// Check for FORCE_COLOR environment variable (enables color) - highest priority
	if os.Getenv("FORCE_COLOR") != "" {
		return true
	}

	// Check for NO_COLOR environment variable (disables color) - second priority
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if stdout is a terminal
	if !isTerminal() {
		return false
	}

	// Check environment variables that indicate color support
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}

	// Check for common terminals that support color
	colorTerms := []string{
		"xterm", "xterm-256color", "xterm-color",
		"screen", "screen-256color",
		"tmux", "tmux-256color",
		"rxvt", "rxvt-unicode",
		"linux", "cygwin",
		"alacritty", "kitty", "iterm",
	}

	termLower := strings.ToLower(term)
	for _, colorTerm := range colorTerms {
		if strings.Contains(termLower, colorTerm) {
			return true
		}
	}

	// Check for color-related environment variables
	if os.Getenv("COLORTERM") != "" {
		return true
	}

	return false
}

// isTerminal checks if stdout is connected to a terminal
func isTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// Colorize applies color to text if color support is enabled
func (cs *ColorSupport) Colorize(text string, color string) string {
	if !cs.enabled {
		return text
	}
	return color + text + Reset
}

// Red colorizes text in red
func (cs *ColorSupport) Red(text string) string {
	return cs.Colorize(text, Red)
}

// Green colorizes text in green
func (cs *ColorSupport) Green(text string) string {
	return cs.Colorize(text, Green)
}

// Yellow colorizes text in yellow
func (cs *ColorSupport) Yellow(text string) string {
	return cs.Colorize(text, Yellow)
}

// Blue colorizes text in blue
func (cs *ColorSupport) Blue(text string) string {
	return cs.Colorize(text, Blue)
}

// Bold makes text bold
func (cs *ColorSupport) Bold(text string) string {
	return cs.Colorize(text, Bold)
}

// IsEnabled returns whether color support is enabled
func (cs *ColorSupport) IsEnabled() bool {
	return cs.enabled
}

// ForceEnable forces color support on (for testing)
func (cs *ColorSupport) ForceEnable() {
	cs.enabled = true
}

// ForceDisable forces color support off (for testing)
func (cs *ColorSupport) ForceDisable() {
	cs.enabled = false
}
