package storage

import (
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzAddTask tests AddTask with random text inputs to ensure no panics
// and proper validation handling.
func FuzzAddTask(f *testing.F) {
	// Seed corpus with interesting cases
	f.Add("", "")
	f.Add("Valid task", "")
	f.Add("Task with project", "work")
	f.Add(strings.Repeat("a", maxTaskTextLen), "")
	f.Add(strings.Repeat("a", maxTaskTextLen+1), "")
	f.Add("Task\nwith\nnewlines", "project")
	f.Add("Task with unicode: 🎉🚀✨", "emoji-project")
	f.Add("   whitespace   ", "  spaces  ")
	f.Add("\x00\x01\x02", "") // null bytes
	f.Add("<script>alert('xss')</script>", "security")
	f.Add("Task with 'quotes' and \"double quotes\"", "")

	f.Fuzz(func(t *testing.T, text string, project string) {
		// Create a fresh storage for each test case
		store := createTestStorage(t)

		// AddTask should never panic, even with invalid input
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("AddTask panicked with text=%q project=%q: %v", text, project, r)
			}
		}()

		task, err := store.AddTask(text, project, PriorityNone, nil)

		// If text is empty (after trimming), should return error
		if strings.TrimSpace(text) == "" {
			if err == nil {
				t.Error("AddTask should return error for empty text")
			}
			return
		}

		// If text is too long, should return error
		if len(text) > maxTaskTextLen {
			if err == nil {
				t.Error("AddTask should return error for overly long text")
			}
			return
		}

		// If project is too long, should return error
		if len(project) > maxProjectLen {
			if err == nil {
				t.Error("AddTask should return error for overly long project")
			}
			return
		}

		// Valid input should succeed
		if err != nil {
			t.Errorf("AddTask failed for valid input: %v", err)
			return
		}

		// Verify task properties
		if task == nil {
			t.Error("task should not be nil")
			return
		}

		if task.ID == "" {
			t.Error("task.ID should not be empty")
		}

		if task.Done {
			t.Error("new task should not be marked as done")
		}

		if task.CreatedAt.IsZero() {
			t.Error("task.CreatedAt should be set")
		}

		// Verify text was trimmed
		expectedText := strings.TrimSpace(text)
		if task.Text != expectedText {
			t.Errorf("task.Text = %q, want %q (trimmed)", task.Text, expectedText)
		}

		// Verify project was trimmed
		expectedProject := strings.TrimSpace(project)
		if task.Project != expectedProject {
			t.Errorf("task.Project = %q, want %q (trimmed)", task.Project, expectedProject)
		}

		// Verify task can be loaded back
		loaded, err := store.LoadTasks()
		if err != nil {
			t.Errorf("LoadTasks failed: %v", err)
			return
		}

		if len(loaded.Tasks) != 1 {
			t.Errorf("expected 1 task after add, got %d", len(loaded.Tasks))
			return
		}

		if loaded.Tasks[0].ID != task.ID {
			t.Errorf("loaded task ID mismatch: got %q, want %q", loaded.Tasks[0].ID, task.ID)
		}
	})
}

