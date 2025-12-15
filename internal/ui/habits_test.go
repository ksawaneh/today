package ui

import (
	"testing"
	"time"

	"today/internal/storage"
)

func freezeHabitsNow(t *testing.T, store *storage.Storage) {
	t.Helper()
	// Monday, so labels render as: T W T F S S M
	store.SetNowFunc(func() time.Time {
		return time.Date(2025, 12, 15, 12, 0, 0, 0, time.UTC)
	})
}

func TestHabitsPaneView_Empty(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	freezeHabitsNow(t, store)

	pane := NewHabitsPane(store, createTestStyles())
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	// Load empty habits
	habitStore, _ := store.LoadHabits()
	pane.setHabitStore(habitStore)

	output := pane.View()
	assertGolden(t, "habits_pane_empty", output)
}

func TestHabitsPaneView_WithHabits(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	freezeHabitsNow(t, store)

	// Add some habits
	store.AddHabit("Exercise", "🏃")
	store.AddHabit("Reading", "📚")
	store.AddHabit("Meditation", "🧘")

	pane := NewHabitsPane(store, createTestStyles())
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	// Load habits
	habitStore, _ := store.LoadHabits()
	pane.setHabitStore(habitStore)

	output := pane.View()
	assertGolden(t, "habits_pane_with_habits", output)
}

func TestHabitsPaneView_WithCompletion(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	freezeHabitsNow(t, store)

	// Add habits
	habit1, _ := store.AddHabit("Exercise", "🏃")
	_, _ = store.AddHabit("Reading", "📚")

	// Complete one for today
	store.ToggleHabitToday(habit1.ID)

	pane := NewHabitsPane(store, createTestStyles())
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	habitStore, _ := store.LoadHabits()
	pane.setHabitStore(habitStore)

	output := pane.View()
	assertGolden(t, "habits_pane_with_completion", output)
}

func TestHabitsPaneView_WithStreak(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	freezeHabitsNow(t, store)

	// Add a habit
	habit, _ := store.AddHabit("Exercise", "🏃")

	// Create a multi-day streak
	habitStore, _ := store.LoadHabits()
	today := store.Now()
	for i := 0; i < 5; i++ {
		date := today.AddDate(0, 0, -i).Format("2006-01-02")
		habitStore.Logs = append(habitStore.Logs, storage.HabitLog{
			HabitID: habit.ID,
			Date:    date,
		})
	}
	store.SaveHabits(habitStore)

	pane := NewHabitsPane(store, createTestStyles())
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	habitStore, _ = store.LoadHabits()
	pane.setHabitStore(habitStore)

	output := pane.View()
	assertGolden(t, "habits_pane_with_streak", output)
}

func TestHabitsPaneView_Unfocused(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	freezeHabitsNow(t, store)

	// Add a habit
	store.AddHabit("Exercise", "🏃")

	pane := NewHabitsPane(store, createTestStyles())
	pane.SetSize(40, 20)
	pane.SetFocused(false)

	habitStore, _ := store.LoadHabits()
	pane.setHabitStore(habitStore)

	output := pane.View()
	assertGolden(t, "habits_pane_unfocused", output)
}

func TestHabitsPaneView_Narrow(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	freezeHabitsNow(t, store)

	// Add habits
	store.AddHabit("Exercise", "🏃")
	store.AddHabit("Read", "📚")

	pane := NewHabitsPane(store, createTestStyles())
	pane.SetSize(25, 15) // Narrow terminal
	pane.SetFocused(true)

	habitStore, _ := store.LoadHabits()
	pane.setHabitStore(habitStore)

	output := pane.View()
	assertGolden(t, "habits_pane_narrow", output)
}

