// Package ui provides terminal user interface components for the today app.
// This file contains tests for the undo/redo system.
package ui

import (
	"errors"
	"testing"
	"time"

	"today/internal/storage"
)

// TestUndoManager_PushAndUndo verifies basic push and undo operations.
func TestUndoManager_PushAndUndo(t *testing.T) {
	manager := NewUndoManager()

	undoCalled := false
	action := &UndoableAction{
		Description: "Test action",
		Undo: func() error {
			undoCalled = true
			return nil
		},
		Redo: func() error {
			return nil
		},
	}

	// Push action
	manager.Push(action)

	if !manager.CanUndo() {
		t.Error("Expected CanUndo() to return true after push")
	}

	// Undo
	desc, err := manager.Undo()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if desc != "Test action" {
		t.Errorf("Expected description 'Test action', got %q", desc)
	}
	if !undoCalled {
		t.Error("Expected undo function to be called")
	}
	if manager.CanUndo() {
		t.Error("Expected CanUndo() to return false after undoing only action")
	}
}

// TestUndoManager_Redo verifies redo functionality after undo.
func TestUndoManager_Redo(t *testing.T) {
	manager := NewUndoManager()

	state := 0
	action := &UndoableAction{
		Description: "Increment",
		Undo: func() error {
			state--
			return nil
		},
		Redo: func() error {
			state++
			return nil
		},
	}

	// Simulate action that sets state to 1
	state = 1
	manager.Push(action)

	// Undo should decrement state to 0
	_, err := manager.Undo()
	if err != nil {
		t.Errorf("Unexpected error on undo: %v", err)
	}
	if state != 0 {
		t.Errorf("Expected state=0 after undo, got %d", state)
	}

	if !manager.CanRedo() {
		t.Error("Expected CanRedo() to return true after undo")
	}

	// Redo should increment state back to 1
	desc, err := manager.Redo()
	if err != nil {
		t.Errorf("Unexpected error on redo: %v", err)
	}
	if state != 1 {
		t.Errorf("Expected state=1 after redo, got %d", state)
	}
	if desc != "Increment" {
		t.Errorf("Expected description 'Increment', got %q", desc)
	}

	// Can undo again after redo
	if !manager.CanUndo() {
		t.Error("Expected CanUndo() to return true after redo")
	}
	if manager.CanRedo() {
		t.Error("Expected CanRedo() to return false after redo")
	}
}

// TestUndoManager_MaxHistory verifies that oldest entries are removed when full.
func TestUndoManager_MaxHistory(t *testing.T) {
	manager := NewUndoManager()

	// Push more than maxHistorySize actions
	for i := 0; i < maxHistorySize+10; i++ {
		manager.Push(&UndoableAction{
			Description: "Action",
			Undo:        func() error { return nil },
		})
	}

	// Count how many can be undone
	count := 0
	for manager.CanUndo() {
		_, _ = manager.Undo()
		count++
	}

	if count != maxHistorySize {
		t.Errorf("Expected %d undos (max history), got %d", maxHistorySize, count)
	}
}

// TestUndoManager_ClearsRedoOnNewAction verifies that redo stack is cleared on new action.
func TestUndoManager_ClearsRedoOnNewAction(t *testing.T) {
	manager := NewUndoManager()

	// Push action 1
	manager.Push(&UndoableAction{
		Description: "Action 1",
		Undo:        func() error { return nil },
		Redo:        func() error { return nil },
	})

	// Undo it (moves to redo stack)
	_, _ = manager.Undo()

	if !manager.CanRedo() {
		t.Error("Expected CanRedo() to return true after undo")
	}

	// Push a new action (should clear redo stack)
	manager.Push(&UndoableAction{
		Description: "Action 2",
		Undo:        func() error { return nil },
	})

	if manager.CanRedo() {
		t.Error("Expected CanRedo() to return false after pushing new action")
	}
}

// TestUndoManager_UndoError verifies behavior when undo fails.
func TestUndoManager_UndoError(t *testing.T) {
	manager := NewUndoManager()

	expectedErr := errors.New("undo failed")
	action := &UndoableAction{
		Description: "Failing action",
		Undo: func() error {
			return expectedErr
		},
	}

	manager.Push(action)

	// Undo should return error
	desc, err := manager.Undo()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
	if desc != "" {
		t.Errorf("Expected empty description on error, got %q", desc)
	}

	// Action should still be on undo stack (not lost)
	if !manager.CanUndo() {
		t.Error("Expected action to remain on undo stack after failed undo")
	}
}

// TestUndoManager_EmptyUndo verifies behavior when nothing to undo.
func TestUndoManager_EmptyUndo(t *testing.T) {
	manager := NewUndoManager()

	desc, err := manager.Undo()
	if err != nil {
		t.Errorf("Expected no error on empty undo, got %v", err)
	}
	if desc != "" {
		t.Errorf("Expected empty description on empty undo, got %q", desc)
	}
}

