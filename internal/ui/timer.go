// Package ui provides terminal user interface components for the today app.
package ui

import (
	"fmt"
	"strings"
	"time"

	"today/internal/config"
	"today/internal/storage"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// TimerPane handles time tracking display and interactions.
type TimerPane struct {
	timerStore *storage.TimerStore
	focused    bool
	width      int
	height     int
	switching  bool // Are we switching projects?
	input      textinput.Model
	storage    *storage.Storage
	styles     *Styles

	// Key bindings
	keys      TimerKeyMap
	inputKeys InputKeyMap
}

// NewTimerPane creates a new timer pane.
func NewTimerPane(store *storage.Storage, styles *Styles) *TimerPane {
	return NewTimerPaneWithKeys(store, styles, &config.KeysConfig{})
}

// NewTimerPaneWithKeys creates a new timer pane with custom key bindings.
func NewTimerPaneWithKeys(store *storage.Storage, styles *Styles, keyCfg *config.KeysConfig) *TimerPane {
	if keyCfg == nil {
		keyCfg = &config.KeysConfig{}
	}
	ti := textinput.New()
	ti.Placeholder = "Project name"
	ti.CharLimit = 50
	ti.Width = 30

	return &TimerPane{
		timerStore: &storage.TimerStore{},
		focused:    false,
		input:      ti,
		storage:    store,
		styles:     styles,
		keys:       NewTimerKeyMap(keyCfg),
		inputKeys:  NewInputKeyMap(keyCfg),
	}
}

// LoadTimerCmd returns a command that loads timer state asynchronously.
func (p *TimerPane) LoadTimerCmd() tea.Cmd {
	return loadTimerCmd(p.storage)
}

// setTimerStore updates the timer store state.
func (p *TimerPane) setTimerStore(store *storage.TimerStore) {
	p.timerStore = store
}

// SetSize sets the pane dimensions.
func (p *TimerPane) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.input.Width = max(10, width-4)
}

// SetFocused sets whether this pane is focused.
func (p *TimerPane) SetFocused(focused bool) {
	p.focused = focused
}

// IsFocused returns whether this pane is focused.
func (p *TimerPane) IsFocused() bool {
	return p.focused
}

// IsSwitching returns whether we're in project switch mode.
func (p *TimerPane) IsSwitching() bool {
	return p.switching
}

// IsRunning returns whether the timer is currently running.
func (p *TimerPane) IsRunning() bool {
	return p.timerStore.Current != nil
}

// Update handles messages for the timer pane.
func (p *TimerPane) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// Handle async messages first
	switch msg := msg.(type) {
	case timerLoadedMsg:
		if msg.store != nil {
			p.setTimerStore(msg.store)
		}
		return nil

	case timerStartedMsg:
		// Reload to get updated state
		return p.LoadTimerCmd()

	case timerStoppedMsg:
		// Reload to get updated state
		return p.LoadTimerCmd()
	}

	// If we're switching projects, handle input
	if p.switching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, p.inputKeys.Confirm):
				project := strings.TrimSpace(p.input.Value())
				if project != "" {
					p.switching = false
					p.input.Reset()
					// Return command to start timer asynchronously
					return startTimerCmd(p.storage, project)
				}
				p.switching = false
				p.input.Reset()
				return nil

			case key.Matches(msg, p.inputKeys.Cancel):
				p.switching = false
				p.input.Reset()
				return nil
			}
		}

		p.input, cmd = p.input.Update(msg)
		return cmd
	}

	// Normal mode
	if !p.focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.MouseMsg:
		return p.handleMouse(msg)

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, p.keys.Toggle):
			// Toggle timer (start/stop) asynchronously
			if p.IsRunning() {
				return stopTimerCmd(p.storage)
			}
			// If no current project, prompt for one
			p.switching = true
			p.input.Focus()
			return textinput.Blink

		case key.Matches(msg, p.keys.Switch):
			// Switch project (stops current, starts new)
			p.switching = true
			p.input.Focus()
			return textinput.Blink

		case key.Matches(msg, p.keys.Stop):
			// Stop timer asynchronously
			if p.IsRunning() {
				return stopTimerCmd(p.storage)
			}
		}
	}

	return nil
}

// handleMouse processes mouse events for the timer pane.
func (p *TimerPane) handleMouse(msg tea.MouseMsg) tea.Cmd {
	// Content starts after title (1) + separator (1) + blank (1) = row 3
	const headerRows = 3

	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
		// Click anywhere in the timer display area toggles the timer
		if msg.Y >= headerRows && msg.Y < headerRows+4 {
			if p.IsRunning() {
				return stopTimerCmd(p.storage)
			}
			// If no current project, prompt for one
			p.switching = true
			p.input.Focus()
			return textinput.Blink
		}
	}

	return nil
}