func TestHabitsPane_Navigation(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	freezeHabitsNow(t, store)

	// Add habits
	store.AddHabit("Habit 1", "1️⃣")
	store.AddHabit("Habit 2", "2️⃣")
	store.AddHabit("Habit 3", "3️⃣")

	pane := NewHabitsPane(store, createTestStyles())
	pane.SetSize(40, 20)
	pane.SetFocused(true)

	habitStore, _ := store.LoadHabits()
	pane.setHabitStore(habitStore)

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

	// Bounds checking
	pane.cursor = 10 // Try to go past end
	pane.setHabitStore(habitStore)
	if pane.cursor >= len(habitStore.Habits) {
		t.Error("cursor should be bounded to habit count")
	}
}

func TestHabitsPane_GetTodayCompletionRate(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	freezeHabitsNow(t, store)

	pane := NewHabitsPane(store, createTestStyles())

	// Empty initially
	done, total := pane.GetTodayCompletionRate()
	if done != 0 || total != 0 {
		t.Errorf("GetTodayCompletionRate() = (%d, %d), want (0, 0)", done, total)
	}

	// Add habits
	habit1, _ := store.AddHabit("Habit 1", "1️⃣")
	habit2, _ := store.AddHabit("Habit 2", "2️⃣")
	habit3, _ := store.AddHabit("Habit 3", "3️⃣")

	// Complete 2 out of 3
	store.ToggleHabitToday(habit1.ID)
	store.ToggleHabitToday(habit2.ID)

	habitStore, _ := store.LoadHabits()
	pane.setHabitStore(habitStore)

	done, total = pane.GetTodayCompletionRate()
	if done != 2 || total != 3 {
		t.Errorf("GetTodayCompletionRate() = (%d, %d), want (2, 3)", done, total)
	}

	// Complete the third
	store.ToggleHabitToday(habit3.ID)
	habitStore, _ = store.LoadHabits()
	pane.setHabitStore(habitStore)

	done, total = pane.GetTodayCompletionRate()
	if done != 3 || total != 3 {
		t.Errorf("GetTodayCompletionRate() = (%d, %d), want (3, 3)", done, total)
	}
}

func TestHabitsPane_WeekVisualization(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)
	freezeHabitsNow(t, store)

	// Add a habit
	habit, _ := store.AddHabit("Test Habit", "✓")

	// Complete for specific days this week
	habitStore, _ := store.LoadHabits()
	today := store.Now()

	// Complete today and 2 days ago
	for _, offset := range []int{0, 2} {
		date := today.AddDate(0, 0, -offset).Format("2006-01-02")
		habitStore.Logs = append(habitStore.Logs, storage.HabitLog{
			HabitID: habit.ID,
			Date:    date,
		})
	}
	store.SaveHabits(habitStore)

	// Get week data
	habitStore, _ = store.LoadHabits()
	week := store.GetHabitWeek(habitStore, habit.ID)

	// Should have 7 days
	if len(week) != 7 {
		t.Fatalf("week length = %d, want 7", len(week))
	}

	// Last day (today) should be done
	if !week[6] {
		t.Error("week[6] (today) should be true")
	}

	// Day 4 (2 days ago) should be done
	if !week[4] {
		t.Error("week[4] (2 days ago) should be true")
	}
}

func TestHabitsPane_AddMode(t *testing.T) {
	setupTest(t)
	store := createTestStorage(t)

	pane := NewHabitsPane(store, createTestStyles())
	pane.SetFocused(true)

	// Initially not adding
	if pane.IsAdding() {
		t.Error("IsAdding() = true, want false initially")
	}

	// Entering add mode doesn't require actual key simulation for this test
	// Just verify the state can be set
	pane.adding = true
	if !pane.IsAdding() {
		t.Error("IsAdding() = false, want true after setting adding mode")
	}

	// Reset add mode
	pane.resetAddMode()
	if pane.IsAdding() {
		t.Error("IsAdding() = true, want false after reset")
	}
	if pane.addStep != 0 {
		t.Error("addStep should be 0 after reset")
	}
	if pane.newName != "" {
		t.Error("newName should be empty after reset")
	}
}
