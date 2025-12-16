// Package ui provides terminal user interface components for the today app.
// This file defines message types for async I/O operations using the Bubble Tea
// command pattern. All storage operations should return these messages to keep
// the event loop non-blocking.
package ui

import (
	"today/internal/storage"
	"today/internal/sync"
)

// =============================================================================
// Undo/Redo Messages
// =============================================================================

// undoResultMsg is sent when an undo operation completes.
type undoResultMsg struct {
	desc string
	err  error
}

// redoResultMsg is sent when a redo operation completes.
type redoResultMsg struct {
	desc string
	err  error
}

// =============================================================================
// Task Messages
// =============================================================================

// tasksLoadedMsg is sent when tasks are loaded from storage.
type tasksLoadedMsg struct {
	tasks []storage.Task
	err   error
}

// taskAddedMsg is sent when a new task is created.
type taskAddedMsg struct {
	task *storage.Task
	err  error
}

// taskCompletedMsg is sent when a task is marked as done.
type taskCompletedMsg struct {
	id   string
	text string // Task text for undo description
	err  error
}

// taskUncompletedMsg is sent when a task is marked as not done.
type taskUncompletedMsg struct {
	id   string
	text string // Task text for undo description
	err  error
}

// taskDeletedMsg is sent when a task is removed.
type taskDeletedMsg struct {
	id   string
	task *storage.Task // Full task for restoration on undo
	err  error
}

// =============================================================================
// Timer Messages
// =============================================================================

// timerLoadedMsg is sent when timer state is loaded from storage.
type timerLoadedMsg struct {
	store *storage.TimerStore
	err   error
}

// timerStartedMsg is sent when a timer is started for a project.
type timerStartedMsg struct {
	project string
	err     error
}

// timerStoppedMsg is sent when the active timer is stopped.
type timerStoppedMsg struct {
	err error
}

// =============================================================================
// Habit Messages
// =============================================================================

// habitsLoadedMsg is sent when habits are loaded from storage.
type habitsLoadedMsg struct {
	store *storage.HabitStore
	err   error
}

// habitAddedMsg is sent when a new habit is created.
type habitAddedMsg struct {
	habit *storage.Habit
	err   error
}

// habitToggledMsg is sent when a habit's completion status is toggled for today.
type habitToggledMsg struct {
	id           string
	name         string // Habit name for undo description
	date         string // YYYY-MM-DD date toggled (for correct undo after midnight)
	isDone       bool
	wasCompleted bool // Previous state for undo
	err          error
}

// habitDeletedMsg is sent when a habit is removed.
type habitDeletedMsg struct {
	id    string
	habit *storage.Habit     // Full habit for restoration on undo
	logs  []storage.HabitLog // Associated logs for restoration
	err   error
}

// =============================================================================
// Sync Messages
// =============================================================================

// syncStatusMsg is sent when git sync status is refreshed.
type syncStatusMsg struct {
	status *sync.Status
	err    error
}