// View renders the timer pane.
func (p *TimerPane) View() string {
	var b strings.Builder

	// Title
	title := p.styles.PaneTitleStyle.Render("⏱️  TIMER")
	b.WriteString(title)
	b.WriteString("\n")

	// Separator
	sepWidth := p.width - 4
	if sepWidth < 10 {
		sepWidth = 30
	}
	b.WriteString(p.styleMutedText(strings.Repeat("─", sepWidth)))
	b.WriteString("\n\n")

	// Current timer status
	if p.timerStore.Current != nil {
		// Timer is running
		elapsed := time.Since(p.timerStore.Current.StartedAt)

		// Running indicator
		indicator := p.styles.TimerRunningStyle.Render("▶")
		project := p.styles.TimerProjectStyle.Render(p.timerStore.Current.Project)
		b.WriteString(fmt.Sprintf("  %s %s\n", indicator, project))

		// Elapsed time (big)
		elapsedStr := formatDuration(elapsed)
		b.WriteString("    " + p.styles.TimerRunningStyle.Render(elapsedStr))
		b.WriteString("\n")
	} else {
		// Timer is stopped
		b.WriteString("  " + p.styles.TimerStoppedStyle.Render("■ Not running"))
		b.WriteString("\n\n")
		b.WriteString("  " + p.styleMutedText("Press space to start"))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Today's total
	todayTotal := p.storage.GetTodayTotal(p.timerStore)
	todayStr := formatDuration(todayTotal)
	b.WriteString("  " + p.styles.StatLabelStyle.Render("Today: ") + p.styles.StatValueStyle.Render(todayStr))
	b.WriteString("\n")

	// Week total
	weekTotal := p.storage.GetWeekTotal(p.timerStore)
	weekStr := formatDuration(weekTotal)
	b.WriteString("  " + p.styles.StatLabelStyle.Render("Week:  ") + p.styles.StatValueStyle.Render(weekStr))
	b.WriteString("\n")

	// Recent entries (last 3)
	b.WriteString("\n")
	b.WriteString("  " + p.styles.StatLabelStyle.Render("Recent:"))
	b.WriteString("\n")

	entries := p.getRecentEntries(3)
	if len(entries) == 0 {
		b.WriteString("  " + p.styleMutedText("  No entries yet"))
		b.WriteString("\n")
	} else {
		for _, entry := range entries {
			duration := entry.EndedAt.Sub(entry.StartedAt)
			timeStr := entry.StartedAt.Format("15:04")
			b.WriteString(fmt.Sprintf("    %s %s (%s)\n",
				p.styleMutedText(timeStr),
				entry.Project,
				formatDurationShort(duration),
			))
		}
	}

	// Input field when switching projects
	if p.switching {
		b.WriteString("\n")
		prompt := p.styles.InputPromptStyle.Render("Project: ")
		b.WriteString("  " + prompt + p.input.View())
		b.WriteString("\n")
	}

	// Apply pane style
	content := b.String()
	style := p.styles.PaneStyle
	if p.focused {
		style = p.styles.PaneFocusedStyle
	}

	return style.Width(p.width).Height(p.height).Render(content)
}

// getRecentEntries returns the N most recent timer entries.
func (p *TimerPane) getRecentEntries(n int) []storage.TimerEntry {
	entries := p.timerStore.Entries
	if len(entries) <= n {
		// Return in reverse order (newest first)
		result := make([]storage.TimerEntry, len(entries))
		for i, e := range entries {
			result[len(entries)-1-i] = e
		}
		return result
	}

	// Return last n entries in reverse order
	result := make([]storage.TimerEntry, n)
	for i := 0; i < n; i++ {
		result[i] = entries[len(entries)-1-i]
	}
	return result
}

// formatDuration formats a duration as HH:MM:SS.
func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// formatDurationShort formats a duration as Xh Xm.
func formatDurationShort(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// styleMutedText applies muted style to text.
func (p *TimerPane) styleMutedText(s string) string {
	return p.styles.StatLabelStyle.Render(s)
}

// GetCurrentProject returns the current project name (if timer is running).
func (p *TimerPane) GetCurrentProject() string {
	if p.timerStore.Current != nil {
		return p.timerStore.Current.Project
	}
	return ""
}

// GetElapsed returns the current elapsed time.
func (p *TimerPane) GetElapsed() time.Duration {
	if p.timerStore.Current != nil {
		return time.Since(p.timerStore.Current.StartedAt)
	}
	return 0
}
