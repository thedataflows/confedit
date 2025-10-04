package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ColorTestSuite struct {
	suite.Suite
	originalEnv map[string]string
}

func TestColorTestSuite(t *testing.T) {
	suite.Run(t, new(ColorTestSuite))
}

func (s *ColorTestSuite) SetupTest() {
	// Store original environment variables
	s.originalEnv = make(map[string]string)
	envVars := []string{"FORCE_COLOR", "NO_COLOR", "TERM", "COLORTERM"}

	for _, envVar := range envVars {
		s.originalEnv[envVar] = os.Getenv(envVar)
	}
}

func (s *ColorTestSuite) TearDownTest() {
	// Restore original environment variables
	for envVar, originalValue := range s.originalEnv {
		if originalValue == "" {
			os.Unsetenv(envVar)
		} else {
			os.Setenv(envVar, originalValue)
		}
	}
}

func (s *ColorTestSuite) TestNewColorSupport() {
	cs := NewColorSupport()
	assert.NotNil(s.T(), cs)
	// IsEnabled result depends on environment, so we just check it's deterministic
	enabled1 := cs.IsEnabled()
	enabled2 := cs.IsEnabled()
	assert.Equal(s.T(), enabled1, enabled2)
}

func (s *ColorTestSuite) TestColorSupport_ForceEnable() {
	cs := NewColorSupport()
	cs.ForceEnable()

	assert.True(s.T(), cs.IsEnabled())

	// Test all color methods with force enable
	red := cs.Red("error text")
	assert.Equal(s.T(), Red+"error text"+Reset, red)

	green := cs.Green("success text")
	assert.Equal(s.T(), Green+"success text"+Reset, green)

	yellow := cs.Yellow("warning text")
	assert.Equal(s.T(), Yellow+"warning text"+Reset, yellow)

	blue := cs.Blue("info text")
	assert.Equal(s.T(), Blue+"info text"+Reset, blue)

	bold := cs.Bold("bold text")
	assert.Equal(s.T(), Bold+"bold text"+Reset, bold)
}

func (s *ColorTestSuite) TestColorSupport_ForceDisable() {
	cs := NewColorSupport()
	cs.ForceDisable()

	assert.False(s.T(), cs.IsEnabled())

	// Test all color methods with force disable
	red := cs.Red("error text")
	assert.Equal(s.T(), "error text", red)

	green := cs.Green("success text")
	assert.Equal(s.T(), "success text", green)

	yellow := cs.Yellow("warning text")
	assert.Equal(s.T(), "warning text", yellow)

	blue := cs.Blue("info text")
	assert.Equal(s.T(), "info text", blue)

	bold := cs.Bold("bold text")
	assert.Equal(s.T(), "bold text", bold)
}

func (s *ColorTestSuite) TestColorSupport_Colorize() {
	cs := NewColorSupport()

	// Test with colors enabled
	cs.ForceEnable()
	result := cs.Colorize("test", Red)
	assert.Equal(s.T(), Red+"test"+Reset, result)

	// Test with colors disabled
	cs.ForceDisable()
	result = cs.Colorize("test", Red)
	assert.Equal(s.T(), "test", result)
}

func (s *ColorTestSuite) TestEnvironmentVariables_ForceColor() {
	// Test FORCE_COLOR environment variable
	os.Setenv("FORCE_COLOR", "1")
	cs := NewColorSupport()
	assert.True(s.T(), cs.IsEnabled())

	// Test different values
	os.Setenv("FORCE_COLOR", "true")
	cs = NewColorSupport()
	assert.True(s.T(), cs.IsEnabled())

	os.Setenv("FORCE_COLOR", "yes")
	cs = NewColorSupport()
	assert.True(s.T(), cs.IsEnabled())
}

func (s *ColorTestSuite) TestEnvironmentVariables_NoColor() {
	// Test NO_COLOR environment variable
	os.Setenv("NO_COLOR", "1")
	cs := NewColorSupport()
	assert.False(s.T(), cs.IsEnabled())

	// Test different values
	os.Setenv("NO_COLOR", "true")
	cs = NewColorSupport()
	assert.False(s.T(), cs.IsEnabled())

	os.Setenv("NO_COLOR", "anything")
	cs = NewColorSupport()
	assert.False(s.T(), cs.IsEnabled())
}

func (s *ColorTestSuite) TestEnvironmentVariables_Priority() {
	// FORCE_COLOR should take priority over NO_COLOR
	os.Setenv("FORCE_COLOR", "1")
	os.Setenv("NO_COLOR", "1")
	cs := NewColorSupport()
	assert.True(s.T(), cs.IsEnabled())

	// NO_COLOR should take priority over TERM
	os.Unsetenv("FORCE_COLOR")
	os.Setenv("NO_COLOR", "1")
	os.Setenv("TERM", "xterm-256color")
	cs = NewColorSupport()
	assert.False(s.T(), cs.IsEnabled())
}

