package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// createTestStorage creates a Storage instance with a temporary directory.
func createTestStorage(t *testing.T) *Storage {
	t.Helper()
	store, err := New(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test storage: %v", err)
	}
	return store
}

// =============================================================================
// Task Tests
// =============================================================================

func TestAddTask(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		project string
	}{
		{
			name:    "simple task",
			text:    "Buy groceries",
			project: "",
		},
		{
			name:    "task with project",
			text:    "Write tests",
			project: "today-app",
		},
		{
			name:    "task with special characters",
			text:    "Fix bug: 'undefined' error in @main",
			project: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := createTestStorage(t)

			task, err := store.AddTask(tt.text, tt.project, PriorityNone, nil)
			if err != nil {
				t.Fatalf("AddTask() error = %v", err)
			}

			if task.Text != tt.text {
				t.Errorf("task.Text = %q, want %q", task.Text, tt.text)
			}
			if task.Project != tt.project {
				t.Errorf("task.Project = %q, want %q", task.Project, tt.project)
			}
			if task.Done {
				t.Error("task.Done = true, want false")
			}
			if task.ID == "" {
				t.Error("task.ID is empty")
			}
			if task.CreatedAt.IsZero() {
				t.Error("task.CreatedAt is zero")
			}

			// Verify task was persisted
			loaded, err := store.LoadTasks()
			if err != nil {
				t.Fatalf("LoadTasks() error = %v", err)
			}
			if len(loaded.Tasks) != 1 {
				t.Fatalf("len(tasks) = %d, want 1", len(loaded.Tasks))
			}
			if loaded.Tasks[0].ID != task.ID {
				t.Errorf("persisted task ID = %q, want %q", loaded.Tasks[0].ID, task.ID)
			}
		})
	}
}

func TestAddTask_Validation(t *testing.T) {
	store := createTestStorage(t)

	if _, err := store.AddTask("   ", "", PriorityNone, nil); err == nil {
		t.Fatal("AddTask() expected error for empty task text")
	}

	long := make([]byte, maxTaskTextLen+1)
	for i := range long {
		long[i] = 'a'
	}
	if _, err := store.AddTask(string(long), "", PriorityNone, nil); err == nil {
		t.Fatal("AddTask() expected error for overly long task text")
	}
}

func TestCompleteTask(t *testing.T) {
	store := createTestStorage(t)

	// Add a task first
	task, err := store.AddTask("Test task", "", PriorityNone, nil)
	if err != nil {
		t.Fatalf("AddTask() error = %v", err)
	}

	// Complete the task
	if err := store.CompleteTask(task.ID); err != nil {
		t.Fatalf("CompleteTask() error = %v", err)
	}

	// Verify task is complete
	loaded, err := store.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}
	if !loaded.Tasks[0].Done {
		t.Error("task.Done = false, want true")
	}
	if loaded.Tasks[0].CompletedAt == nil {
		t.Error("task.CompletedAt is nil")
	}
}

func TestCompleteTask_NotFound(t *testing.T) {
	store := createTestStorage(t)

	err := store.CompleteTask("nonexistent")
	if err == nil {
		t.Error("CompleteTask() expected error for nonexistent task")
	}
}

func TestUncompleteTask(t *testing.T) {
	store := createTestStorage(t)

	// Add and complete a task
	task, _ := store.AddTask("Test task", "", PriorityNone, nil)
	store.CompleteTask(task.ID)

	// Uncomplete the task
	if err := store.UncompleteTask(task.ID); err != nil {
		t.Fatalf("UncompleteTask() error = %v", err)
	}

	// Verify task is incomplete
	loaded, _ := store.LoadTasks()
	if loaded.Tasks[0].Done {
		t.Error("task.Done = true, want false")
	}
	if loaded.Tasks[0].CompletedAt != nil {
		t.Error("task.CompletedAt should be nil")
	}
}

