// Package ui provides terminal user interface components for the today app.
package ui

import (
	"fmt"
	"strings"

	"today/internal/config"
	"today/internal/storage"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// HabitsPane handles habit tracking display and interactions.
type HabitsPane struct {
	habitStore *storage.HabitStore
	cursor     int
	focused    bool
	width      int
	height     int
	adding     bool
	addStep    int // 0 = name, 1 = icon
	input      textinput.Model
	newName    string
	storage    *storage.Storage
	styles     *Styles

	// Key bindings
	keys      HabitKeyMap
	inputKeys InputKeyMap
}

// NewHabitsPane creates a new habits pane.
func NewHabitsPane(store *storage.Storage, styles *Styles) *HabitsPane {
	return NewHabitsPaneWithKeys(store, styles, &config.KeysConfig{})
}

// NewHabitsPaneWithKeys creates a new habits pane with custom key bindings.
func NewHabitsPaneWithKeys(store *storage.Storage, styles *Styles, keyCfg *config.KeysConfig) *HabitsPane {
	if keyCfg == nil {
		keyCfg = &config.KeysConfig{}
	}
	ti := textinput.New()
	ti.Placeholder = "Habit name (e.g., Exercise)"
	ti.CharLimit = 30
	ti.Width = 30

	return &HabitsPane{
		habitStore: &storage.HabitStore{},
		cursor:     0,
		focused:    false,
		input:      ti,
		storage:    store,
		styles:     styles,
		keys:       NewHabitKeyMap(keyCfg),
		inputKeys:  NewInputKeyMap(keyCfg),
	}
}

// LoadHabitsCmd returns a command that loads habits asynchronously.
func (p *HabitsPane) LoadHabitsCmd() tea.Cmd {
	return loadHabitsCmd(p.storage)
}

// setHabitStore updates the habit store and adjusts cursor bounds.
func (p *HabitsPane) setHabitStore(store *storage.HabitStore) {
	p.habitStore = store
	if p.cursor >= len(p.habitStore.Habits) {
		p.cursor = max(0, len(p.habitStore.Habits)-1)
	}
}

// SetSize sets the pane dimensions.
func (p *HabitsPane) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.input.Width = max(10, width-6)
}

// SetFocused sets whether this pane is focused.
func (p *HabitsPane) SetFocused(focused bool) {
	p.focused = focused
}

// IsFocused returns whether this pane is focused.
func (p *HabitsPane) IsFocused() bool {
	return p.focused
}

// IsAdding returns whether we're in add mode.
func (p *HabitsPane) IsAdding() bool {
	return p.adding
}

// Update handles messages for the habits pane.
func (p *HabitsPane) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// Handle async messages first
	switch msg := msg.(type) {
	case habitsLoadedMsg:
		if msg.store != nil {
			p.setHabitStore(msg.store)
		}
		return nil

	case habitAddedMsg:
		if msg.err == nil {
			// Reload to get updated list
			return p.LoadHabitsCmd()
		}
		return nil

	case habitToggledMsg:
		// Reload to refresh state
		return p.LoadHabitsCmd()

	case habitDeletedMsg:
		// Reload to refresh list
		return p.LoadHabitsCmd()
	}

	// If we're adding a habit, handle input
	if p.adding {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, p.inputKeys.Confirm):
				if p.addStep == 0 {
					// Got name, now get icon
					p.newName = strings.TrimSpace(p.input.Value())
					if p.newName != "" {
						p.addStep = 1
						p.input.Reset()
						p.input.Placeholder = "Icon (emoji, e.g., 🏃)"
						p.input.CharLimit = 4
					}
				} else {
					// Got icon, create habit asynchronously
					icon := strings.TrimSpace(p.input.Value())
					if icon == "" {
						icon = "✓" // Default icon
					}
					name := p.newName
					p.resetAddMode()
					return addHabitCmd(p.storage, name, icon)
				}
				return nil

			case key.Matches(msg, p.inputKeys.Cancel):
				p.resetAddMode()
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
		case key.Matches(msg, p.keys.Down):
			if len(p.habitStore.Habits) > 0 {
				p.cursor = min(p.cursor+1, len(p.habitStore.Habits)-1)
			}

		case key.Matches(msg, p.keys.Up):
			if len(p.habitStore.Habits) > 0 {
				p.cursor = max(p.cursor-1, 0)
			}

		case key.Matches(msg, p.keys.Add):
			p.adding = true
			p.addStep = 0
			p.input.Placeholder = "Habit name (e.g., Exercise)"
			p.input.CharLimit = 30
			p.input.Focus()
			return textinput.Blink

		case key.Matches(msg, p.keys.Toggle):
			// Toggle habit for today asynchronously
			if len(p.habitStore.Habits) > 0 && p.cursor < len(p.habitStore.Habits) {
				habit := p.habitStore.Habits[p.cursor]
				return toggleHabitCmd(p.storage, habit.ID)
			}

		case key.Matches(msg, p.keys.Delete):
			// Delete habit asynchronously
			if len(p.habitStore.Habits) > 0 && p.cursor < len(p.habitStore.Habits) {
				habit := p.habitStore.Habits[p.cursor]
				return deleteHabitCmd(p.storage, habit.ID)
			}
		}
	}

	return nil
}

// resetAddMode resets the add habit state.
func (p *HabitsPane) resetAddMode() {
	p.adding = false
	p.addStep = 0
	p.newName = ""
	p.input.Reset()
	p.input.Placeholder = "Habit name (e.g., Exercise)"
	p.input.CharLimit = 30
}

