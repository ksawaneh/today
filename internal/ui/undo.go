// Package ui provides terminal user interface components for the today app.
// This file implements the undo/redo system using a command pattern with
// captured state snapshots for each undoable operation.
package ui

import (
	"sync"

	"today/internal/storage"

	"github.com/mattn/go-runewidth"
)

// maxHistorySize limits the undo stack to prevent unbounded memory growth.
const maxHistorySize = 50

// UndoableAction represents an action that can be undone.
// It captures the state needed to reverse the operation.
type UndoableAction struct {
	Description string       // Human-readable description for status messages
	Undo        func() error // Function to reverse the action
	Redo        func() error // Function to redo the action (optional)
}

// UndoManager maintains the undo/redo history stacks.
type UndoManager struct {
	mu        sync.Mutex
	undoStack []*UndoableAction
	redoStack []*UndoableAction
}

// NewUndoManager creates a new UndoManager instance.
func NewUndoManager() *UndoManager {
	return &UndoManager{
		undoStack: make([]*UndoableAction, 0, maxHistorySize),
		redoStack: make([]*UndoableAction, 0, maxHistorySize),
	}
}

// Push adds an undoable action to the history.
// Clears the redo stack since a new action invalidates redo history.
func (m *UndoManager) Push(action *UndoableAction) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear redo stack on new action
	m.redoStack = m.redoStack[:0]

	// Enforce max size (remove oldest if full)
	if len(m.undoStack) >= maxHistorySize {
		m.undoStack = m.undoStack[1:]
	}

	m.undoStack = append(m.undoStack, action)
}

// CanUndo returns true if there are actions to undo.
func (m *UndoManager) CanUndo() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.undoStack) > 0
}

// CanRedo returns true if there are actions to redo.
func (m *UndoManager) CanRedo() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.redoStack) > 0
}

// Undo reverses the most recent action and returns its description.
// Returns empty string and nil error if nothing to undo.
func (m *UndoManager) Undo() (string, error) {
	m.mu.Lock()
	if len(m.undoStack) == 0 {
		m.mu.Unlock()
		return "", nil
	}
	action := m.undoStack[len(m.undoStack)-1]
	m.undoStack = m.undoStack[:len(m.undoStack)-1]
	m.mu.Unlock()

	// Execute undo
	if err := action.Undo(); err != nil {
		// Push back on failure (action not undone)
		m.mu.Lock()
		m.undoStack = append(m.undoStack, action)
		m.mu.Unlock()
		return "", err
	}

	// Push to redo stack if redo is available
	if action.Redo != nil {
		m.mu.Lock()
		m.redoStack = append(m.redoStack, action)
		m.mu.Unlock()
	}

	return action.Description, nil
}

// Redo reapplies the most recently undone action and returns its description.
// Returns empty string and nil error if nothing to redo.
func (m *UndoManager) Redo() (string, error) {
	m.mu.Lock()
	if len(m.redoStack) == 0 {
		m.mu.Unlock()
		return "", nil
	}
	action := m.redoStack[len(m.redoStack)-1]
	m.redoStack = m.redoStack[:len(m.redoStack)-1]
	m.mu.Unlock()

	// Execute redo
	if err := action.Redo(); err != nil {
		// Push back on failure
		m.mu.Lock()
		m.redoStack = append(m.redoStack, action)
		m.mu.Unlock()
		return "", err
	}

	// Push back to undo stack
	m.mu.Lock()
	m.undoStack = append(m.undoStack, action)
	m.mu.Unlock()

	return action.Description, nil
}

// Clear removes all undo/redo history.
func (m *UndoManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.undoStack = m.undoStack[:0]
	m.redoStack = m.redoStack[:0]
}

// =============================================================================
// Undoable Action Factories
// =============================================================================

// NewDeleteTaskAction creates an undoable action for task deletion.
// The task is captured before deletion so it can be restored.
func NewDeleteTaskAction(store *storage.Storage, task storage.Task) *UndoableAction {
	return &UndoableAction{
		Description: "Deleted task: " + truncateText(task.Text, 20),
		Undo: func() error {
			return store.RestoreTask(task)
		},
		Redo: func() error {
			return store.DeleteTask(task.ID)
		},
	}
}

// NewCompleteTaskAction creates an undoable action for task completion.
func NewCompleteTaskAction(store *storage.Storage, taskID string, taskText string) *UndoableAction {
	return &UndoableAction{
		Description: "Completed: " + truncateText(taskText, 20),
		Undo: func() error {
			return store.UncompleteTask(taskID)
		},
		Redo: func() error {
			return store.CompleteTask(taskID)
		},
	}
}

// NewUncompleteTaskAction creates an undoable action for task uncompletion.
func NewUncompleteTaskAction(store *storage.Storage, taskID string, taskText string) *UndoableAction {
	return &UndoableAction{
		Description: "Uncompleted: " + truncateText(taskText, 20),
		Undo: func() error {
			return store.CompleteTask(taskID)
		},
		Redo: func() error {
			return store.UncompleteTask(taskID)
		},
	}
}

// NewDeleteHabitAction creates an undoable action for habit deletion.
// Captures the habit and all its logs for full restoration.
func NewDeleteHabitAction(store *storage.Storage, habit storage.Habit, logs []storage.HabitLog) *UndoableAction {
	return &UndoableAction{
		Description: "Deleted habit: " + truncateText(habit.Name, 20),
		Undo: func() error {
			return store.RestoreHabit(habit, logs)
		},
		Redo: func() error {
			return store.DeleteHabit(habit.ID)
		},
	}
}

// NewToggleHabitAction creates an undoable action for habit toggle.
func NewToggleHabitAction(store *storage.Storage, habitID string, habitName string, date string, wasCompleted bool) *UndoableAction {
	desc := "Completed: " + truncateText(habitName, 20)
	if wasCompleted {
		desc = "Uncompleted: " + truncateText(habitName, 20)
	}
	return &UndoableAction{
		Description: desc,
		Undo: func() error {
			return store.SetHabitDoneOnDate(habitID, date, wasCompleted)
		},
		Redo: func() error {
			return store.SetHabitDoneOnDate(habitID, date, !wasCompleted)
		},
	}
}

// truncateText shortens text to maxLen with ellipsis if needed.
func truncateText(text string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	return runewidth.Truncate(text, maxLen, "..")
}