// FuzzAddHabit tests AddHabit with random name and icon inputs
func FuzzAddHabit(f *testing.F) {
	// Seed corpus
	f.Add("Exercise", "🏃")
	f.Add("", "")
	f.Add(strings.Repeat("a", maxHabitNameLen), "🎯")
	f.Add(strings.Repeat("a", maxHabitNameLen+1), "🎯")
	f.Add("Reading", strings.Repeat("📚", 10))
	f.Add("Meditation", "")
	f.Add("   spaces   ", "  🧘  ")
	f.Add("Habit\nwith\nnewlines", "🔥")
	f.Add("\x00\x01", "\x00")

	f.Fuzz(func(t *testing.T, name string, icon string) {
		store := createTestStorage(t)

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("AddHabit panicked with name=%q icon=%q: %v", name, icon, r)
			}
		}()

		habit, err := store.AddHabit(name, icon)

		// Empty name (after trimming) should error
		if strings.TrimSpace(name) == "" {
			if err == nil {
				t.Error("AddHabit should return error for empty name")
			}
			return
		}

		// Name too long should error
		if len(name) > maxHabitNameLen {
			if err == nil {
				t.Error("AddHabit should return error for overly long name")
			}
			return
		}

		// Empty icon (after trimming) should error
		if strings.TrimSpace(icon) == "" {
			if err == nil {
				t.Error("AddHabit should return error for empty icon")
			}
			return
		}

		// Icon too long should error
		if len(icon) > maxHabitIconLen {
			if err == nil {
				t.Error("AddHabit should return error for overly long icon")
			}
			return
		}

		// Valid input should succeed
		if err != nil {
			t.Errorf("AddHabit failed for valid input: %v", err)
			return
		}

		if habit == nil {
			t.Error("habit should not be nil")
			return
		}

		if habit.ID == "" {
			t.Error("habit.ID should not be empty")
		}

		if habit.CreatedAt.IsZero() {
			t.Error("habit.CreatedAt should be set")
		}

		// Verify trimming
		if habit.Name != strings.TrimSpace(name) {
			t.Errorf("habit.Name not trimmed: got %q", habit.Name)
		}

		if habit.Icon != strings.TrimSpace(icon) {
			t.Errorf("habit.Icon not trimmed: got %q", habit.Icon)
		}

		// Verify persistence
		loaded, err := store.LoadHabits()
		if err != nil {
			t.Errorf("LoadHabits failed: %v", err)
			return
		}

		if len(loaded.Habits) != 1 {
			t.Errorf("expected 1 habit, got %d", len(loaded.Habits))
		}
	})
}

// FuzzStartTimer tests StartTimer with random project names
func FuzzStartTimer(f *testing.F) {
	// Seed corpus
	f.Add("")
	f.Add("Project Alpha")
	f.Add(strings.Repeat("p", maxTimerProjLen))
	f.Add(strings.Repeat("p", maxTimerProjLen+1))
	f.Add("Project\nwith\nnewlines")
	f.Add("   spaces   ")
	f.Add("Unicode: 日本語")
	f.Add("\x00\x01\x02")

	f.Fuzz(func(t *testing.T, project string) {
		store := createTestStorage(t)

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("StartTimer panicked with project=%q: %v", project, r)
			}
		}()

		err := store.StartTimer(project)

		// Empty project (after trimming) should error
		if strings.TrimSpace(project) == "" {
			if err == nil {
				t.Error("StartTimer should return error for empty project")
			}
			return
		}

		// Project too long should error
		if len(project) > maxTimerProjLen {
			if err == nil {
				t.Error("StartTimer should return error for overly long project")
			}
			return
		}

		// Valid input should succeed
		if err != nil {
			t.Errorf("StartTimer failed for valid input: %v", err)
			return
		}

		// Verify timer is running
		timer, err := store.LoadTimer()
		if err != nil {
			t.Errorf("LoadTimer failed: %v", err)
			return
		}

		if timer.Current == nil {
			t.Error("timer.Current should not be nil after start")
			return
		}

		// Verify project was trimmed
		if timer.Current.Project != strings.TrimSpace(project) {
			t.Errorf("Current.Project not trimmed: got %q", timer.Current.Project)
		}

		if timer.Current.StartedAt.IsZero() {
			t.Error("Current.StartedAt should be set")
		}
	})
}

// FuzzTaskStoreJSON tests JSON parsing robustness
func FuzzTaskStoreJSON(f *testing.F) {
	// Seed with valid JSON and edge cases
	f.Add(`{"tasks":[]}`)
	f.Add(`{"tasks":[{"id":"t1","text":"Test","done":false,"created_at":"2025-01-01T00:00:00Z"}]}`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(`{`)
	f.Add(`}`)
	f.Add(`{"tasks":null}`)
	f.Add(`{"tasks":[null]}`)
	f.Add(`{"tasks":[{"id":null}]}`)
	f.Add(`{"extra":"field","tasks":[]}`)

	f.Fuzz(func(t *testing.T, jsonData string) {
		store := createTestStorage(t)

		// Write the fuzzed JSON directly to the file
		path := store.path("tasks.json")

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("LoadTasks panicked with JSON: %q, panic: %v", jsonData, r)
			}
		}()

		// Write potentially malformed JSON
		if err := writeFile(path, []byte(jsonData)); err != nil {
			t.Skip("cannot write file")
		}

		// LoadTasks should handle gracefully (error or recovery, but no panic)
		_, err := store.LoadTasks()

		// We don't check for specific errors because the function
		// has recovery mechanisms. The important thing is it doesn't panic.
		_ = err
	})
}

