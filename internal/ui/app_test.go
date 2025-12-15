// Package ui provides terminal user interface components for the today app.
// This file contains tests for the main App model, including layout behavior.
package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"today/internal/config"
)

// TestApp_LayoutModeTransitions verifies layout mode changes based on width.
func TestApp_LayoutModeTransitions(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	cfg := &AppConfig{
		Keys:                  &config.KeysConfig{},
		NarrowLayoutThreshold: 80,
	}

	app := NewApp(store, styles, cfg)

	tests := []struct {
		name         string
		width        int
		expectedMode LayoutMode
	}{
		{"Very narrow (40)", 40, LayoutNarrow},
		{"Narrow (60)", 60, LayoutNarrow},
		{"At threshold (79)", 79, LayoutNarrow},
		{"At threshold (80)", 80, LayoutWide},
		{"Wide (100)", 100, LayoutWide},
		{"Very wide (200)", 200, LayoutWide},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Send window size message
			msg := tea.WindowSizeMsg{Width: tc.width, Height: 30}
			app.Update(msg)

			if app.layoutMode != tc.expectedMode {
				t.Errorf("Width %d: expected layout mode %v, got %v",
					tc.width, tc.expectedMode, app.layoutMode)
			}
		})
	}
}

// TestApp_NarrowLayoutShowsOnlyActivePane verifies only focused pane is shown in narrow mode.
func TestApp_NarrowLayoutShowsOnlyActivePane(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	cfg := &AppConfig{
		Keys:                  &config.KeysConfig{},
		NarrowLayoutThreshold: 80,
	}

	app := NewApp(store, styles, cfg)

	// Set narrow width
	msg := tea.WindowSizeMsg{Width: 60, Height: 30}
	app.Update(msg)

	// Default active pane should be Tasks
	if app.activePane != PaneTasks {
		t.Errorf("Expected default active pane to be Tasks")
	}

	view := app.View()

	// In narrow mode, should show tab bar
	if !strings.Contains(view, "[Tasks]") {
		t.Error("Expected to see [Tasks] tab highlighted in narrow mode")
	}
	if !strings.Contains(view, "Timer") {
		t.Error("Expected to see Timer tab in narrow mode")
	}
	if !strings.Contains(view, "Habits") {
		t.Error("Expected to see Habits tab in narrow mode")
	}
}

// TestApp_WideLayoutShowsAllPanes verifies all panes are shown in wide mode.
func TestApp_WideLayoutShowsAllPanes(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	cfg := &AppConfig{
		Keys:                  &config.KeysConfig{},
		NarrowLayoutThreshold: 80,
	}

	app := NewApp(store, styles, cfg)

	// Set wide width
	msg := tea.WindowSizeMsg{Width: 120, Height: 30}
	app.Update(msg)

	if app.layoutMode != LayoutWide {
		t.Errorf("Expected LayoutWide at width 120, got %v", app.layoutMode)
	}

	view := app.View()

	// In wide mode, all pane titles should be visible (titles are uppercase)
	if !strings.Contains(view, "TASKS") {
		t.Error("Expected to see TASKS pane in wide mode")
	}
	if !strings.Contains(view, "TIMER") {
		t.Error("Expected to see TIMER pane in wide mode")
	}
	if !strings.Contains(view, "HABITS") {
		t.Error("Expected to see HABITS pane in wide mode")
	}
}

// TestApp_CustomThreshold verifies custom threshold configuration.
func TestApp_CustomThreshold(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()

	// Custom threshold of 100
	cfg := &AppConfig{
		Keys:                  &config.KeysConfig{},
		NarrowLayoutThreshold: 100,
	}

	app := NewApp(store, styles, cfg)

	// Width 90 should be narrow with threshold 100
	msg := tea.WindowSizeMsg{Width: 90, Height: 30}
	app.Update(msg)

	if app.layoutMode != LayoutNarrow {
		t.Errorf("Expected LayoutNarrow at width 90 with threshold 100, got %v", app.layoutMode)
	}

	// Width 100 should be wide
	msg = tea.WindowSizeMsg{Width: 100, Height: 30}
	app.Update(msg)

	if app.layoutMode != LayoutWide {
		t.Errorf("Expected LayoutWide at width 100 with threshold 100, got %v", app.layoutMode)
	}
}

// TestApp_PaneSwitchingInNarrowMode verifies pane switching works in narrow mode.
func TestApp_PaneSwitchingInNarrowMode(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	cfg := &AppConfig{
		Keys:                  &config.KeysConfig{},
		NarrowLayoutThreshold: 80,
	}

	app := NewApp(store, styles, cfg)

	// Set narrow width
	msg := tea.WindowSizeMsg{Width: 60, Height: 30}
	app.Update(msg)

	// Verify initial state
	if app.activePane != PaneTasks {
		t.Errorf("Expected initial pane to be Tasks")
	}

	// Switch to next pane
	app.switchPane()
	if app.activePane != PaneTimer {
		t.Errorf("Expected pane to be Timer after switch, got %v", app.activePane)
	}

	view := app.View()
	if !strings.Contains(view, "[Timer]") {
		t.Error("Expected [Timer] tab to be highlighted after switch")
	}

	// Switch again
	app.switchPane()
	if app.activePane != PaneHabits {
		t.Errorf("Expected pane to be Habits after second switch, got %v", app.activePane)
	}

	// Switch back to Tasks (cycles)
	app.switchPane()
	if app.activePane != PaneTasks {
		t.Errorf("Expected pane to cycle back to Tasks, got %v", app.activePane)
	}
}

// TestApp_LayoutModeAfterResize verifies layout adapts after resize.
func TestApp_LayoutModeAfterResize(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	cfg := &AppConfig{
		Keys:                  &config.KeysConfig{},
		NarrowLayoutThreshold: 80,
	}

	app := NewApp(store, styles, cfg)

	// Start wide
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	if app.layoutMode != LayoutWide {
		t.Error("Expected LayoutWide initially")
	}

	// Resize to narrow
	app.Update(tea.WindowSizeMsg{Width: 60, Height: 30})
	if app.layoutMode != LayoutNarrow {
		t.Error("Expected LayoutNarrow after resize")
	}

	// Resize back to wide
	app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	if app.layoutMode != LayoutWide {
		t.Error("Expected LayoutWide after resize back")
	}
}
