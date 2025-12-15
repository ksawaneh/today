package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// HelpOverlay renders a help screen
type HelpOverlay struct {
	width  int
	height int
	styles *Styles
}

// NewHelpOverlay creates a new help overlay
func NewHelpOverlay(styles *Styles) *HelpOverlay {
	return &HelpOverlay{
		styles: styles,
	}
}

// SetSize sets the overlay dimensions
func (h *HelpOverlay) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// View renders the help overlay
func (h *HelpOverlay) View() string {
	overlayWidth := 60
	if h.width > 0 {
		overlayWidth = min(60, max(20, h.width-4))
	}

	// Styles for help overlay
	overlayStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(h.styles.ColorPrimary).
		Padding(1, 2).
		Width(overlayWidth)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(h.styles.ColorPrimary).
		MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(h.styles.ColorAccent).
		MarginTop(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(h.styles.ColorWarning).
		Width(12)

	descStyle := lipgloss.NewStyle().
		Foreground(h.styles.ColorText)

	mutedStyle := lipgloss.NewStyle().
		Foreground(h.styles.ColorTextMuted).
		Italic(true)

	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("📖 today - Keyboard Shortcuts"))
	b.WriteString("\n\n")

	// Global
	b.WriteString(sectionStyle.Render("Global"))
	b.WriteString("\n")
	b.WriteString(keyStyle.Render("Tab") + descStyle.Render("Switch pane") + "\n")
	b.WriteString(keyStyle.Render("1 / 2 / 3") + descStyle.Render("Jump to pane") + "\n")
	b.WriteString(keyStyle.Render("?") + descStyle.Render("Toggle help") + "\n")
	b.WriteString(keyStyle.Render("q") + descStyle.Render("Quit") + "\n")

	// Tasks
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("Tasks"))
	b.WriteString("\n")
	b.WriteString(keyStyle.Render("a") + descStyle.Render("Add task") + "\n")
	b.WriteString(keyStyle.Render("d / Space") + descStyle.Render("Toggle done") + "\n")
	b.WriteString(keyStyle.Render("x") + descStyle.Render("Delete task") + "\n")
	b.WriteString(keyStyle.Render("j / k") + descStyle.Render("Navigate up/down") + "\n")
	b.WriteString(keyStyle.Render("g / G") + descStyle.Render("Go to top/bottom") + "\n")

	// Timer
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("Timer"))
	b.WriteString("\n")
	b.WriteString(keyStyle.Render("Space") + descStyle.Render("Start/stop timer") + "\n")
	b.WriteString(keyStyle.Render("s") + descStyle.Render("Switch project") + "\n")

	// Habits
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("Habits"))
	b.WriteString("\n")
	b.WriteString(keyStyle.Render("a") + descStyle.Render("Add habit") + "\n")
	b.WriteString(keyStyle.Render("Space / d") + descStyle.Render("Toggle today") + "\n")
	b.WriteString(keyStyle.Render("x") + descStyle.Render("Delete habit") + "\n")
	b.WriteString(keyStyle.Render("j / k") + descStyle.Render("Navigate up/down") + "\n")

	// Input mode
	b.WriteString("\n")
	b.WriteString(sectionStyle.Render("Input Mode"))
	b.WriteString("\n")
	b.WriteString(keyStyle.Render("Enter") + descStyle.Render("Save") + "\n")
	b.WriteString(keyStyle.Render("Esc") + descStyle.Render("Cancel") + "\n")

	// Footer
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Press ? or Esc to close"))

	content := overlayStyle.Render(b.String())

	// Center the overlay
	return lipgloss.Place(
		h.width,
		h.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

// RenderCentered centers content in the terminal
func RenderCentered(content string, width, height int) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
