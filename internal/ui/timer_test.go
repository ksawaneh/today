package ui

import (
	"testing"
	"time"
)

func TestTimerPaneView_Empty(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	pane := NewTimerPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	// Load initial empty state
	timerStore, _ := store.LoadTimer()
	pane.setTimerStore(timerStore)

	output := pane.View()
	assertGolden(t, "timer_pane_empty", output)
}

func TestTimerPaneView_Running(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	// Start a timer
	store.StartTimer("Test Project")

	pane := NewTimerPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	// Load timer state
	timerStore, _ := store.LoadTimer()
	pane.setTimerStore(timerStore)

	// Note: The view will show elapsed time, which is dynamic
	// For deterministic testing, we'll just verify the pane renders without panic
	output := pane.View()
	if len(output) == 0 {
		t.Error("timer pane view should not be empty when timer is running")
	}
	// We can't use golden test here due to dynamic time, but we verify structure
	if !contains(output, "Test Project") {
		t.Error("view should contain project name")
	}
}

func TestTimerPaneView_WithHistory(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	// Create some timer entries
	store.StartTimer("Project A")
	time.Sleep(10 * time.Millisecond)
	store.StartTimer("Project B")
	time.Sleep(10 * time.Millisecond)
	store.StopTimer()

	pane := NewTimerPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	timerStore, _ := store.LoadTimer()
	pane.setTimerStore(timerStore)

	output := pane.View()

	// Verify the view contains recent entries
	// Note: Golden test would be unstable due to time formatting
	if !contains(output, "Recent") {
		t.Error("view should show recent entries section")
	}
}

func TestTimerPaneView_Unfocused(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	pane := NewTimerPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(false)

	timerStore, _ := store.LoadTimer()
	pane.setTimerStore(timerStore)

	output := pane.View()
	assertGolden(t, "timer_pane_unfocused", output)
}

func TestTimerPaneView_Narrow(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	pane := NewTimerPane(store, styles)
	pane.SetSize(25, 15) // Narrow terminal
	pane.SetFocused(true)

	timerStore, _ := store.LoadTimer()
	pane.setTimerStore(timerStore)

	output := pane.View()
	assertGolden(t, "timer_pane_narrow", output)
}

func TestTimerPane_IsRunning(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	pane := NewTimerPane(store, styles)

	// Initially not running
	if pane.IsRunning() {
		t.Error("IsRunning() = true, want false initially")
	}

	// Start timer
	store.StartTimer("Project")
	timerStore, _ := store.LoadTimer()
	pane.setTimerStore(timerStore)

	if !pane.IsRunning() {
		t.Error("IsRunning() = false, want true after start")
	}

	// Stop timer
	store.StopTimer()
	timerStore, _ = store.LoadTimer()
	pane.setTimerStore(timerStore)

	if pane.IsRunning() {
		t.Error("IsRunning() = true, want false after stop")
	}
}

func TestTimerPane_GetCurrentProject(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	pane := NewTimerPane(store, styles)

	// Start timer with specific project
	projectName := "Test Project Alpha"
	store.StartTimer(projectName)
	timerStore, _ := store.LoadTimer()
	pane.setTimerStore(timerStore)

	got := pane.GetCurrentProject()
	if got != projectName {
		t.Errorf("GetCurrentProject() = %q, want %q", got, projectName)
	}
}

func TestTimerPane_GetElapsed(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	pane := NewTimerPane(store, styles)

	// Start timer
	store.StartTimer("Project")
	time.Sleep(100 * time.Millisecond)

	timerStore, _ := store.LoadTimer()
	pane.setTimerStore(timerStore)

	elapsed := pane.GetElapsed()
	if elapsed < 100*time.Millisecond {
		t.Errorf("GetElapsed() = %v, want >= 100ms", elapsed)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("GetElapsed() = %v, seems too long (test may be slow)", elapsed)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