func TestDeleteTask(t *testing.T) {
	store := createTestStorage(t)

	// Add two tasks
	task1, _ := store.AddTask("Task 1", "", PriorityNone, nil)
	task2, _ := store.AddTask("Task 2", "", PriorityNone, nil)

	// Delete the first task
	if err := store.DeleteTask(task1.ID); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	// Verify only task2 remains
	loaded, _ := store.LoadTasks()
	if len(loaded.Tasks) != 1 {
		t.Fatalf("len(tasks) = %d, want 1", len(loaded.Tasks))
	}
	if loaded.Tasks[0].ID != task2.ID {
		t.Errorf("remaining task ID = %q, want %q", loaded.Tasks[0].ID, task2.ID)
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	store := createTestStorage(t)

	err := store.DeleteTask("nonexistent")
	if err == nil {
		t.Error("DeleteTask() expected error for nonexistent task")
	}
}

// =============================================================================
// Habit Tests
// =============================================================================

func TestAddHabit(t *testing.T) {
	tests := []struct {
		name  string
		hName string
		icon  string
	}{
		{
			name:  "simple habit",
			hName: "Exercise",
			icon:  "🏃",
		},
		{
			name:  "habit with text icon",
			hName: "Read",
			icon:  "📚",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := createTestStorage(t)

			habit, err := store.AddHabit(tt.hName, tt.icon)
			if err != nil {
				t.Fatalf("AddHabit() error = %v", err)
			}

			if habit.Name != tt.hName {
				t.Errorf("habit.Name = %q, want %q", habit.Name, tt.hName)
			}
			if habit.Icon != tt.icon {
				t.Errorf("habit.Icon = %q, want %q", habit.Icon, tt.icon)
			}
			if habit.ID == "" {
				t.Error("habit.ID is empty")
			}

			// Verify persistence
			loaded, _ := store.LoadHabits()
			if len(loaded.Habits) != 1 {
				t.Fatalf("len(habits) = %d, want 1", len(loaded.Habits))
			}
		})
	}
}

func TestAddHabit_Validation(t *testing.T) {
	store := createTestStorage(t)

	if _, err := store.AddHabit("   ", "🏃"); err == nil {
		t.Fatal("AddHabit() expected error for empty habit name")
	}
	if _, err := store.AddHabit("Exercise", ""); err == nil {
		t.Fatal("AddHabit() expected error for empty icon")
	}
}

func TestToggleHabitToday(t *testing.T) {
	store := createTestStorage(t)

	habit, _ := store.AddHabit("Exercise", "🏃")

	// Toggle on
	isDone, err := store.ToggleHabitToday(habit.ID)
	if err != nil {
		t.Fatalf("ToggleHabitToday() error = %v", err)
	}
	if !isDone {
		t.Error("isDone = false, want true after toggle on")
	}

	// Verify log was created
	loaded, _ := store.LoadHabits()
	if len(loaded.Logs) != 1 {
		t.Fatalf("len(logs) = %d, want 1", len(loaded.Logs))
	}

	// Toggle off
	isDone, err = store.ToggleHabitToday(habit.ID)
	if err != nil {
		t.Fatalf("ToggleHabitToday() error = %v", err)
	}
	if isDone {
		t.Error("isDone = true, want false after toggle off")
	}

	// Verify log was removed
	loaded, _ = store.LoadHabits()
	if len(loaded.Logs) != 0 {
		t.Fatalf("len(logs) = %d, want 0", len(loaded.Logs))
	}
}

func TestToggleHabitToday_RemovesDuplicates(t *testing.T) {
	store := createTestStorage(t)

	habit, _ := store.AddHabit("Exercise", "🏃")
	today := time.Now().Format("2006-01-02")

	hs, _ := store.LoadHabits()
	hs.Logs = append(hs.Logs,
		HabitLog{HabitID: habit.ID, Date: today},
		HabitLog{HabitID: habit.ID, Date: today},
	)
	if err := store.SaveHabits(hs); err != nil {
		t.Fatalf("SaveHabits() error = %v", err)
	}

	isDone, err := store.ToggleHabitToday(habit.ID)
	if err != nil {
		t.Fatalf("ToggleHabitToday() error = %v", err)
	}
	if isDone {
		t.Fatal("ToggleHabitToday() expected to toggle off (false)")
	}

	hs, _ = store.LoadHabits()
	for _, log := range hs.Logs {
		if log.HabitID == habit.ID && log.Date == today {
			t.Fatal("expected all duplicate logs for today to be removed")
		}
	}
}

func TestIsHabitDoneOnDate(t *testing.T) {
	store := createTestStorage(t)

	habit, _ := store.AddHabit("Exercise", "🏃")
	today := time.Now().Format("2006-01-02")

	// Not done initially
	hs, _ := store.LoadHabits()
	if store.IsHabitDoneOnDate(hs, habit.ID, today) {
		t.Error("IsHabitDoneOnDate() = true, want false before toggle")
	}

	// Toggle on
	store.ToggleHabitToday(habit.ID)
	hs, _ = store.LoadHabits()
	if !store.IsHabitDoneOnDate(hs, habit.ID, today) {
		t.Error("IsHabitDoneOnDate() = false, want true after toggle")
	}
}