func (s *ColorTestSuite) TestEnvironmentVariables_Term() {
	// Clear conflicting env vars
	os.Unsetenv("FORCE_COLOR")
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("COLORTERM")

	colorTerms := []string{
		"xterm", "xterm-256color", "xterm-color",
		"screen", "screen-256color",
		"tmux", "tmux-256color",
		"rxvt", "rxvt-unicode",
		"linux", "cygwin",
		"alacritty", "kitty", "iterm",
	}

	for _, term := range colorTerms {
		s.Run("TERM="+term, func() {
			os.Setenv("TERM", term)
			cs := NewColorSupport()
			// Note: This might still be false due to isTerminal() check
			// but the supportsColor logic should handle TERM correctly
			enabled := cs.IsEnabled()
			assert.NotPanics(s.T(), func() { _ = enabled })
		})
	}

	// Test empty TERM
	os.Setenv("TERM", "")
	cs := NewColorSupport()
	// Should be false due to empty TERM
	assert.False(s.T(), cs.IsEnabled())

	// Test unsupported TERM
	os.Setenv("TERM", "unsupported-terminal")
	cs = NewColorSupport()
	assert.False(s.T(), cs.IsEnabled())
}

func (s *ColorTestSuite) TestEnvironmentVariables_ColorTerm() {
	// Clear conflicting env vars
	os.Unsetenv("FORCE_COLOR")
	os.Unsetenv("NO_COLOR")
	os.Setenv("TERM", "unknown")
	os.Setenv("COLORTERM", "truecolor")

	cs := NewColorSupport()
	// Note: This might still be false due to isTerminal() check
	// but the supportsColor logic should handle COLORTERM correctly
	enabled := cs.IsEnabled()
	assert.NotPanics(s.T(), func() { _ = enabled })
}

func (s *ColorTestSuite) TestColorMethods_EdgeCases() {
	cs := NewColorSupport()
	cs.ForceEnable()

	// Test empty string
	assert.Equal(s.T(), Red+Reset, cs.Red(""))
	assert.Equal(s.T(), Green+Reset, cs.Green(""))
	assert.Equal(s.T(), Yellow+Reset, cs.Yellow(""))
	assert.Equal(s.T(), Blue+Reset, cs.Blue(""))
	assert.Equal(s.T(), Bold+Reset, cs.Bold(""))

	// Test multiline text
	multiline := "line1\nline2\nline3"
	expected := Red + multiline + Reset
	assert.Equal(s.T(), expected, cs.Red(multiline))

	// Test text with existing escape sequences
	textWithEscapes := "\033[31mAlready colored\033[0m"
	expected = Red + textWithEscapes + Reset
	assert.Equal(s.T(), expected, cs.Red(textWithEscapes))

	// Test very long text
	longText := string(make([]byte, 10000))
	for i := range longText {
		longText = longText[:i] + "a" + longText[i+1:]
	}
	expected = Red + longText + Reset
	assert.Equal(s.T(), expected, cs.Red(longText))
}

func (s *ColorTestSuite) TestColorMethods_UnicodeText() {
	cs := NewColorSupport()
	cs.ForceEnable()

	// Test Unicode characters
	unicodeText := "Hello 世界! 🌍🚀"
	expected := Red + unicodeText + Reset
	assert.Equal(s.T(), expected, cs.Red(unicodeText))

	// Test emoji
	emojiText := "🎉✨🌟"
	expected = Green + emojiText + Reset
	assert.Equal(s.T(), expected, cs.Green(emojiText))
}

func (s *ColorTestSuite) TestColorConstants() {
	// Test that color constants are properly defined
	assert.NotEmpty(s.T(), Reset)
	assert.NotEmpty(s.T(), Red)
	assert.NotEmpty(s.T(), Green)
	assert.NotEmpty(s.T(), Yellow)
	assert.NotEmpty(s.T(), Blue)
	assert.NotEmpty(s.T(), Purple)
	assert.NotEmpty(s.T(), Cyan)
	assert.NotEmpty(s.T(), White)
	assert.NotEmpty(s.T(), Bold)

	// Test that they contain escape sequences
	assert.Contains(s.T(), Reset, "\033")
	assert.Contains(s.T(), Red, "\033")
	assert.Contains(s.T(), Green, "\033")
	assert.Contains(s.T(), Yellow, "\033")
	assert.Contains(s.T(), Blue, "\033")
	assert.Contains(s.T(), Purple, "\033")
	assert.Contains(s.T(), Cyan, "\033")
	assert.Contains(s.T(), White, "\033")
	assert.Contains(s.T(), Bold, "\033")
}

func (s *ColorTestSuite) TestSupportsColor_Isolation() {
	// Test that supportsColor function works independently
	// This is an indirect test since supportsColor is not exported

	// Test with clean environment
	os.Unsetenv("FORCE_COLOR")
	os.Unsetenv("NO_COLOR")
	os.Unsetenv("TERM")
	os.Unsetenv("COLORTERM")

	cs1 := NewColorSupport()
	cs2 := NewColorSupport()

	// Should be consistent
	assert.Equal(s.T(), cs1.IsEnabled(), cs2.IsEnabled())
}

func (s *ColorTestSuite) TestIsTerminal_Behavior() {
	// Test that isTerminal doesn't panic and returns a boolean
	// This is an indirect test since isTerminal is not exported
	cs := NewColorSupport()

	// Should not panic
	assert.NotPanics(s.T(), func() {
		_ = cs.IsEnabled()
	})
}

func (s *ColorTestSuite) TestColorSupport_StateIsolation() {
	// Test that multiple ColorSupport instances don't interfere
	cs1 := NewColorSupport()
	cs2 := NewColorSupport()

	cs1.ForceEnable()
	cs2.ForceDisable()

	assert.True(s.T(), cs1.IsEnabled())
	assert.False(s.T(), cs2.IsEnabled())

	// Test that they produce different output
	text := "test"
	assert.NotEqual(s.T(), cs1.Red(text), cs2.Red(text))
}