// TestUndoManager_EmptyRedo verifies behavior when nothing to redo.
func TestUndoManager_EmptyRedo(t *testing.T) {
	manager := NewUndoManager()

	desc, err := manager.Redo()
	if err != nil {
		t.Errorf("Expected no error on empty redo, got %v", err)
	}
	if desc != "" {
		t.Errorf("Expected empty description on empty redo, got %q", desc)
	}
}

// TestUndoManager_Clear verifies that Clear removes all history.
func TestUndoManager_Clear(t *testing.T) {
	manager := NewUndoManager()

	// Push some actions
	manager.Push(&UndoableAction{
		Description: "Action 1",
		Undo:        func() error { return nil },
		Redo:        func() error { return nil },
	})
	manager.Push(&UndoableAction{
		Description: "Action 2",
		Undo:        func() error { return nil },
		Redo:        func() error { return nil },
	})

	// Undo one to populate redo stack
	_, _ = manager.Undo()

	// Clear
	manager.Clear()

	if manager.CanUndo() {
		t.Error("Expected CanUndo() to return false after Clear")
	}
	if manager.CanRedo() {
		t.Error("Expected CanRedo() to return false after Clear")
	}
}

// TestTruncateText verifies text truncation helper.
func TestTruncateText(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a longer text", 10, "this is .."},
		{"ab", 5, "ab"},
		{"abc", 3, "abc"},
		{"abcd", 3, "a.."},
	}

	for _, tc := range tests {
		result := truncateText(tc.input, tc.maxLen)
		if result != tc.expected {
			t.Errorf("truncateText(%q, %d) = %q, want %q", tc.input, tc.maxLen, result, tc.expected)
		}
	}
}

// TestNewDeleteTaskAction verifies delete task undo action creation.
func TestNewDeleteTaskAction(t *testing.T) {
	store := createTestStorage(t)

	// Add a task first
	task, err := store.AddTask("Buy milk", "groceries", storage.PriorityNone, nil)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Create undo action with the task data
	action := NewDeleteTaskAction(store, *task)

	if action.Description != "Deleted task: Buy milk" {
		t.Errorf("Unexpected description: %s", action.Description)
	}

	// Now delete the task
	if err := store.DeleteTask(task.ID); err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Undo should restore the task
	if err := action.Undo(); err != nil {
		t.Errorf("Undo failed: %v", err)
	}

	// Verify task exists again
	tasks, err := store.LoadTasks()
	if err != nil {
		t.Fatalf("Failed to load tasks: %v", err)
	}

	found := false
	for _, t := range tasks.Tasks {
		if t.Text == "Buy milk" && t.Project == "groceries" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Task was not restored after undo")
	}
}

// TestNewCompleteTaskAction verifies complete task undo action.
func TestNewCompleteTaskAction(t *testing.T) {
	store := createTestStorage(t)

	// Add a task
	task, err := store.AddTask("Write tests", "", storage.PriorityNone, nil)
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	// Create undo action before completing
	action := NewCompleteTaskAction(store, task.ID, task.Text)

	// Complete the task
	if err := store.CompleteTask(task.ID); err != nil {
		t.Fatalf("Failed to complete task: %v", err)
	}

	// Verify it's done
	tasks, _ := store.LoadTasks()
	for _, tsk := range tasks.Tasks {
		if tsk.ID == task.ID && !tsk.Done {
			t.Error("Task should be completed")
		}
	}

	// Undo should uncomplete the task
	if err := action.Undo(); err != nil {
		t.Errorf("Undo failed: %v", err)
	}

	// Verify task is not done
	tasks, _ = store.LoadTasks()
	for _, tsk := range tasks.Tasks {
		if tsk.ID == task.ID && tsk.Done {
			t.Error("Task should be uncompleted after undo")
		}
	}

	// Redo should complete again
	if err := action.Redo(); err != nil {
		t.Errorf("Redo failed: %v", err)
	}

	tasks, _ = store.LoadTasks()
	for _, tsk := range tasks.Tasks {
		if tsk.ID == task.ID && !tsk.Done {
			t.Error("Task should be completed after redo")
		}
	}
}

// TestNewToggleHabitAction verifies toggle habit undo action.
func TestNewToggleHabitAction(t *testing.T) {
	store := createTestStorage(t)

	// Add a habit
	habit, err := store.AddHabit("Exercise", "🏃")
	if err != nil {
		t.Fatalf("Failed to add habit: %v", err)
	}

	date := time.Now().Format("2006-01-02")

	// Create undo action (habit was not completed)
	action := NewToggleHabitAction(store, habit.ID, habit.Name, date, false)

	// Toggle should mark as done
	isDone, err := store.ToggleHabitToday(habit.ID)
	if err != nil {
		t.Fatalf("Failed to toggle habit: %v", err)
	}
	if !isDone {
		t.Error("Expected habit to be done after toggle")
	}

	// Undo should restore previous state (not done)
	if err := action.Undo(); err != nil {
		t.Errorf("Undo failed: %v", err)
	}

	habits, err := store.LoadHabits()
	if err != nil {
		t.Fatalf("Failed to load habits: %v", err)
	}
	if store.IsHabitDoneOnDate(habits, habit.ID, date) {
		t.Error("Expected habit to be not done after undo")
	}
}
