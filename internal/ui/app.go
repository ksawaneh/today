// Package ui provides terminal user interface components for the today app.
// This file contains the main App model which coordinates all panes and
// routes messages using the Bubble Tea architecture.
package ui

import (
	"fmt"
	"strings"
	"time"

	"today/internal/config"
	"today/internal/storage"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PaneID identifies each pane in the application.
type PaneID int

const (
	PaneTasks PaneID = iota
	PaneTimer
	PaneHabits
)

// LayoutMode determines how panes are arranged based on terminal width.
type LayoutMode int

const (
	// LayoutWide shows all three panes side-by-side.
	LayoutWide LayoutMode = iota
	// LayoutNarrow shows only the focused pane with a tab bar.
	LayoutNarrow
)

// AppConfig holds user configuration for the app behavior.
type AppConfig struct {
	Keys                  *config.KeysConfig
	ConfirmDeletions      bool
	ShowOnboarding        bool
	NarrowLayoutThreshold int
}

// App is the main application model that coordinates all panes.
type App struct {
	storage     *storage.Storage
	styles      *Styles
	config      *AppConfig
	taskPane    *TaskPane
	timerPane   *TimerPane
	habitsPane  *HabitsPane
	helpOverlay *HelpOverlay
	undoManager *UndoManager
	undoBusy    bool
	confirmDel  *confirmDeleteState
	activePane  PaneID
	layoutMode  LayoutMode
	showHelp    bool
	showWelcome bool
	width       int
	height      int
	status      string
	statusErr   bool
	statusUntil time.Time
	quitting    bool

	// Key bindings
	keys     GlobalKeyMap
	helpKeys HelpKeyMap

	// Pane positions for mouse click detection (x coordinates)
	tasksPaneStart  int
	tasksPaneEnd    int
	timerPaneStart  int
	timerPaneEnd    int
	habitsPaneStart int
	habitsPaneEnd   int
	contentTop      int // Y coordinate where content starts
}

type confirmDeleteState struct {
	title string
	body  string
	cmd   tea.Cmd
}

// NewApp creates a new application. Data loading is deferred to Init()
// to keep the constructor non-blocking.
func NewApp(store *storage.Storage, styles *Styles, cfg *AppConfig) *App {
	// Use default config if nil
	if cfg == nil {
		cfg = &AppConfig{
			Keys:                  &config.KeysConfig{},
			ConfirmDeletions:      true,
			ShowOnboarding:        true,
			NarrowLayoutThreshold: 80,
		}
	}
	if cfg.Keys == nil {
		cfg.Keys = &config.KeysConfig{}
	}

	// Create panes with config-aware key bindings
	taskPane := NewTaskPaneWithKeys(store, styles, cfg.Keys)
	timerPane := NewTimerPaneWithKeys(store, styles, cfg.Keys)
	habitsPane := NewHabitsPaneWithKeys(store, styles, cfg.Keys)
	helpOverlay := NewHelpOverlay(styles)

	// Determine if we should show welcome screen
	showWelcome := cfg.ShowOnboarding && isFirstRun(store)

	app := &App{
		storage:     store,
		styles:      styles,
		config:      cfg,
		taskPane:    taskPane,
		timerPane:   timerPane,
		habitsPane:  habitsPane,
		helpOverlay: helpOverlay,
		undoManager: NewUndoManager(),
		activePane:  PaneTasks,
		showHelp:    false,
		showWelcome: showWelcome,
		keys:        NewGlobalKeyMap(cfg.Keys),
		helpKeys:    DefaultHelpKeyMap(),
	}

	// Set initial focus
	taskPane.SetFocused(true)
	timerPane.SetFocused(false)
	habitsPane.SetFocused(false)

	return app
}

// isFirstRun checks if this appears to be the first time running the app.
// We detect this by checking if data files exist and are empty.
func isFirstRun(store *storage.Storage) bool {
	tasks, err := store.LoadTasks()
	if err != nil {
		return false
	}
	if len(tasks.Tasks) > 0 {
		return false
	}

	habits, err := store.LoadHabits()
	if err != nil {
		return false
	}
	if len(habits.Habits) > 0 || len(habits.Logs) > 0 {
		return false
	}

	timer, err := store.LoadTimer()
	if err != nil {
		return false
	}
	if len(timer.Entries) > 0 || timer.Current != nil {
		return false
	}

	return true
}

// tickMsg is sent periodically for time updates.
type tickMsg time.Time

// tickCmd returns a command that sends a tick every second.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Init initializes the app and loads all data asynchronously.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		tickCmd(),
		a.taskPane.LoadTasksCmd(),
		a.timerPane.LoadTimerCmd(),
		a.habitsPane.LoadHabitsCmd(),
	)
}