func TestGetHabitStreak(t *testing.T) {
	store := createTestStorage(t)

	habit, _ := store.AddHabit("Exercise", "🏃")
	hs, _ := store.LoadHabits()

	// No streak initially
	streak := store.GetHabitStreak(hs, habit.ID)
	if streak != 0 {
		t.Errorf("GetHabitStreak() = %d, want 0", streak)
	}

	// Add log for today
	store.ToggleHabitToday(habit.ID)
	hs, _ = store.LoadHabits()
	streak = store.GetHabitStreak(hs, habit.ID)
	if streak != 1 {
		t.Errorf("GetHabitStreak() = %d, want 1", streak)
	}
}

func TestGetHabitWeek(t *testing.T) {
	store := createTestStorage(t)

	habit, _ := store.AddHabit("Exercise", "🏃")

	// All false initially
	hs, _ := store.LoadHabits()
	week := store.GetHabitWeek(hs, habit.ID)
	if len(week) != 7 {
		t.Fatalf("len(week) = %d, want 7", len(week))
	}
	for i, done := range week {
		if done {
			t.Errorf("week[%d] = true, want false", i)
		}
	}

	// Toggle today
	store.ToggleHabitToday(habit.ID)
	hs, _ = store.LoadHabits()
	week = store.GetHabitWeek(hs, habit.ID)

	// Last element (today) should be true
	if !week[6] {
		t.Error("week[6] (today) = false, want true")
	}
}

func TestDeleteHabit(t *testing.T) {
	store := createTestStorage(t)

	habit, _ := store.AddHabit("Exercise", "🏃")
	store.ToggleHabitToday(habit.ID) // Add a log

	// Delete
	if err := store.DeleteHabit(habit.ID); err != nil {
		t.Fatalf("DeleteHabit() error = %v", err)
	}

	// Verify habit and logs are removed
	hs, _ := store.LoadHabits()
	if len(hs.Habits) != 0 {
		t.Errorf("len(habits) = %d, want 0", len(hs.Habits))
	}
	if len(hs.Logs) != 0 {
		t.Errorf("len(logs) = %d, want 0", len(hs.Logs))
	}
}

func TestDeleteHabit_NotFound(t *testing.T) {
	store := createTestStorage(t)

	if err := store.DeleteHabit("missing"); err == nil {
		t.Fatal("DeleteHabit() expected error for missing habit")
	}
}

// =============================================================================
// Timer Tests
// =============================================================================

func TestStartTimer(t *testing.T) {
	store := createTestStorage(t)

	if err := store.StartTimer("Project A"); err != nil {
		t.Fatalf("StartTimer() error = %v", err)
	}

	ts, _ := store.LoadTimer()
	if ts.Current == nil {
		t.Fatal("ts.Current is nil")
	}
	if ts.Current.Project != "Project A" {
		t.Errorf("Current.Project = %q, want %q", ts.Current.Project, "Project A")
	}
}

func TestStartTimer_Validation(t *testing.T) {
	store := createTestStorage(t)
	if err := store.StartTimer("   "); err == nil {
		t.Fatal("StartTimer() expected error for empty project")
	}
}

func TestStartTimer_SwitchProject(t *testing.T) {
	store := createTestStorage(t)

	// Start first timer
	store.StartTimer("Project A")
	time.Sleep(10 * time.Millisecond) // Ensure some time passes

	// Switch to second project
	if err := store.StartTimer("Project B"); err != nil {
		t.Fatalf("StartTimer() error = %v", err)
	}

	ts, _ := store.LoadTimer()

	// Current should be Project B
	if ts.Current.Project != "Project B" {
		t.Errorf("Current.Project = %q, want %q", ts.Current.Project, "Project B")
	}

	// Should have one entry for Project A
	if len(ts.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(ts.Entries))
	}
	if ts.Entries[0].Project != "Project A" {
		t.Errorf("Entries[0].Project = %q, want %q", ts.Entries[0].Project, "Project A")
	}
}

func TestStopTimer(t *testing.T) {
	store := createTestStorage(t)

	store.StartTimer("Project A")
	time.Sleep(10 * time.Millisecond)

	if err := store.StopTimer(); err != nil {
		t.Fatalf("StopTimer() error = %v", err)
	}

	ts, _ := store.LoadTimer()

	// Current should be nil
	if ts.Current != nil {
		t.Error("ts.Current should be nil after stop")
	}

	// Should have one entry
	if len(ts.Entries) != 1 {
		t.Fatalf("len(Entries) = %d, want 1", len(ts.Entries))
	}
}