// handleMouse processes mouse events for the habits pane.
func (p *HabitsPane) handleMouse(msg tea.MouseMsg) tea.Cmd {
	if len(p.habitStore.Habits) == 0 {
		return nil
	}

	// Content starts after title (1) + separator (1) + week view rows
	// Week view: header (1) + days (1) + blank (1) = 3 lines
	// Then habits list starts
	const headerRows = 5

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		p.cursor = max(p.cursor-1, 0)
		return nil

	case tea.MouseButtonWheelDown:
		p.cursor = min(p.cursor+1, len(p.habitStore.Habits)-1)
		return nil

	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return nil
		}

		// Calculate which habit was clicked
		habitRow := msg.Y - headerRows
		if habitRow < 0 || habitRow >= len(p.habitStore.Habits) {
			return nil
		}

		// Move cursor to clicked habit
		p.cursor = habitRow

		// Check if click was on the icon/checkbox area (first few chars)
		// Habit format: "🏃 Exercise  ●○○" - icon is at start
		if msg.X < 4 {
			// Toggle the clicked habit
			habit := p.habitStore.Habits[p.cursor]
			return toggleHabitCmd(p.storage, habit.ID)
		}
	}

	return nil
}

// View renders the habits pane.
func (p *HabitsPane) View() string {
	var b strings.Builder

	// Title
	title := p.styles.PaneTitleStyle.Render("🔥 HABITS")
	b.WriteString(title)
	b.WriteString("\n")

	// Separator
	sepWidth := p.width - 4
	if sepWidth < 10 {
		sepWidth = 30
	}
	b.WriteString(p.styleMutedText(strings.Repeat("─", sepWidth)))
	b.WriteString("\n")

	// Habits list
	if len(p.habitStore.Habits) == 0 && !p.adding {
		b.WriteString("\n")
		b.WriteString(p.styleMutedText("  No habits yet."))
		b.WriteString("\n")
		b.WriteString(p.styleMutedText("  Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		b.WriteString("\n")

		// Calculate max streak for display
		maxStreak := 0
		for _, habit := range p.habitStore.Habits {
			streak := p.storage.GetHabitStreak(p.habitStore, habit.ID)
			if streak > maxStreak {
				maxStreak = streak
			}
		}

		for i, habit := range p.habitStore.Habits {
			// Selection indicator
			prefix := "  "
			if i == p.cursor && p.focused && !p.adding {
				prefix = "▶ "
			}

			// Icon and name
			line := fmt.Sprintf("%s%s %s  ", prefix, habit.Icon, habit.Name)

			// Week view (last 7 days)
			week := p.storage.GetHabitWeek(p.habitStore, habit.ID)
			weekView := p.renderWeekView(week)
			line += weekView

			// Count for this week
			weekCount := 0
			for _, done := range week {
				if done {
					weekCount++
				}
			}
			line += fmt.Sprintf("  %d/7", weekCount)

			// Streak (if > 1)
			streak := p.storage.GetHabitStreak(p.habitStore, habit.ID)
			if streak > 1 {
				line += " " + p.styles.HabitStreakStyle.Render(fmt.Sprintf("🔥%d", streak))
			}

			// Highlight if selected
			if i == p.cursor && p.focused && !p.adding {
				line = p.styles.TaskSelectedStyle.Render(line)
			}

			b.WriteString(line)
			b.WriteString("\n")
		}

		// Overall streak
		if maxStreak > 0 {
			b.WriteString("\n")
			b.WriteString("  " + p.styles.StatLabelStyle.Render("Best streak: ") + p.styles.HabitStreakStyle.Render(fmt.Sprintf("%d days 🔥", maxStreak)))
			b.WriteString("\n")
		}
	}

	// Day labels
	b.WriteString("\n")
	b.WriteString("  " + p.styleMutedText(p.getDayLabels()))
	b.WriteString("\n")

	// Input field when adding
	if p.adding {
		b.WriteString("\n")
		var prompt string
		if p.addStep == 0 {
			prompt = p.styles.InputPromptStyle.Render("Name: ")
		} else {
			prompt = p.styles.InputPromptStyle.Render("Icon: ")
		}
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

// renderWeekView creates the visual week representation.
func (p *HabitsPane) renderWeekView(week []bool) string {
	var result string
	for _, done := range week {
		if done {
			result += p.styles.HabitDoneIcon + " "
		} else {
			result += p.styles.HabitUndoneIcon + " "
		}
	}
	return strings.TrimSuffix(result, " ")
}

// styleMutedText applies muted style to text.
func (p *HabitsPane) styleMutedText(s string) string {
	return p.styles.StatLabelStyle.Render(s)
}

// getDayLabels returns the day labels for the week view.
func (p *HabitsPane) getDayLabels() string {
	today := p.storage.Now()
	days := make([]string, 7)

	for i := 0; i < 7; i++ {
		day := today.AddDate(0, 0, -(6 - i))
		days[i] = day.Format("Mon")[:1] // First letter of day
	}

	// Add spacing to align with circles
	result := "       " // Indent to align with habit icons
	for _, d := range days {
		result += d + " "
	}
	return strings.TrimSuffix(result, " ")
}

// GetTodayCompletionRate returns how many habits were completed today.
func (p *HabitsPane) GetTodayCompletionRate() (done, total int) {
	today := p.storage.Now().Format("2006-01-02")
	total = len(p.habitStore.Habits)

	for _, habit := range p.habitStore.Habits {
		if p.storage.IsHabitDoneOnDate(p.habitStore, habit.ID, today) {
			done++
		}
	}

	return done, total
}