// Update handles all messages and routes them appropriately.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Route async messages to all panes first (before key handling).
	// This ensures storage operation results are processed regardless
	// of which pane is active.
	switch msg := msg.(type) {
	case tasksLoadedMsg:
		if msg.err != nil {
			a.SetStatus("Tasks: "+msg.err.Error(), true)
		}
		cmd := a.taskPane.Update(msg)
		return a, cmd

	case taskAddedMsg:
		if msg.err != nil {
			a.SetStatus("Add task: "+msg.err.Error(), true)
		}
		cmd := a.taskPane.Update(msg)
		return a, cmd

	case taskCompletedMsg:
		if msg.err != nil {
			a.SetStatus("Complete task: "+msg.err.Error(), true)
		} else {
			// Push undo action on successful completion
			a.undoManager.Push(NewCompleteTaskAction(a.storage, msg.id, msg.text))
		}
		cmd := a.taskPane.Update(msg)
		return a, cmd

	case taskUncompletedMsg:
		if msg.err != nil {
			a.SetStatus("Uncomplete task: "+msg.err.Error(), true)
		} else {
			// Push undo action on successful uncompletion
			a.undoManager.Push(NewUncompleteTaskAction(a.storage, msg.id, msg.text))
		}
		cmd := a.taskPane.Update(msg)
		return a, cmd

	case taskDeletedMsg:
		if msg.err != nil {
			a.SetStatus("Delete task: "+msg.err.Error(), true)
		} else if msg.task != nil {
			// Push undo action on successful deletion (only if task was captured)
			a.undoManager.Push(NewDeleteTaskAction(a.storage, *msg.task))
		}
		cmd := a.taskPane.Update(msg)
		return a, cmd

	case timerLoadedMsg:
		if msg.err != nil {
			a.SetStatus("Timer: "+msg.err.Error(), true)
		}
		cmd := a.timerPane.Update(msg)
		return a, cmd

	case timerStartedMsg:
		if msg.err != nil {
			a.SetStatus("Start timer: "+msg.err.Error(), true)
		}
		cmd := a.timerPane.Update(msg)
		return a, cmd

	case timerStoppedMsg:
		if msg.err != nil {
			a.SetStatus("Stop timer: "+msg.err.Error(), true)
		}
		cmd := a.timerPane.Update(msg)
		return a, cmd

	case habitsLoadedMsg:
		if msg.err != nil {
			a.SetStatus("Habits: "+msg.err.Error(), true)
		}
		cmd := a.habitsPane.Update(msg)
		return a, cmd

	case habitAddedMsg:
		if msg.err != nil {
			a.SetStatus("Add habit: "+msg.err.Error(), true)
		}
		cmd := a.habitsPane.Update(msg)
		return a, cmd

	case habitToggledMsg:
		if msg.err != nil {
			a.SetStatus("Toggle habit: "+msg.err.Error(), true)
		} else {
			// Push undo action on successful toggle
			a.undoManager.Push(NewToggleHabitAction(a.storage, msg.id, msg.name, msg.date, msg.wasCompleted))
		}
		cmd := a.habitsPane.Update(msg)
		return a, cmd

	case habitDeletedMsg:
		if msg.err != nil {
			a.SetStatus("Delete habit: "+msg.err.Error(), true)
		} else if msg.habit != nil {
			// Push undo action on successful deletion (only if habit was captured)
			a.undoManager.Push(NewDeleteHabitAction(a.storage, *msg.habit, msg.logs))
		}
		cmd := a.habitsPane.Update(msg)
		return a, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if a.showWelcome {
			a.showWelcome = false
			return a, nil
		}

		if a.confirmDel != nil {
			switch msg.String() {
			case "y", "Y", "enter":
				cmd := a.confirmDel.cmd
				a.confirmDel = nil
				return a, cmd
			case "n", "N", "esc":
				a.confirmDel = nil
				a.SetStatus("Canceled", false)
				return a, nil
			default:
				return a, nil
			}
		}

		// Help overlay takes priority
		if a.showHelp {
			if key.Matches(msg, a.helpKeys.Close) {
				a.showHelp = false
				return a, nil
			}
			return a, nil
		}

		// Check if any pane is in input mode
		inInputMode := a.taskPane.IsAdding() || a.timerPane.IsSwitching() || a.habitsPane.IsAdding()

		if !inInputMode {
			// Confirm deletions (tasks/habits) if enabled.
			if a.config.ConfirmDeletions {
				switch a.activePane {
				case PaneTasks:
					if key.Matches(msg, a.taskPane.keys.Delete) {
						if len(a.taskPane.tasks) == 0 || a.taskPane.cursor < 0 || a.taskPane.cursor >= len(a.taskPane.tasks) {
							a.SetStatus("No task selected", true)
							return a, nil
						}
						task := a.taskPane.tasks[a.taskPane.cursor]
						a.confirmDel = &confirmDeleteState{
							title: "Delete task?",
							body:  truncateText(task.Text, 60),
							cmd:   deleteTaskCmd(a.storage, task.ID),
						}
						return a, nil
					}
				case PaneHabits:
					if key.Matches(msg, a.habitsPane.keys.Delete) {
						if len(a.habitsPane.habitStore.Habits) == 0 || a.habitsPane.cursor < 0 || a.habitsPane.cursor >= len(a.habitsPane.habitStore.Habits) {
							a.SetStatus("No habit selected", true)
							return a, nil
						}
						habit := a.habitsPane.habitStore.Habits[a.habitsPane.cursor]
						a.confirmDel = &confirmDeleteState{
							title: "Delete habit?",
							body:  truncateText(habit.Name, 60),
							cmd:   deleteHabitCmd(a.storage, habit.ID),
						}
						return a, nil
					}
				}
			}

			// Global keys only when not in input mode
			switch {
			case key.Matches(msg, a.keys.Quit):
				a.quitting = true
				return a, tea.Quit

			case key.Matches(msg, a.keys.Help):
				a.showHelp = true
				return a, nil

			case key.Matches(msg, a.keys.NextPane):
				a.switchPane()
				return a, nil

			case key.Matches(msg, a.keys.Pane1):
				a.setActivePane(PaneTasks)
				return a, nil

			case key.Matches(msg, a.keys.Pane2):
				a.setActivePane(PaneTimer)
				return a, nil

			case key.Matches(msg, a.keys.Pane3):
				a.setActivePane(PaneHabits)
				return a, nil

			case key.Matches(msg, a.keys.Undo):
				if a.undoBusy {
					a.SetStatus("Undo: busy", true)
					return a, nil
				}
				a.undoBusy = true
				return a, undoCmd(a.undoManager)

			case key.Matches(msg, a.keys.Redo):
				if a.undoBusy {
					a.SetStatus("Redo: busy", true)
					return a, nil
				}
				a.undoBusy = true
				return a, redoCmd(a.undoManager)
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.updateLayout()
		return a, nil

	case tea.MouseMsg:
		if a.showWelcome {
			if msg.Action == tea.MouseActionPress {
				a.showWelcome = false
			}
			return a, nil
		}

		if a.confirmDel != nil {
			if msg.Action == tea.MouseActionPress {
				a.confirmDel = nil
				a.SetStatus("Canceled", false)
			}
			return a, nil
		}

		// Ignore mouse events when help overlay is shown
		if a.showHelp {
			// Any click closes help
			if msg.Action == tea.MouseActionPress {
				a.showHelp = false
			}
			return a, nil
		}

		// Handle mouse events
		switch msg.Action {
		case tea.MouseActionPress:
			// In narrow mode, check for tab bar clicks
			if a.layoutMode == LayoutNarrow && msg.Y == a.contentTop-1 {
				// Tab bar click - determine which tab based on X position
				// Tabs are roughly evenly distributed
				tabWidth := a.width / 3
				if msg.X < tabWidth {
					a.setActivePane(PaneTasks)
				} else if msg.X < tabWidth*2 {
					a.setActivePane(PaneTimer)
				} else {
					a.setActivePane(PaneHabits)
				}
				return a, nil
			}

			// Determine which pane was clicked (in wide mode)
			clickedPane := a.paneAtPosition(msg.X)
			if clickedPane >= 0 && clickedPane != a.activePane {
				a.setActivePane(clickedPane)
			}

			// Forward click to active pane with adjusted coordinates
			if msg.Y >= a.contentTop {
				localMsg := msg
				localMsg.Y = msg.Y - a.contentTop
				// Adjust X for non-tasks panes in wide mode
				if a.layoutMode == LayoutWide {
					switch a.activePane {
					case PaneTimer:
						localMsg.X = msg.X - a.timerPaneStart
					case PaneHabits:
						localMsg.X = msg.X - a.habitsPaneStart
					}
				}

				switch a.activePane {
				case PaneTasks:
					cmd := a.taskPane.Update(localMsg)
					return a, cmd
				case PaneTimer:
					cmd := a.timerPane.Update(localMsg)
					return a, cmd
				case PaneHabits:
					cmd := a.habitsPane.Update(localMsg)
					return a, cmd
				}
			}

		case tea.MouseActionMotion:
			// Ignore motion events for now

		}

		// Handle scroll wheel
		if msg.Button == tea.MouseButtonWheelUp || msg.Button == tea.MouseButtonWheelDown {
			// Forward scroll to active pane
			localMsg := msg
			localMsg.Y = msg.Y - a.contentTop

			switch a.activePane {
			case PaneTasks:
				cmd := a.taskPane.Update(localMsg)
				return a, cmd
			case PaneHabits:
				cmd := a.habitsPane.Update(localMsg)
				return a, cmd
			}
		}

		return a, nil

	case tickMsg:
		if a.status != "" && !a.statusUntil.IsZero() && time.Now().After(a.statusUntil) {
			a.status = ""
			a.statusErr = false
			a.statusUntil = time.Time{}
		}
		return a, tickCmd()
	}

	switch msg := msg.(type) {
	case undoResultMsg:
		a.undoBusy = false
		if msg.err != nil {
			a.SetStatus("Undo failed: "+msg.err.Error(), true)
		} else if msg.desc != "" {
			a.SetStatus("Undid: "+msg.desc, false)
		} else {
			a.SetStatus("Nothing to undo", false)
		}
		return a, tea.Batch(
			a.taskPane.LoadTasksCmd(),
			a.habitsPane.LoadHabitsCmd(),
		)

	case redoResultMsg:
		a.undoBusy = false
		if msg.err != nil {
			a.SetStatus("Redo failed: "+msg.err.Error(), true)
		} else if msg.desc != "" {
			a.SetStatus("Redid: "+msg.desc, false)
		} else {
			a.SetStatus("Nothing to redo", false)
		}
		return a, tea.Batch(
			a.taskPane.LoadTasksCmd(),
			a.habitsPane.LoadHabitsCmd(),
		)
	}

	// Forward to active pane (only if help is not shown)
	if !a.showHelp {
		switch a.activePane {
		case PaneTasks:
			cmd := a.taskPane.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case PaneTimer:
			cmd := a.timerPane.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case PaneHabits:
			cmd := a.habitsPane.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	return a, tea.Batch(cmds...)
}

// switchPane cycles through panes.
func (a *App) switchPane() {
	switch a.activePane {
	case PaneTasks:
		a.setActivePane(PaneTimer)
	case PaneTimer:
		a.setActivePane(PaneHabits)
	case PaneHabits:
		a.setActivePane(PaneTasks)
	}
}

// setActivePane sets the active pane and updates focus states.
func (a *App) setActivePane(pane PaneID) {
	a.activePane = pane

	a.taskPane.SetFocused(pane == PaneTasks)
	a.timerPane.SetFocused(pane == PaneTimer)
	a.habitsPane.SetFocused(pane == PaneHabits)
}

// paneAtPosition returns which pane is at the given X coordinate.
// Returns -1 if no pane is at that position.
func (a *App) paneAtPosition(x int) PaneID {
	if a.layoutMode == LayoutNarrow {
		// In narrow mode, return the active pane
		return a.activePane
	}

	if x >= a.tasksPaneStart && x < a.tasksPaneEnd {
		return PaneTasks
	}
	if x >= a.timerPaneStart && x < a.timerPaneEnd {
		return PaneTimer
	}
	if x >= a.habitsPaneStart && x < a.habitsPaneEnd {
		return PaneHabits
	}
	return -1
}

// updateLayout recalculates pane sizes based on terminal dimensions.
func (a *App) updateLayout() {
	// Leave room for title bar (2) and help bar (1)
	contentHeight := a.height - 4
	if contentHeight < 10 {
		contentHeight = 10
	}

	// Content starts after title bar (1 line title + 1 line space)
	a.contentTop = 1

	// Update help overlay size
	a.helpOverlay.SetSize(a.width, a.height)

	totalWidth := a.width - 4

	// Determine layout mode based on configured threshold
	threshold := a.config.NarrowLayoutThreshold
	if threshold <= 0 {
		threshold = 80 // Default threshold
	}

	if a.width < threshold {
		// Narrow mode: single focused pane with tab bar
		a.layoutMode = LayoutNarrow

		// In narrow mode, leave room for tab bar (1 line)
		narrowHeight := contentHeight - 1
		if narrowHeight < 8 {
			narrowHeight = 8
		}

		// Give full width to all panes (only focused one will be shown)
		paneWidth := totalWidth
		if paneWidth < 20 {
			paneWidth = 20
		}

		a.taskPane.SetSize(paneWidth, narrowHeight)
		a.timerPane.SetSize(paneWidth, narrowHeight)
		a.habitsPane.SetSize(paneWidth, narrowHeight)

		// In narrow mode, all panes occupy the same space
		a.tasksPaneStart = 0
		a.tasksPaneEnd = a.width
		a.timerPaneStart = 0
		a.timerPaneEnd = a.width
		a.habitsPaneStart = 0
		a.habitsPaneEnd = a.width
		// Content starts after tab bar in narrow mode
		a.contentTop = 2
	} else {
		// Wide mode: three panes side-by-side
		a.layoutMode = LayoutWide

		var tasksWidth, timerWidth, habitsWidth int
		if totalWidth < 120 {
			// Medium: balanced three-column
			tasksWidth = (totalWidth * 33) / 100
			timerWidth = (totalWidth * 28) / 100
			habitsWidth = totalWidth - tasksWidth - timerWidth - 2
		} else {
			// Wide: comfortable three-column with max widths
			tasksWidth = min((totalWidth*35)/100, 50)
			timerWidth = min((totalWidth*28)/100, 40)
			habitsWidth = min(totalWidth-tasksWidth-timerWidth-2, 55)
		}

		a.taskPane.SetSize(tasksWidth, contentHeight)
		a.timerPane.SetSize(timerWidth, contentHeight)
		a.habitsPane.SetSize(habitsWidth, contentHeight)

		// Calculate pane positions (with 1 space gaps between panes)
		a.tasksPaneStart = 0
		a.tasksPaneEnd = tasksWidth
		a.timerPaneStart = tasksWidth + 1
		a.timerPaneEnd = a.timerPaneStart + timerWidth
		a.habitsPaneStart = a.timerPaneEnd + 1
		a.habitsPaneEnd = a.habitsPaneStart + habitsWidth
	}
}

// View renders the entire app.
func (a *App) View() string {
	if a.quitting {
		return a.renderGoodbye()
	}

	if a.showWelcome {
		return a.renderWelcome()
	}

	if a.confirmDel != nil {
		return a.renderConfirmDelete()
	}

	// Show help overlay if active
	if a.showHelp {
		return a.helpOverlay.View()
	}

	var b strings.Builder

	// Title bar
	titleBar := a.renderTitleBar()
	b.WriteString(titleBar)
	b.WriteString("\n")

	// Main content - switch based on layout mode
	switch a.layoutMode {
	case LayoutNarrow:
		b.WriteString(a.renderNarrowContent())
	default:
		b.WriteString(a.renderWideContent())
	}
	b.WriteString("\n")

	// Help bar
	helpBar := a.renderHelpBar()
	b.WriteString(helpBar)

	return b.String()
}

func (a *App) renderWelcome() string {
	overlayWidth := 60
	if a.width > 0 {
		overlayWidth = min(60, max(20, a.width-4))
	}

	overlayStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(a.styles.ColorPrimary).
		Padding(1, 2).
		Width(overlayWidth)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(a.styles.ColorPrimary).
		MarginBottom(1)

	bodyStyle := lipgloss.NewStyle().
		Foreground(a.styles.ColorText)

	mutedStyle := lipgloss.NewStyle().
		Foreground(a.styles.ColorTextMuted).
		Italic(true)

	var b strings.Builder
	b.WriteString(titleStyle.Render("Welcome to today"))
	b.WriteString("\n\n")
	b.WriteString(bodyStyle.Render("Tab switches panes. ? opens help.\n"))
	b.WriteString(bodyStyle.Render("Add your first task with 'a'.\n"))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("Press any key to continue"))

	content := overlayStyle.Render(b.String())
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
}

func (a *App) renderConfirmDelete() string {
	overlayWidth := 60
	if a.width > 0 {
		overlayWidth = min(60, max(20, a.width-4))
	}

	overlayStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(a.styles.ColorDanger).
		Padding(1, 2).
		Width(overlayWidth)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(a.styles.ColorDanger).
		MarginBottom(1)

	bodyStyle := lipgloss.NewStyle().
		Foreground(a.styles.ColorText)

	hintStyle := lipgloss.NewStyle().
		Foreground(a.styles.ColorTextMuted)

	var b strings.Builder
	b.WriteString(titleStyle.Render(a.confirmDel.title))
	b.WriteString("\n\n")
	b.WriteString(bodyStyle.Render(a.confirmDel.body))
	b.WriteString("\n\n")
	b.WriteString(hintStyle.Render("[y/enter] delete    [n/esc] cancel"))

	content := overlayStyle.Render(b.String())
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, content)
}

// renderWideContent renders all three panes side by side.
func (a *App) renderWideContent() string {
	tasksView := a.taskPane.View()
	timerView := a.timerPane.View()
	habitsView := a.habitsPane.View()

	return lipgloss.JoinHorizontal(lipgloss.Top, tasksView, " ", timerView, " ", habitsView)
}

// renderNarrowContent renders the focused pane with a tab bar.
func (a *App) renderNarrowContent() string {
	var b strings.Builder

	// Tab bar at top
	b.WriteString(a.renderPaneTabs())
	b.WriteString("\n")

	// Only render the active pane
	switch a.activePane {
	case PaneTasks:
		b.WriteString(a.taskPane.View())
	case PaneTimer:
		b.WriteString(a.timerPane.View())
	case PaneHabits:
		b.WriteString(a.habitsPane.View())
	}

	return b.String()
}

// renderPaneTabs renders a tab bar showing available panes.
func (a *App) renderPaneTabs() string {
	// Tab labels
	tabs := []struct {
		id    PaneID
		label string
	}{
		{PaneTasks, "Tasks"},
		{PaneTimer, "Timer"},
		{PaneHabits, "Habits"},
	}

	// Create tab styles
	activeTabStyle := lipgloss.NewStyle().
		Foreground(a.styles.ColorPrimary).
		Bold(true)
	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(a.styles.ColorTextMuted)

	var parts []string
	for _, tab := range tabs {
		label := tab.label
		if tab.id == a.activePane {
			// Active tab: highlighted with brackets
			label = activeTabStyle.Render("[" + label + "]")
		} else {
			// Inactive tab: muted
			label = inactiveTabStyle.Render(" " + label + " ")
		}
		parts = append(parts, label)
	}

	// Center the tabs
	tabBar := strings.Join(parts, "  ")
	padding := (a.width - lipgloss.Width(tabBar)) / 2
	if padding > 0 {
		tabBar = strings.Repeat(" ", padding) + tabBar
	}

	return tabBar
}

// renderGoodbye shows a nice exit message with session summary.
func (a *App) renderGoodbye() string {
	tasksDone, tasksTotal := a.taskPane.Stats()
	habitsDone, habitsTotal := a.habitsPane.GetTodayCompletionRate()

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("  See you later!\n")
	b.WriteString("\n")

	if tasksTotal > 0 || habitsTotal > 0 {
		b.WriteString("  Today's progress:\n")
		if tasksTotal > 0 {
			pct := (tasksDone * 100) / tasksTotal
			b.WriteString(fmt.Sprintf("     Tasks:  %d/%d (%d%%)\n", tasksDone, tasksTotal, pct))
		}
		if habitsTotal > 0 {
			pct := (habitsDone * 100) / habitsTotal
			b.WriteString(fmt.Sprintf("     Habits: %d/%d (%d%%)\n", habitsDone, habitsTotal, pct))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderTitleBar creates the top title bar with stats and timer status.
func (a *App) renderTitleBar() string {
	title := a.styles.TitleStyle.Render(" today ")

	// Stats summary
	tasksDone, tasksTotal := a.taskPane.Stats()
	habitsDone, habitsTotal := a.habitsPane.GetTodayCompletionRate()

	var statsItems []string
	if tasksTotal > 0 {
		statsItems = append(statsItems, fmt.Sprintf("Tasks: %d/%d", tasksDone, tasksTotal))
	}
	if habitsTotal > 0 {
		statsItems = append(statsItems, fmt.Sprintf("Habits: %d/%d", habitsDone, habitsTotal))
	}
	stats := a.styles.StatLabelStyle.Render(strings.Join(statsItems, "  "))

	// Timer status indicator
	var timerStatus string
	if a.timerPane.IsRunning() {
		project := a.timerPane.GetCurrentProject()
		elapsed := formatDuration(a.timerPane.GetElapsed())
		// Truncate project name if too long
		if len(project) > 12 {
			project = project[:11] + "…"
		}
		timerStatus = a.styles.TimerRunningStyle.Render(fmt.Sprintf("▶ %s %s", project, elapsed))
	}

	// Current date/time
	now := time.Now()
	dateStr := now.Format("Mon Jan 2 · 15:04")
	date := a.styles.DateStyle.Render(dateStr)

	// Calculate spacing
	titleWidth := lipgloss.Width(title)
	statsWidth := lipgloss.Width(stats)
	timerWidth := lipgloss.Width(timerStatus)
	dateWidth := lipgloss.Width(date)

	usedWidth := titleWidth + statsWidth + timerWidth + dateWidth
	spacerWidth := a.width - usedWidth - 6
	if spacerWidth < 2 {
		spacerWidth = 2
	}

	// Build the title bar
	var parts []string
	parts = append(parts, title)

	if stats != "" {
		parts = append(parts, "  "+stats)
	}

	// Distribute spacing
	leftSpacer := strings.Repeat(" ", spacerWidth/2)
	rightSpacer := strings.Repeat(" ", spacerWidth-spacerWidth/2)

	parts = append(parts, leftSpacer)

	if timerStatus != "" {
		parts = append(parts, timerStatus)
	}

	parts = append(parts, rightSpacer)
	parts = append(parts, date)

	return strings.Join(parts, "")
}

// renderHelpBar creates the bottom help bar with context-sensitive hints.
func (a *App) renderHelpBar() string {
	if a.status != "" {
		if a.statusErr {
			return a.styles.ErrorStyle.Render(a.status)
		}
		return a.styles.StatusStyle.Render(a.status)
	}

	// Input mode help
	if a.taskPane.IsAdding() {
		return a.styles.RenderHelp(
			"enter", "save",
			"esc", "cancel",
		)
	}

	if a.timerPane.IsSwitching() {
		return a.styles.RenderHelp(
			"enter", "start",
			"esc", "cancel",
		)
	}

	if a.habitsPane.IsAdding() {
		return a.styles.RenderHelp(
			"enter", "next/save",
			"esc", "cancel",
		)
	}

	// Normal mode help based on active pane
	switch a.activePane {
	case PaneTasks:
		return a.styles.RenderHelp(
			"a", "add",
			"d", "done",
			"x", "del",
			"j/k", "nav",
			"tab", "pane",
			"?", "help",
		)
	case PaneTimer:
		if a.timerPane.IsRunning() {
			return a.styles.RenderHelp(
				"space", "stop",
				"s", "switch",
				"tab", "pane",
				"?", "help",
			)
		}
		return a.styles.RenderHelp(
			"space", "start",
			"s", "project",
			"tab", "pane",
			"?", "help",
		)
	case PaneHabits:
		return a.styles.RenderHelp(
			"a", "add",
			"space", "toggle",
			"x", "del",
			"j/k", "nav",
			"tab", "pane",
			"?", "help",
		)
	}

	return ""
}

// SetStatus sets a status message to display to the user.
func (a *App) SetStatus(msg string, isErr bool) {
	a.status = msg
	a.statusErr = isErr
	ttl := 5 * time.Second
	if isErr {
		ttl = 8 * time.Second
	}
	a.statusUntil = time.Now().Add(ttl)
}

// Run starts the Bubble Tea program with the given storage backend, styles, and config.
func Run(store *storage.Storage, styles *Styles, cfg *AppConfig) error {
	app := NewApp(store, styles, cfg)
	p := tea.NewProgram(app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // Enable mouse support
	)
	_, err := p.Run()
	return err
}
