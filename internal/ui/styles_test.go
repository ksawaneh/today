package ui

import (
	"testing"

	"today/internal/config"

	"github.com/charmbracelet/lipgloss"
)

func TestNewStyles_UsesThemeColors(t *testing.T) {
	// Create a custom theme config
	theme := &config.ThemeConfig{
		Primary:    "#FF0000", // Red
		Accent:     "#00FF00", // Green
		Muted:      "#0000FF", // Blue
		Background: "#000000", // Black
		Text:       "#FFFFFF", // White
	}

	styles := NewStylesFromTheme(theme)

	// Verify colors are applied
	if styles.ColorPrimary != lipgloss.Color("#FF0000") {
		t.Errorf("ColorPrimary = %v, want #FF0000", styles.ColorPrimary)
	}
	if styles.ColorAccent != lipgloss.Color("#00FF00") {
		t.Errorf("ColorAccent = %v, want #00FF00", styles.ColorAccent)
	}
	if styles.ColorMuted != lipgloss.Color("#0000FF") {
		t.Errorf("ColorMuted = %v, want #0000FF", styles.ColorMuted)
	}
	if styles.ColorBg != lipgloss.Color("#000000") {
		t.Errorf("ColorBg = %v, want #000000", styles.ColorBg)
	}
	if styles.ColorText != lipgloss.Color("#FFFFFF") {
		t.Errorf("ColorText = %v, want #FFFFFF", styles.ColorText)
	}
}

func TestNewStyles_UsesDefaults(t *testing.T) {
	// Create theme with empty values
	theme := &config.ThemeConfig{}

	styles := NewStylesFromTheme(theme)

	// Verify defaults are applied
	if styles.ColorPrimary != lipgloss.Color("#7C3AED") {
		t.Errorf("ColorPrimary = %v, want default #7C3AED", styles.ColorPrimary)
	}
	if styles.ColorAccent != lipgloss.Color("#3B82F6") {
		t.Errorf("ColorAccent = %v, want default #3B82F6", styles.ColorAccent)
	}
	if styles.ColorMuted != lipgloss.Color("#6B7280") {
		t.Errorf("ColorMuted = %v, want default #6B7280", styles.ColorMuted)
	}
}

func TestNewStyles_ComponentStylesInitialized(t *testing.T) {
	theme := &config.ThemeConfig{
		Primary: "#FF0000",
	}

	styles := NewStylesFromTheme(theme)

	// Verify component styles are initialized (non-nil)
	if styles.TitleStyle.GetBackground() != lipgloss.Color("#FF0000") {
		t.Error("TitleStyle should use Primary color for background")
	}

	if styles.PaneFocusedStyle.GetBorderTopForeground() != lipgloss.Color("#FF0000") {
		t.Error("PaneFocusedStyle should use Primary color for border")
	}

	if styles.PaneTitleStyle.GetForeground() != lipgloss.Color("#FF0000") {
		t.Error("PaneTitleStyle should use Primary color for foreground")
	}
}

func TestNewStyles_FromConfig(t *testing.T) {
	// Test the convenience function that accepts full Config
	cfg := config.Default()
	cfg.Theme.Primary = "#123456"

	styles := NewStyles(cfg)

	if styles.ColorPrimary != lipgloss.Color("#123456") {
		t.Errorf("ColorPrimary = %v, want #123456", styles.ColorPrimary)
	}
}

func TestRenderHelp(t *testing.T) {
	styles := createTestStyles()

	// Test RenderHelp method
	output := styles.RenderHelp(
		"a", "add",
		"d", "done",
	)

	if len(output) == 0 {
		t.Error("RenderHelp should produce output")
	}

	// The output should contain the keys and descriptions
	// Note: exact format depends on lipgloss rendering, so we just check it's not empty
}
