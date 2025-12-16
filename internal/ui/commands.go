// Package ui provides terminal user interface components for the today app.
// This file contains tea.Cmd factories that wrap storage operations. These
// commands run I/O operations asynchronously to keep the Bubble Tea event
// loop responsive. Each command returns a corresponding message type defined
// in messages.go.
package ui

import (
	"time"

	"today/internal/storage"
	"today/internal/sync"

	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Task Commands
// =============================================================================

// loadTasksCmd returns a command that loads all tasks from storage.
func loadTasksCmd(store *storage.Storage) tea.Cmd {
	return func() tea.Msg {
		taskStore, err := store.LoadTasks()
		if taskStore == nil {
			return tasksLoadedMsg{err: err}
		}
		return tasksLoadedMsg{tasks: taskStore.Tasks, err: err}
	}
}

// addTaskCmd returns a command that creates a new task.
func addTaskCmd(store *storage.Storage, text, project string, priority storage.Priority, dueDate *time.Time) tea.Cmd {
	return func() tea.Msg {
		task, err := store.AddTask(text, project, priority, dueDate)
		return taskAddedMsg{task: task, err: err}
	}
}

// completeTaskCmd returns a command that marks a task as done.
// Captures task text before completing for undo description.
func completeTaskCmd(store *storage.Storage, id string) tea.Cmd {
	return func() tea.Msg {
		// Capture task text before completing for undo description
		var taskText string
		if tasks, err := store.LoadTasks(); err == nil {
			for _, t := range tasks.Tasks {
				if t.ID == id {
					taskText = t.Text
					break
				}
			}
		}

		err := store.CompleteTask(id)
		return taskCompletedMsg{id: id, text: taskText, err: err}
	}
}

// uncompleteTaskCmd returns a command that marks a task as not done.
// Captures task text before uncompleting for undo description.
func uncompleteTaskCmd(store *storage.Storage, id string) tea.Cmd {
	return func() tea.Msg {
		// Capture task text before uncompleting for undo description
		var taskText string
		if tasks, err := store.LoadTasks(); err == nil {
			for _, t := range tasks.Tasks {
				if t.ID == id {
					taskText = t.Text
					break
				}
			}
		}

		err := store.UncompleteTask(id)
		return taskUncompletedMsg{id: id, text: taskText, err: err}
	}
}

// deleteTaskCmd returns a command that removes a task.
// Captures the full task before deletion for undo restoration.
func deleteTaskCmd(store *storage.Storage, id string) tea.Cmd {
	return func() tea.Msg {
		// Capture full task before deletion for undo
		var deletedTask *storage.Task
		if tasks, err := store.LoadTasks(); err == nil {
			for _, t := range tasks.Tasks {
				if t.ID == id {
					taskCopy := t
					deletedTask = &taskCopy
					break
				}
			}
		}

		err := store.DeleteTask(id)
		return taskDeletedMsg{id: id, task: deletedTask, err: err}
	}
}

// =============================================================================
// Timer Commands
// =============================================================================

// loadTimerCmd returns a command that loads timer state from storage.
func loadTimerCmd(store *storage.Storage) tea.Cmd {
	return func() tea.Msg {
		timerStore, err := store.LoadTimer()
		return timerLoadedMsg{store: timerStore, err: err}
	}
}

// startTimerCmd returns a command that starts a timer for a project.
// If a timer is already running, it will be stopped and the new one started.
func startTimerCmd(store *storage.Storage, project string) tea.Cmd {
	return func() tea.Msg {
		err := store.StartTimer(project)
		return timerStartedMsg{project: project, err: err}
	}
}

// stopTimerCmd returns a command that stops the current timer.
func stopTimerCmd(store *storage.Storage) tea.Cmd {
	return func() tea.Msg {
		err := store.StopTimer()
		return timerStoppedMsg{err: err}
	}
}

// =============================================================================
// Habit Commands
// =============================================================================

// loadHabitsCmd returns a command that loads all habits from storage.
func loadHabitsCmd(store *storage.Storage) tea.Cmd {
	return func() tea.Msg {
		habitStore, err := store.LoadHabits()
		return habitsLoadedMsg{store: habitStore, err: err}
	}
}

// addHabitCmd returns a command that creates a new habit.
func addHabitCmd(store *storage.Storage, name, icon string) tea.Cmd {
	return func() tea.Msg {
		habit, err := store.AddHabit(name, icon)
		return habitAddedMsg{habit: habit, err: err}
	}
}

// toggleHabitCmd returns a command that toggles a habit's completion for today.
// Captures habit name and previous state for undo.
func toggleHabitCmd(store *storage.Storage, id string) tea.Cmd {
	return func() tea.Msg {
		// Capture habit name and current completion state before toggle
		var habitName string
		var wasCompleted bool
		today := time.Now().Format("2006-01-02")
		if habits, err := store.LoadHabits(); err == nil {
			// Find habit name
			for _, h := range habits.Habits {
				if h.ID == id {
					habitName = h.Name
					break
				}
			}
			// Check if completed today
			for _, log := range habits.Logs {
				if log.HabitID == id && log.Date == today {
					wasCompleted = true
					break
				}
			}
		}

		isDone, err := store.ToggleHabitToday(id)
		return habitToggledMsg{id: id, name: habitName, date: today, isDone: isDone, wasCompleted: wasCompleted, err: err}
	}
}

// deleteHabitCmd returns a command that removes a habit and its logs.
// Captures the full habit and all logs for undo restoration.
func deleteHabitCmd(store *storage.Storage, id string) tea.Cmd {
	return func() tea.Msg {
		// Capture habit and all its logs before deletion for undo
		var deletedHabit *storage.Habit
		var deletedLogs []storage.HabitLog

		if habits, err := store.LoadHabits(); err == nil {
			// Find the habit
			for _, h := range habits.Habits {
				if h.ID == id {
					habitCopy := h
					deletedHabit = &habitCopy
					break
				}
			}
			// Collect all logs for this habit
			for _, log := range habits.Logs {
				if log.HabitID == id {
					deletedLogs = append(deletedLogs, log)
				}
			}
		}

		err := store.DeleteHabit(id)
		return habitDeletedMsg{id: id, habit: deletedHabit, logs: deletedLogs, err: err}
	}
}

// =============================================================================
// Undo/Redo Commands
// =============================================================================

func undoCmd(manager *UndoManager) tea.Cmd {
	return func() tea.Msg {
		desc, err := manager.Undo()
		return undoResultMsg{desc: desc, err: err}
	}
}

func redoCmd(manager *UndoManager) tea.Cmd {
	return func() tea.Msg {
		desc, err := manager.Redo()
		return redoResultMsg{desc: desc, err: err}
	}
}

// =============================================================================
// Sync Commands
// =============================================================================

// refreshSyncStatusCmd returns a command that checks git sync status.
// Returns nil command if gitSync is nil (sync disabled).
func refreshSyncStatusCmd(gs *sync.GitSync) tea.Cmd {
	if gs == nil {
		return nil
	}
	return func() tea.Msg {
		status, err := gs.Status()
		return syncStatusMsg{status: status, err: err}
	}
}
