// Package ui provides terminal user interface components for the today app.
// This file contains tests for mouse interaction support.
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"today/internal/config"
	"today/internal/storage"
)

// TestApp_MousePaneSwitching verifies clicking on panes switches focus.
func TestApp_MousePaneSwitching(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	cfg := &AppConfig{
		Keys:                  &config.KeysConfig{},
		NarrowLayoutThreshold: 80,
	}

	app := NewApp(store, styles, cfg)

	// Set wide width to enable 3-pane layout
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	// Initially tasks pane should be active
	if app.activePane != PaneTasks {
		t.Errorf("Expected initial pane to be Tasks, got %v", app.activePane)
	}

	// Click on timer pane area (middle of screen)
	// With 120 width, timer starts around 33% = ~40
	mouseMsg := tea.MouseMsg{
		X:      50,
		Y:      5,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}
	app.Update(mouseMsg)

	if app.activePane != PaneTimer {
		t.Errorf("Expected pane to be Timer after click, got %v", app.activePane)
	}

	// Click on habits pane area (right side)
	mouseMsg.X = 90
	app.Update(mouseMsg)

	if app.activePane != PaneHabits {
		t.Errorf("Expected pane to be Habits after click, got %v", app.activePane)
	}

	// Click on tasks pane area (left side)
	mouseMsg.X = 10
	app.Update(mouseMsg)

	if app.activePane != PaneTasks {
		t.Errorf("Expected pane to be Tasks after click, got %v", app.activePane)
	}
}

// TestApp_MouseClosesHelp verifies clicking closes help overlay.
func TestApp_MouseClosesHelp(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	app := NewApp(store, styles, &AppConfig{
		Keys:                  &config.KeysConfig{},
		ConfirmDeletions:      false,
		ShowOnboarding:        false,
		NarrowLayoutThreshold: 80,
	})

	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	// Open help
	app.showHelp = true

	// Click anywhere
	mouseMsg := tea.MouseMsg{
		X:      50,
		Y:      15,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}
	app.Update(mouseMsg)

	if app.showHelp {
		t.Error("Expected help to close after click")
	}
}

// TestTaskPane_MouseSelection verifies clicking selects tasks.
func TestTaskPane_MouseSelection(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	pane := NewTaskPane(store, styles)

	// Add some tasks
	store.AddTask("Task 1", "", storage.PriorityNone, nil)
	store.AddTask("Task 2", "", storage.PriorityNone, nil)
	store.AddTask("Task 3", "", storage.PriorityNone, nil)

	// Load tasks
	tasks, _ := store.LoadTasks()
	pane.setTasks(tasks.Tasks)

	pane.SetSize(40, 20)
	pane.SetFocused(true)

	// Initial cursor should be at 0
	if pane.cursor != 0 {
		t.Errorf("Expected initial cursor 0, got %d", pane.cursor)
	}

	// Click on second task (row 2 + header rows)
	mouseMsg := tea.MouseMsg{
		X:      10,
		Y:      3, // header (2) + task row 1
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}
	pane.Update(mouseMsg)

	if pane.cursor != 1 {
		t.Errorf("Expected cursor 1 after click, got %d", pane.cursor)
	}

	// Click on third task
	mouseMsg.Y = 4
	pane.Update(mouseMsg)

	if pane.cursor != 2 {
		t.Errorf("Expected cursor 2 after click, got %d", pane.cursor)
	}
}

// TestTaskPane_MouseScroll verifies scroll wheel navigates tasks.
func TestTaskPane_MouseScroll(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	pane := NewTaskPane(store, styles)

	// Add some tasks
	store.AddTask("Task 1", "", storage.PriorityNone, nil)
	store.AddTask("Task 2", "", storage.PriorityNone, nil)
	store.AddTask("Task 3", "", storage.PriorityNone, nil)

	// Load tasks
	tasks, _ := store.LoadTasks()
	pane.setTasks(tasks.Tasks)

	pane.SetSize(40, 20)
	pane.SetFocused(true)

	// Initial cursor at 0
	if pane.cursor != 0 {
		t.Errorf("Expected initial cursor 0, got %d", pane.cursor)
	}

	// Scroll down
	mouseMsg := tea.MouseMsg{
		Button: tea.MouseButtonWheelDown,
	}
	pane.Update(mouseMsg)

	if pane.cursor != 1 {
		t.Errorf("Expected cursor 1 after scroll down, got %d", pane.cursor)
	}

	// Scroll down again
	pane.Update(mouseMsg)

	if pane.cursor != 2 {
		t.Errorf("Expected cursor 2 after second scroll down, got %d", pane.cursor)
	}

	// Scroll up
	mouseMsg.Button = tea.MouseButtonWheelUp
	pane.Update(mouseMsg)

	if pane.cursor != 1 {
		t.Errorf("Expected cursor 1 after scroll up, got %d", pane.cursor)
	}
}

// TestHabitsPane_MouseSelection verifies clicking selects habits.
func TestHabitsPane_MouseSelection(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	pane := NewHabitsPane(store, styles)

	// Add some habits
	store.AddHabit("Exercise", "🏃")
	store.AddHabit("Read", "📚")
	store.AddHabit("Meditate", "🧘")

	// Load habits
	habits, _ := store.LoadHabits()
	pane.setHabitStore(habits)

	pane.SetSize(40, 20)
	pane.SetFocused(true)

	// Initial cursor should be at 0
	if pane.cursor != 0 {
		t.Errorf("Expected initial cursor 0, got %d", pane.cursor)
	}

	// Click on second habit (header rows = 5)
	mouseMsg := tea.MouseMsg{
		X:      10,
		Y:      6, // header (5) + habit row 1
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	}
	pane.Update(mouseMsg)

	if pane.cursor != 1 {
		t.Errorf("Expected cursor 1 after click, got %d", pane.cursor)
	}
}

// TestApp_PaneAtPosition verifies pane position calculation.
func TestApp_PaneAtPosition(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	cfg := &AppConfig{
		Keys:                  &config.KeysConfig{},
		NarrowLayoutThreshold: 80,
	}

	app := NewApp(store, styles, cfg)

	// Set wide width
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	tests := []struct {
		x        int
		expected PaneID
	}{
		{0, PaneTasks},
		{10, PaneTasks},
		{50, PaneTimer},
		{90, PaneHabits},
		{110, PaneHabits},
	}

	for _, tc := range tests {
		got := app.paneAtPosition(tc.x)
		if got != tc.expected {
			t.Errorf("paneAtPosition(%d) = %v, want %v", tc.x, got, tc.expected)
		}
	}
}
