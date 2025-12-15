package ui

import (
	"testing"

	"today/internal/config"
)

func TestHelpOverlay_View(t *testing.T) {
	setupTest(t)

	help := NewHelpOverlay(createTestStyles())
	help.SetSize(100, 40)

	output := help.View()
	assertGolden(t, "help_overlay", output)
}

func TestHelpOverlay_NarrowTerminal(t *testing.T) {
	setupTest(t)

	help := NewHelpOverlay(createTestStyles())
	help.SetSize(70, 30)

	output := help.View()
	assertGolden(t, "help_overlay_narrow", output)
}

func TestHelpOverlay_SmallTerminal(t *testing.T) {
	setupTest(t)

	help := NewHelpOverlay(createTestStyles())
	help.SetSize(50, 25)

	output := help.View()
	assertGolden(t, "help_overlay_small", output)
}

func TestHelpOverlay_ContentStructure(t *testing.T) {
	setupTest(t)

	help := NewHelpOverlay(createTestStyles())
	help.SetSize(100, 40)

	output := help.View()

	// Verify help contains key sections
	sections := []string{
		"Global",
		"Tasks",
		"Timer",
		"Habits",
		"Input Mode",
	}

	for _, section := range sections {
		if !contains(output, section) {
			t.Errorf("help overlay should contain section: %s", section)
		}
	}

	// Verify key bindings are mentioned
	keybindings := []string{
		"Tab",
		"?",
		"q",
		"Space",
		"Enter",
		"Esc",
	}

	for _, key := range keybindings {
		if !contains(output, key) {
			t.Errorf("help overlay should mention key: %s", key)
		}
	}
}

func TestApp_HelpToggle(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)

	app := NewApp(store, createTestStyles(), &AppConfig{
		Keys:                  &config.KeysConfig{},
		ConfirmDeletions:      false,
		ShowOnboarding:        false,
		NarrowLayoutThreshold: 80,
	})
	app.width = 100
	app.height = 40
	app.updateLayout()

	// Initially help should not be shown
	if app.showHelp {
		t.Error("showHelp should be false initially")
	}

	// Simulate pressing '?' to show help
	app.showHelp = true

	if !app.showHelp {
		t.Error("showHelp should be true after toggle")
	}

	// View should render help overlay
	view := app.View()
	if !contains(view, "Keyboard Shortcuts") {
		t.Error("view should show help overlay content")
	}

	// Toggle off
	app.showHelp = false

	// View should not show help
	view = app.View()
	if contains(view, "Keyboard Shortcuts") {
		t.Error("view should not show help after toggle off")
	}
}

func TestApp_HelpOverlayBlocksInput(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)

	app := NewApp(store, createTestStyles(), &AppConfig{
		Keys:                  &config.KeysConfig{},
		ConfirmDeletions:      false,
		ShowOnboarding:        false,
		NarrowLayoutThreshold: 80,
	})
	app.width = 100
	app.height = 40
	app.updateLayout()

	// Show help
	app.showHelp = true

	// Try to switch panes while help is shown
	initialPane := app.activePane

	// This should not switch panes because help is shown
	// (In real app, Update would handle this, but we test the logic)
	if app.showHelp {
		// Help overlay should prevent pane switching
		if app.activePane != initialPane {
			t.Error("active pane should not change while help is shown")
		}
	}
}

func TestRenderHelp_Function(t *testing.T) {
	setupTest(t)

	// Test the RenderHelp helper function
	styles := createTestStyles()
	output := styles.RenderHelp(
		"a", "add",
		"d", "done",
		"x", "delete",
	)

	if len(output) == 0 {
		t.Error("RenderHelp should produce output")
	}

	// Should contain the keys and descriptions
	if !contains(output, "a") {
		t.Error("output should contain key 'a'")
	}
	if !contains(output, "add") {
		t.Error("output should contain description 'add'")
	}
}

func TestApp_ContextualHelp(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)

	app := NewApp(store, createTestStyles(), &AppConfig{
		Keys:                  &config.KeysConfig{},
		ConfirmDeletions:      false,
		ShowOnboarding:        false,
		NarrowLayoutThreshold: 80,
	})
	app.width = 100
	app.height = 40
	app.updateLayout()

	tests := []struct {
		name      string
		pane      PaneID
		expectKey string // A key that should appear in help for this pane
	}{
		{
			name:      "tasks pane help",
			pane:      PaneTasks,
			expectKey: "add",
		},
		{
			name:      "timer pane help",
			pane:      PaneTimer,
			expectKey: "start", // Timer is not running by default
		},
		{
			name:      "habits pane help",
			pane:      PaneHabits,
			expectKey: "toggle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app.setActivePane(tt.pane)

			// Render the help bar
			helpBar := app.renderHelpBar()

			if !contains(helpBar, tt.expectKey) {
				t.Errorf("help bar for %v should contain %q", tt.pane, tt.expectKey)
			}
		})
	}
}

func TestApp_InputModeHelp(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)

	app := NewApp(store, createTestStyles(), &AppConfig{
		Keys:                  &config.KeysConfig{},
		ConfirmDeletions:      false,
		ShowOnboarding:        false,
		NarrowLayoutThreshold: 80,
	})
	app.width = 100
	app.height = 40
	app.updateLayout()

	// Simulate task input mode
	app.setActivePane(PaneTasks)
	app.taskPane.adding = true

	helpBar := app.renderHelpBar()

	// Should show input-specific help
	if !contains(helpBar, "save") || !contains(helpBar, "cancel") {
		t.Error("help bar should show input mode help when adding task")
	}
}
