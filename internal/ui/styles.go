package ui

import (
	"today/internal/config"

	"github.com/charmbracelet/lipgloss"
)

// Styles holds all application styles, initialized with theme configuration.
type Styles struct {
	// Colors
	ColorPrimary   lipgloss.Color
	ColorSecondary lipgloss.Color
	ColorMuted     lipgloss.Color
	ColorDanger    lipgloss.Color
	ColorWarning   lipgloss.Color
	ColorSuccess   lipgloss.Color
	ColorAccent    lipgloss.Color
	ColorBg        lipgloss.Color
	ColorBgLight   lipgloss.Color
	ColorText      lipgloss.Color
	ColorTextMuted lipgloss.Color

	// Component styles
	TitleStyle       lipgloss.Style
	DateStyle        lipgloss.Style
	PaneStyle        lipgloss.Style
	PaneFocusedStyle lipgloss.Style
	PaneTitleStyle   lipgloss.Style

	TaskDoneStyle       lipgloss.Style
	TaskPendingStyle    lipgloss.Style
	TaskSelectedStyle   lipgloss.Style
	TaskCheckboxDone    string
	TaskCheckboxPending string

	// Priority badge styles
	PriorityHighStyle   lipgloss.Style
	PriorityMediumStyle lipgloss.Style
	PriorityLowStyle    lipgloss.Style

	// Due date indicator styles
	DueDateOverdueStyle lipgloss.Style
	DueDateTodayStyle   lipgloss.Style
	DueDateFutureStyle  lipgloss.Style

	HabitDoneIcon   string
	HabitUndoneIcon string
	HabitStreakStyle lipgloss.Style

	TimerRunningStyle lipgloss.Style
	TimerStoppedStyle lipgloss.Style
	TimerProjectStyle lipgloss.Style

	HelpStyle    lipgloss.Style
	HelpKeyStyle lipgloss.Style

	StatusStyle lipgloss.Style
	ErrorStyle  lipgloss.Style

	InputPromptStyle lipgloss.Style
	InputTextStyle   lipgloss.Style

	StatLabelStyle lipgloss.Style
	StatValueStyle lipgloss.Style

	// Sync status styles
	SyncSyncedStyle   lipgloss.Style // Synced (green checkmark)
	SyncPendingStyle  lipgloss.Style // Has uncommitted changes (yellow)
	SyncAheadStyle    lipgloss.Style // Ahead of remote (blue)
	SyncBehindStyle   lipgloss.Style // Behind remote (orange)
	SyncDisabledStyle lipgloss.Style // No remote / sync disabled (muted)
}

// NewStyles creates a new Styles instance from the given config.
// If a theme color is empty, it uses the appropriate default.
func NewStyles(cfg *config.Config) *Styles {
	return NewStylesFromTheme(&cfg.Theme)
}

// NewStylesFromTheme creates a new Styles instance from a ThemeConfig.
// If a theme color is empty, it uses the appropriate default.
func NewStylesFromTheme(theme *config.ThemeConfig) *Styles {
	s := &Styles{}

	// Initialize colors from config with fallbacks to defaults
	s.ColorPrimary = colorOrDefault(theme.Primary, "#7C3AED")
	s.ColorSecondary = colorOrDefault(theme.Accent, "#10B981")
	s.ColorMuted = colorOrDefault(theme.Muted, "#6B7280")

	// Fixed semantic colors (not configurable from theme)
	s.ColorDanger = lipgloss.Color("#EF4444")
	s.ColorWarning = lipgloss.Color("#F59E0B")
	s.ColorSuccess = lipgloss.Color("#10B981")
	s.ColorAccent = colorOrDefault(theme.Accent, "#3B82F6")

	// Background and text colors
	s.ColorBg = colorOrDefault(theme.Background, "#1F2937")
	s.ColorBgLight = lipgloss.Color("#374151")
	s.ColorText = colorOrDefault(theme.Text, "#F9FAFB")
	s.ColorTextMuted = lipgloss.Color("#9CA3AF")

	// Initialize all component styles
	s.initComponentStyles()

	return s
}

// colorOrDefault returns the lipgloss.Color from hex string, or default if empty.
func colorOrDefault(hex, defaultHex string) lipgloss.Color {
	if hex != "" {
		return lipgloss.Color(hex)
	}
	return lipgloss.Color(defaultHex)
}

