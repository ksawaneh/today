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
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// TaskPane handles the task list display and interactions.
type TaskPane struct {
	tasks   []storage.Task
	cursor  int
	focused bool
	width   int
	height  int
	adding  bool
	input   textinput.Model
	storage *storage.Storage
	styles  *Styles

	// Key bindings
	keys      TaskKeyMap
	inputKeys InputKeyMap
}

// NewTaskPane creates a new task pane.
func NewTaskPane(store *storage.Storage, styles *Styles) *TaskPane {
	return NewTaskPaneWithKeys(store, styles, &config.KeysConfig{})
}

// NewTaskPaneWithKeys creates a new task pane with custom key bindings.
func NewTaskPaneWithKeys(store *storage.Storage, styles *Styles, keyCfg *config.KeysConfig) *TaskPane {
	if keyCfg == nil {
		keyCfg = &config.KeysConfig{}
	}
	ti := textinput.New()
	ti.Placeholder = "What needs to be done?"
	ti.CharLimit = 100
	ti.Width = 40

	return &TaskPane{
		tasks:     []storage.Task{},
		cursor:    0,
		focused:   true,
		input:     ti,
		storage:   store,
		styles:    styles,
		keys:      NewTaskKeyMap(keyCfg),
		inputKeys: NewInputKeyMap(keyCfg),
	}
}

// LoadTasksCmd returns a command that loads tasks asynchronously.
func (p *TaskPane) LoadTasksCmd() tea.Cmd {
	return loadTasksCmd(p.storage)
}

// setTasks updates the task list, sorts it, and adjusts cursor bounds.
func (p *TaskPane) setTasks(tasks []storage.Task) {
	// Sort tasks by priority and due date
	p.tasks = p.storage.SortTasks(tasks)
	if p.cursor >= len(p.tasks) {
		p.cursor = max(0, len(p.tasks)-1)
	}
}

// SetSize sets the pane dimensions.
func (p *TaskPane) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.input.Width = max(10, width-4)
}

// SetFocused sets whether this pane is focused.
func (p *TaskPane) SetFocused(focused bool) {
	p.focused = focused
}

// IsFocused returns whether this pane is focused.
func (p *TaskPane) IsFocused() bool {
	return p.focused
}

// IsAdding returns whether we're in add mode.
func (p *TaskPane) IsAdding() bool {
	return p.adding
}

// Update handles messages for the task pane.
func (p *TaskPane) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	// Handle async messages first
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		if msg.tasks != nil {
			p.setTasks(msg.tasks)
		}
		return nil

	case taskAddedMsg:
		if msg.err == nil {
			// Reload to get updated list with new task
			return p.LoadTasksCmd()
		}
		return nil

	case taskCompletedMsg, taskUncompletedMsg:
		// Reload to refresh task state
		return p.LoadTasksCmd()

	case taskDeletedMsg:
		// Reload to refresh list
		return p.LoadTasksCmd()
	}

	// If we're adding a task, handle input
	if p.adding {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, p.inputKeys.Confirm):
				text := strings.TrimSpace(p.input.Value())
				if text != "" {
					p.adding = false
					p.input.Reset()
					// Return command to add task asynchronously (default priority, no due date)
					return addTaskCmd(p.storage, text, "", storage.PriorityNone, nil)
				}
				p.adding = false
				p.input.Reset()
				return nil

			case key.Matches(msg, p.inputKeys.Cancel):
				p.adding = false
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
		case key.Matches(msg, p.keys.Down):
			if len(p.tasks) > 0 {
				p.cursor = min(p.cursor+1, len(p.tasks)-1)
			}

		case key.Matches(msg, p.keys.Up):
			if len(p.tasks) > 0 {
				p.cursor = max(p.cursor-1, 0)
			}

		case key.Matches(msg, p.keys.Top):
			p.cursor = 0

		case key.Matches(msg, p.keys.Bottom):
			if len(p.tasks) > 0 {
				p.cursor = len(p.tasks) - 1
			}

		case key.Matches(msg, p.keys.Add):
			p.adding = true
			p.input.Focus()
			return textinput.Blink

		case key.Matches(msg, p.keys.Toggle):
			// Toggle done asynchronously
			if len(p.tasks) > 0 && p.cursor < len(p.tasks) {
				task := p.tasks[p.cursor]
				if task.Done {
					return uncompleteTaskCmd(p.storage, task.ID)
				}
				return completeTaskCmd(p.storage, task.ID)
			}

		case key.Matches(msg, p.keys.Delete):
			// Delete task asynchronously
			if len(p.tasks) > 0 && p.cursor < len(p.tasks) {
				task := p.tasks[p.cursor]
				return deleteTaskCmd(p.storage, task.ID)
			}
		}
	}

	return nil
}