// FuzzHabitStoreJSON tests JSON parsing robustness for habits
func FuzzHabitStoreJSON(f *testing.F) {
	f.Add(`{"habits":[],"logs":[]}`)
	f.Add(`{"habits":[{"id":"h1","name":"Exercise","icon":"🏃","created_at":"2025-01-01T00:00:00Z"}],"logs":[]}`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(`{"habits":null,"logs":null}`)

	f.Fuzz(func(t *testing.T, jsonData string) {
		store := createTestStorage(t)
		path := store.path("habits.json")

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("LoadHabits panicked with JSON: %q, panic: %v", jsonData, r)
			}
		}()

		if err := writeFile(path, []byte(jsonData)); err != nil {
			t.Skip("cannot write file")
		}

		_, err := store.LoadHabits()
		_ = err // Recovery is acceptable
	})
}

// FuzzTimerStoreJSON tests JSON parsing robustness for timer
func FuzzTimerStoreJSON(f *testing.F) {
	f.Add(`{"entries":[]}`)
	f.Add(`{"current":{"project":"Test","started_at":"2025-01-01T00:00:00Z"},"entries":[]}`)
	f.Add(`{}`)
	f.Add(``)
	f.Add(`{"current":null,"entries":null}`)

	f.Fuzz(func(t *testing.T, jsonData string) {
		store := createTestStorage(t)
		path := store.path("timer.json")

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("LoadTimer panicked with JSON: %q, panic: %v", jsonData, r)
			}
		}()

		if err := writeFile(path, []byte(jsonData)); err != nil {
			t.Skip("cannot write file")
		}

		_, err := store.LoadTimer()
		_ = err // Recovery is acceptable
	})
}

// FuzzUnicodeHandling tests that all functions properly handle Unicode
func FuzzUnicodeHandling(f *testing.F) {
	// Seed with various Unicode edge cases
	f.Add("Emoji: 🎉🚀✨")
	f.Add("Japanese: 日本語")
	f.Add("Arabic: مرحبا")
	f.Add("Chinese: 你好")
	f.Add("Mixed: Hello世界🌍")
	f.Add("Zero-width: A\u200BZ")
	f.Add("RTL: \u202Etext")
	f.Add("Combining: e\u0301") // é as e + combining acute

	f.Fuzz(func(t *testing.T, text string) {
		// Ensure the text is valid UTF-8
		if !utf8.ValidString(text) {
			return
		}

		store := createTestStorage(t)

		// Try adding as task
		if len(text) <= maxTaskTextLen && strings.TrimSpace(text) != "" {
			task, err := store.AddTask(text, "", PriorityNone, nil)
			if err != nil {
				t.Errorf("AddTask failed for valid Unicode: %v", err)
				return
			}

			// Verify round-trip
			loaded, err := store.LoadTasks()
			if err != nil {
				t.Errorf("LoadTasks failed: %v", err)
				return
			}

			if len(loaded.Tasks) > 0 && loaded.Tasks[0].Text != strings.TrimSpace(text) {
				t.Errorf("text corrupted after round-trip: got %q, want %q",
					loaded.Tasks[0].Text, strings.TrimSpace(text))
			}

			// Clean up
			store.DeleteTask(task.ID)
		}

		// Try as timer project
		if len(text) <= maxTimerProjLen && strings.TrimSpace(text) != "" {
			err := store.StartTimer(text)
			if err != nil {
				t.Errorf("StartTimer failed for valid Unicode: %v", err)
				return
			}

			timer, err := store.LoadTimer()
			if err != nil {
				t.Errorf("LoadTimer failed: %v", err)
				return
			}

			if timer.Current != nil && timer.Current.Project != strings.TrimSpace(text) {
				t.Errorf("project corrupted after round-trip: got %q, want %q",
					timer.Current.Project, strings.TrimSpace(text))
			}

			store.StopTimer()
		}
	})
}

// Helper function for writing files in fuzz tests
func writeFile(path string, data []byte) error {
	return writeFileAtomicForTest(path, data, dataFilePerm)
}

func writeFileAtomicForTest(path string, data []byte, perm os.FileMode) error {
	// Simple write for testing purposes
	return os.WriteFile(path, data, perm)
}
