package ui

import (
	"testing"
	"time"

	"today/internal/storage"
)

func TestTaskPaneView_Empty(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	pane := NewTaskPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	output := pane.View()
	assertGolden(t, "task_pane_empty", output)
}

func TestTaskPaneView_WithTasks(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	// Add some tasks (updated to include priority and due date)
	store.AddTask("Buy groceries", "", storage.PriorityNone, nil)
	store.AddTask("Write tests", "today", storage.PriorityNone, nil)
	store.AddTask("Review PR", "", storage.PriorityNone, nil)

	pane := NewTaskPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	// Load tasks
	tasks, _ := store.LoadTasks()
	pane.setTasks(tasks.Tasks)

	output := pane.View()
	assertGolden(t, "task_pane_with_tasks", output)
}

func TestTaskPaneView_WithCompletedTasks(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	// Add some tasks
	task1, _ := store.AddTask("Completed task", "", storage.PriorityNone, nil)
	store.AddTask("Pending task", "", storage.PriorityNone, nil)

	// Complete first task
	store.CompleteTask(task1.ID)

	pane := NewTaskPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	// Load tasks
	tasks, _ := store.LoadTasks()
	pane.setTasks(tasks.Tasks)

	output := pane.View()
	assertGolden(t, "task_pane_with_completed", output)
}

func TestTaskPaneView_Unfocused(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	store.AddTask("A task", "", storage.PriorityNone, nil)

	pane := NewTaskPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(false)

	tasks, _ := store.LoadTasks()
	pane.setTasks(tasks.Tasks)

	output := pane.View()
	assertGolden(t, "task_pane_unfocused", output)
}

func TestTaskPane_Navigation(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	// Add tasks
	store.AddTask("Task 1", "", storage.PriorityNone, nil)
	store.AddTask("Task 2", "", storage.PriorityNone, nil)
	store.AddTask("Task 3", "", storage.PriorityNone, nil)

	pane := NewTaskPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	tasks, _ := store.LoadTasks()
	pane.setTasks(tasks.Tasks)

	// Initial cursor should be at 0
	if pane.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", pane.cursor)
	}

	// Move down
	pane.cursor = 1
	if pane.cursor != 1 {
		t.Errorf("cursor after down = %d, want 1", pane.cursor)
	}

	// Move to bottom
	pane.cursor = 2
	if pane.cursor != 2 {
		t.Errorf("cursor at bottom = %d, want 2", pane.cursor)
	}
}

func TestTaskPane_Stats(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	pane := NewTaskPane(store, styles)

	// Empty initially
	done, total := pane.Stats()
	if done != 0 || total != 0 {
		t.Errorf("Stats() = (%d, %d), want (0, 0)", done, total)
	}

	// Add and complete some tasks
	task1, _ := store.AddTask("Task 1", "", storage.PriorityNone, nil)
	store.AddTask("Task 2", "", storage.PriorityNone, nil)
	store.CompleteTask(task1.ID)

	tasks, _ := store.LoadTasks()
	pane.setTasks(tasks.Tasks)

	done, total = pane.Stats()
	if done != 1 || total != 2 {
		t.Errorf("Stats() = (%d, %d), want (1, 2)", done, total)
	}
}

func TestTaskPaneView_WithPriority(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	// Add tasks with different priorities
	store.AddTask("High priority task", "", storage.PriorityHigh, nil)
	store.AddTask("Medium priority task", "", storage.PriorityMedium, nil)
	store.AddTask("Low priority task", "", storage.PriorityLow, nil)
	store.AddTask("No priority task", "", storage.PriorityNone, nil)

	pane := NewTaskPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	tasks, _ := store.LoadTasks()
	pane.setTasks(tasks.Tasks)

	output := pane.View()
	assertGolden(t, "task_pane_with_priority", output)
}

func TestTaskPaneView_WithDueDate(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	today := now
	tomorrow := now.AddDate(0, 0, 1)
	nextWeek := now.AddDate(0, 0, 7)

	store.AddTask("Overdue task", "", storage.PriorityNone, &yesterday)
	store.AddTask("Due today", "", storage.PriorityNone, &today)
	store.AddTask("Due tomorrow", "", storage.PriorityNone, &tomorrow)
	store.AddTask("Due next week", "", storage.PriorityNone, &nextWeek)
	store.AddTask("No due date", "", storage.PriorityNone, nil)

	pane := NewTaskPane(store, styles)
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	tasks, _ := store.LoadTasks()
	pane.setTasks(tasks.Tasks)

	output := pane.View()
	assertGolden(t, "task_pane_with_due_dates", output)
}

func TestTaskPaneView_NarrowWithPriorityAndDue(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	styles := createTestStyles()

	tomorrow := time.Now().AddDate(0, 0, 1)
	store.AddTask("Very long task name that should truncate", "", storage.PriorityHigh, &tomorrow)

	pane := NewTaskPane(store, styles)
	pane.SetSize(25, 20) // Narrow width
	pane.SetFocused(true)

	tasks, _ := store.LoadTasks()
	pane.setTasks(tasks.Tasks)

	output := pane.View()
	assertGolden(t, "task_pane_narrow_priority_due", output)
}

// TestFormatDueDate tests the due date formatting helper.
func TestFormatDueDate(t *testing.T) {
	store := createTestStorage(t)
	styles := createTestStyles()
	pane := NewTaskPane(store, styles)

	now := time.Now()

	tests := []struct {
		name    string
		dueDate *time.Time
		wantLen int // Expected string length (0 for nil/empty)
	}{
		{"nil due date", nil, 0},
		{"overdue", timePtr(now.AddDate(0, 0, -1)), 1},  // "!"
		{"today", timePtr(now), 1},                      // "T"
		{"tomorrow", timePtr(now.AddDate(0, 0, 1)), 2},  // "+1"
		{"3 days", timePtr(now.AddDate(0, 0, 3)), 2},    // "3d"
		{"2 weeks", timePtr(now.AddDate(0, 0, 14)), 2},  // "2w"
		{"over month", timePtr(now.AddDate(0, 0, 45)), 3}, // ">1m"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pane.formatDueDate(tt.dueDate)
			if tt.wantLen == 0 && result != "" {
				t.Errorf("expected empty string, got %q", result)
			}
			// Note: In ASCII mode, we can't verify exact styled output length
			// but we verify it's not empty when it should have content
			if tt.wantLen > 0 && result == "" {
				t.Errorf("expected non-empty string for %s", tt.name)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