// handleMouse processes mouse events for the task pane.
func (p *TaskPane) handleMouse(msg tea.MouseMsg) tea.Cmd {
	if len(p.tasks) == 0 {
		return nil
	}

	// Content starts after title (1) + separator (1) = row 2
	const headerRows = 2

	// Mirror the view windowing logic so clicks map to the visible slice.
	maxTasks := p.height - 6 // Account for title, separator, input, stats
	if maxTasks < 3 {
		maxTasks = 5
	}
	startIdx := 0
	if p.cursor >= maxTasks {
		startIdx = p.cursor - maxTasks + 1
	}

	switch msg.Button {
	case tea.MouseButtonWheelUp:
		p.cursor = max(p.cursor-1, 0)
		return nil

	case tea.MouseButtonWheelDown:
		p.cursor = min(p.cursor+1, len(p.tasks)-1)
		return nil

	case tea.MouseButtonLeft:
		if msg.Action != tea.MouseActionPress {
			return nil
		}

		// Calculate which task was clicked
		taskRow := msg.Y - headerRows
		if taskRow < 0 || taskRow >= maxTasks {
			return nil
		}

		taskIdx := startIdx + taskRow
		if taskIdx < 0 || taskIdx >= len(p.tasks) {
			return nil
		}

		// Move cursor to clicked task
		p.cursor = taskIdx

		// Check if click was on the checkbox area (first few chars)
		// Checkbox format: "![ ] " or "~[x] " - about 5 chars
		if msg.X < 5 {
			// Toggle the clicked task
			task := p.tasks[taskIdx]
			if task.Done {
				return uncompleteTaskCmd(p.storage, task.ID)
			}
			return completeTaskCmd(p.storage, task.ID)
		}
	}

	return nil
}