// initComponentStyles initializes all component styles based on the color palette.
func (s *Styles) initComponentStyles() {
	// Title bar
	s.TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.ColorText).
		Background(s.ColorPrimary).
		Padding(0, 1)

	// Date in title bar
	s.DateStyle = lipgloss.NewStyle().
		Foreground(s.ColorTextMuted)

	// Pane styles
	s.PaneStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.ColorMuted).
		Padding(0, 1)

	s.PaneFocusedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(s.ColorPrimary).
		Padding(0, 1)

	s.PaneTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(s.ColorPrimary).
		MarginBottom(1)

	// Task styles
	s.TaskDoneStyle = lipgloss.NewStyle().
		Foreground(s.ColorTextMuted).
		Strikethrough(true)

	s.TaskPendingStyle = lipgloss.NewStyle().
		Foreground(s.ColorText)

	s.TaskSelectedStyle = lipgloss.NewStyle().
		Background(s.ColorBgLight).
		Foreground(s.ColorText).
		Bold(true)

	s.TaskCheckboxDone = lipgloss.NewStyle().Foreground(s.ColorSuccess).Render("[✓]")
	s.TaskCheckboxPending = lipgloss.NewStyle().Foreground(s.ColorMuted).Render("[ ]")

	// Priority badge styles
	s.PriorityHighStyle = lipgloss.NewStyle().
		Foreground(s.ColorDanger).
		Bold(true)

	s.PriorityMediumStyle = lipgloss.NewStyle().
		Foreground(s.ColorWarning)

	s.PriorityLowStyle = lipgloss.NewStyle().
		Foreground(s.ColorMuted)

	// Due date indicator styles
	s.DueDateOverdueStyle = lipgloss.NewStyle().
		Foreground(s.ColorDanger).
		Bold(true)

	s.DueDateTodayStyle = lipgloss.NewStyle().
		Foreground(s.ColorWarning)

	s.DueDateFutureStyle = lipgloss.NewStyle().
		Foreground(s.ColorTextMuted)

	// Habit styles
	s.HabitDoneIcon = lipgloss.NewStyle().Foreground(s.ColorSuccess).Render("●")
	s.HabitUndoneIcon = lipgloss.NewStyle().Foreground(s.ColorMuted).Render("○")

	s.HabitStreakStyle = lipgloss.NewStyle().
		Foreground(s.ColorWarning).
		Bold(true)

	// Timer styles
	s.TimerRunningStyle = lipgloss.NewStyle().
		Foreground(s.ColorSuccess).
		Bold(true)

	s.TimerStoppedStyle = lipgloss.NewStyle().
		Foreground(s.ColorMuted)

	s.TimerProjectStyle = lipgloss.NewStyle().
		Foreground(s.ColorAccent).
		Bold(true)

	// Help bar
	s.HelpStyle = lipgloss.NewStyle().
		Foreground(s.ColorTextMuted)

	s.HelpKeyStyle = lipgloss.NewStyle().
		Foreground(s.ColorAccent).
		Bold(true)

	// Status messages
	s.StatusStyle = lipgloss.NewStyle().
		Foreground(s.ColorSuccess).
		Italic(true)

	s.ErrorStyle = lipgloss.NewStyle().
		Foreground(s.ColorDanger).
		Bold(true)

	// Input
	s.InputPromptStyle = lipgloss.NewStyle().
		Foreground(s.ColorPrimary).
		Bold(true)

	s.InputTextStyle = lipgloss.NewStyle().
		Foreground(s.ColorText)

	// Summary stats
	s.StatLabelStyle = lipgloss.NewStyle().
		Foreground(s.ColorTextMuted)

	s.StatValueStyle = lipgloss.NewStyle().
		Foreground(s.ColorText).
		Bold(true)

	// Sync status styles
	s.SyncSyncedStyle = lipgloss.NewStyle().
		Foreground(s.ColorSuccess)

	s.SyncPendingStyle = lipgloss.NewStyle().
		Foreground(s.ColorWarning)

	s.SyncAheadStyle = lipgloss.NewStyle().
		Foreground(s.ColorAccent)

	s.SyncBehindStyle = lipgloss.NewStyle().
		Foreground(s.ColorWarning).
		Bold(true)

	s.SyncDisabledStyle = lipgloss.NewStyle().
		Foreground(s.ColorMuted)
}

// RenderHelp renders help text with key bindings using the given styles.
func (s *Styles) RenderHelp(keys ...string) string {
	var result string
	for i := 0; i < len(keys); i += 2 {
		if i > 0 {
			result += "  "
		}
		key := keys[i]
		desc := keys[i+1]
		result += s.HelpKeyStyle.Render("["+key+"]") + " " + s.HelpStyle.Render(desc)
	}
	return result
}