func TestStopTimer_NoCurrentTimer(t *testing.T) {
	store := createTestStorage(t)

	// Should not error when no timer is running
	if err := store.StopTimer(); err != nil {
		t.Fatalf("StopTimer() error = %v", err)
	}
}

func TestGetTodayTotal(t *testing.T) {
	store := createTestStorage(t)

	// No entries initially
	ts, _ := store.LoadTimer()
	total := store.GetTodayTotal(ts)
	if total != 0 {
		t.Errorf("GetTodayTotal() = %v, want 0", total)
	}

	// Add a completed entry
	store.StartTimer("Project A")
	time.Sleep(50 * time.Millisecond)
	store.StopTimer()

	ts, _ = store.LoadTimer()
	total = store.GetTodayTotal(ts)
	if total < 50*time.Millisecond {
		t.Errorf("GetTodayTotal() = %v, want >= 50ms", total)
	}
}

func TestGetWeekTotal(t *testing.T) {
	store := createTestStorage(t)

	// Similar to GetTodayTotal but for the week
	store.StartTimer("Project A")
	time.Sleep(50 * time.Millisecond)
	store.StopTimer()

	ts, _ := store.LoadTimer()
	total := store.GetWeekTotal(ts)
	if total < 50*time.Millisecond {
		t.Errorf("GetWeekTotal() = %v, want >= 50ms", total)
	}
}

func TestTimerTotals_SplitAcrossBoundaries(t *testing.T) {
	store := createTestStorage(t)
	loc := time.Local

	// Sunday, so week starts today (Sunday-based).
	now := time.Date(2025, 12, 14, 10, 0, 0, 0, loc)
	dayStart := time.Date(2025, 12, 14, 0, 0, 0, 0, loc)

	ts := &TimerStore{
		Entries: []TimerEntry{
			// Spans midnight: counts 1h for today/week.
			{Project: "X", StartedAt: dayStart.Add(-1 * time.Hour), EndedAt: dayStart.Add(1 * time.Hour)},
			// Fully today: counts 1h.
			{Project: "Y", StartedAt: dayStart.Add(9 * time.Hour), EndedAt: dayStart.Add(10 * time.Hour)},
		},
		Current: &CurrentTimer{
			Project:   "Z",
			StartedAt: dayStart.Add(9*time.Hour + 30*time.Minute),
		},
	}

	today := store.getTodayTotalAt(ts, now)
	week := store.getWeekTotalAt(ts, now)

	// Today should include: 1h (midnight overlap) + 1h + 30m (current running from 09:30 -> 10:00)
	want := 2*time.Hour + 30*time.Minute
	if today != want {
		t.Fatalf("getTodayTotalAt() = %v, want %v", today, want)
	}
	if week != want {
		t.Fatalf("getWeekTotalAt() = %v, want %v", week, want)
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestStorageInitialization(t *testing.T) {
	store := createTestStorage(t)

	// All stores should be initialized and loadable
	tasks, err := store.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error = %v", err)
	}
	if tasks == nil {
		t.Error("tasks is nil")
	}

	habits, err := store.LoadHabits()
	if err != nil {
		t.Fatalf("LoadHabits() error = %v", err)
	}
	if habits == nil {
		t.Error("habits is nil")
	}

	timer, err := store.LoadTimer()
	if err != nil {
		t.Fatalf("LoadTimer() error = %v", err)
	}
	if timer == nil {
		t.Error("timer is nil")
	}
}

func TestMultipleTasks(t *testing.T) {
	store := createTestStorage(t)

	// Add multiple tasks
	for i := 0; i < 10; i++ {
		_, err := store.AddTask("Task", "", PriorityNone, nil)
		if err != nil {
			t.Fatalf("AddTask() error = %v", err)
		}
	}

	loaded, _ := store.LoadTasks()
	if len(loaded.Tasks) != 10 {
		t.Errorf("len(tasks) = %d, want 10", len(loaded.Tasks))
	}
}

func TestStorage_PermissionsArePrivate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permissions are not meaningful on Windows")
	}

	dataDir := t.TempDir()
	_, err := New(dataDir)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	paths := []string{
		filepath.Join(dataDir, "tasks.json"),
		filepath.Join(dataDir, "habits.json"),
		filepath.Join(dataDir, "timer.json"),
	}
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("Stat(%s) error = %v", p, err)
		}
		if info.Mode().Perm()&0o077 != 0 {
			t.Fatalf("%s permissions = %o, want no group/other bits", p, info.Mode().Perm())
		}
	}
}