// View renders the task pane.
func (p *TaskPane) View() string {
	var b strings.Builder

	// Title
	title := p.styles.PaneTitleStyle.Render("✅ TASKS")
	b.WriteString(title)
	b.WriteString("\n")

	// Separator
	sepWidth := p.width - 4
	if sepWidth < 10 {
		sepWidth = 30
	}
	b.WriteString(lipgloss.NewStyle().Foreground(p.styles.ColorMuted).Render(strings.Repeat("─", sepWidth)))
	b.WriteString("\n")

	// Tasks list
	if len(p.tasks) == 0 && !p.adding {
		b.WriteString(lipgloss.NewStyle().Foreground(p.styles.ColorTextMuted).Italic(true).Render("  No tasks yet. Press 'a' to add one."))
		b.WriteString("\n")
	} else {
		// Calculate how many tasks we can show
		maxTasks := p.height - 6 // Account for title, separator, input, stats
		if maxTasks < 3 {
			maxTasks = 5
		}

		// Show tasks
		startIdx := 0
		if p.cursor >= maxTasks {
			startIdx = p.cursor - maxTasks + 1
		}

		doneCount := 0

		for i, task := range p.tasks {
			if task.Done {
				doneCount++
			}

			if i < startIdx || i >= startIdx+maxTasks {
				continue
			}

			// Priority badge (1 char: "!", "~", or " ")
			priorityBadge := p.formatPriorityBadge(task.Priority)

			// Checkbox
			var checkbox string
			if task.Done {
				checkbox = p.styles.TaskCheckboxDone
			} else {
				checkbox = p.styles.TaskCheckboxPending
			}

			// Due date indicator
			dueIndicator := p.formatDueDate(task.DueDate)
			dueWidth := lipgloss.Width(dueIndicator)

			// Calculate available width for task text
			// Layout: [space][priority][checkbox][space][text][space?][due]
			// Fixed parts: 1 (leading space) + 1 (priority) + 3 (checkbox) + 1 (space after checkbox)
			fixedWidth := 6
			if dueWidth > 0 {
				fixedWidth += dueWidth + 1 // due indicator + space before it
			}
			availableTextWidth := p.width - 4 - fixedWidth // 4 for pane padding/borders
			if availableTextWidth < 5 {
				availableTextWidth = 5
			}

			taskText := runewidth.Truncate(task.Text, availableTextWidth, "..")
			taskTextWidth := runewidth.StringWidth(taskText)

			// Build the line
			var line string
			if i == p.cursor && p.focused && !p.adding {
				// Selected: highlight entire line
				textPart := fmt.Sprintf("%s%s %s", priorityBadge, checkbox, taskText)
				if dueWidth > 0 {
					padding := availableTextWidth - taskTextWidth
					if padding < 1 {
						padding = 1
					}
					textPart += strings.Repeat(" ", padding) + dueIndicator
				}
				line = p.styles.TaskSelectedStyle.Render(" " + textPart + " ")
			} else {
				// Normal: assemble with styles
				var styledText string
				if task.Done {
					styledText = p.styles.TaskDoneStyle.Render(taskText)
				} else {
					styledText = p.styles.TaskPendingStyle.Render(taskText)
				}

				textPart := fmt.Sprintf(" %s%s %s", priorityBadge, checkbox, styledText)
				if dueWidth > 0 {
					padding := availableTextWidth - taskTextWidth
					if padding < 1 {
						padding = 1
					}
					textPart += strings.Repeat(" ", padding) + dueIndicator
				}
				line = textPart
			}

			b.WriteString(line)
			b.WriteString("\n")
		}

		// Stats
		b.WriteString("\n")
		stats := p.styles.StatLabelStyle.Render(fmt.Sprintf("%d/%d complete", doneCount, len(p.tasks)))
		b.WriteString("  " + stats)
		b.WriteString("\n")
	}

	// Input field when adding
	if p.adding {
		b.WriteString("\n")
		prompt := p.styles.InputPromptStyle.Render("+ ")
		b.WriteString(prompt + p.input.View())
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

// Stats returns task statistics.
func (p *TaskPane) Stats() (done, total int) {
	for _, task := range p.tasks {
		if task.Done {
			done++
		}
	}
	return done, len(p.tasks)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// formatPriorityBadge returns a styled priority indicator.
// Returns: "!" for high, "~" for medium, " " for low/none
func (p *TaskPane) formatPriorityBadge(priority storage.Priority) string {
	switch priority {
	case storage.PriorityHigh:
		return p.styles.PriorityHighStyle.Render("!")
	case storage.PriorityMedium:
		return p.styles.PriorityMediumStyle.Render("~")
	default:
		return " " // space placeholder for alignment
	}
}

// formatDueDate returns a compact, styled due date indicator.
// Returns empty string if no due date, otherwise: "!" (overdue), "T" (today),
// "+1" (tomorrow), "3d" (days), "2w" (weeks), ">1m" (over a month).
func (p *TaskPane) formatDueDate(dueDate *time.Time) string {
	if dueDate == nil {
		return ""
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	due := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())

	days := int(due.Sub(today).Hours() / 24)

	switch {
	case days < 0:
		return p.styles.DueDateOverdueStyle.Render("!")
	case days == 0:
		return p.styles.DueDateTodayStyle.Render("T")
	case days == 1:
		return p.styles.DueDateFutureStyle.Render("+1")
	case days <= 7:
		return p.styles.DueDateFutureStyle.Render(fmt.Sprintf("%dd", days))
	case days <= 30:
		weeks := days / 7
		return p.styles.DueDateFutureStyle.Render(fmt.Sprintf("%dw", weeks))
	default:
		return p.styles.DueDateFutureStyle.Render(">1m")
	}
}
