package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"today/internal/fsutil"
)

// SaveContext contains information about a save operation for semantic commit messages.
// It provides context about what operation was performed, enabling meaningful git commits
// like "Complete task: Review PR" instead of generic "Update tasks".
type SaveContext struct {
	Filename  string // The file being saved (e.g., "tasks.json")
	Operation string // The operation type: "add", "complete", "delete", "toggle", "start", "stop", "update"
	ItemType  string // The item type: "task", "habit", "timer"
	ItemName  string // Human-readable name (truncated task text, habit name, project name)
}

// Storage handles all file I/O operations
type Storage struct {
	dataDir           string
	onSave            func(filename string)  // Legacy callback triggered after file saves
	onSaveWithContext func(ctx SaveContext)  // Context-aware callback for semantic commits
	now               func() time.Time       // injectable clock for deterministic tests
}

const (
	dataDirPerm  os.FileMode = 0700
	dataFilePerm os.FileMode = 0600

	maxTaskTextLen  = 200
	maxProjectLen   = 60
	maxHabitNameLen = 60
	maxHabitIconLen = 12
	maxTimerProjLen = 60
)

// New creates a new Storage instance with the given data directory
func New(dataDir string) (*Storage, error) {
	// Create data directory if it doesn't exist
	if err := os.MkdirAll(dataDir, dataDirPerm); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	s := &Storage{dataDir: dataDir, now: time.Now}

	// Initialize files if they don't exist
	if err := s.initFiles(); err != nil {
		return nil, err
	}

	return s, nil
}

// SetNowFunc overrides the clock used by time-dependent storage operations.
// Passing nil resets it to time.Now.
func (s *Storage) SetNowFunc(now func() time.Time) {
	if now == nil {
		s.now = time.Now
		return
	}
	s.now = now
}

// Now returns the current time according to the storage clock.
func (s *Storage) Now() time.Time {
	if s.now == nil {
		return time.Now()
	}
	return s.now()
}

// SetOnSave registers a callback to be called after each file save.
// This is used for git sync auto-commit functionality.
// Deprecated: Use SetOnSaveWithContext for semantic commit messages.
func (s *Storage) SetOnSave(fn func(filename string)) {
	s.onSave = fn
}

// SetOnSaveWithContext registers a context-aware callback for git sync.
// The callback receives semantic information about the operation, enabling
// meaningful commit messages like "Complete task: Review PR".
func (s *Storage) SetOnSaveWithContext(fn func(ctx SaveContext)) {
	s.onSaveWithContext = fn
}

// GetDataDir returns the path to the data directory.
func (s *Storage) GetDataDir() string {
	return s.dataDir
}

// initFiles creates default JSON files if they don't exist
func (s *Storage) initFiles() error {
	// Tasks
	tasksPath := s.path("tasks.json")
	if !fileExists(tasksPath) {
		if err := s.SaveTasks(&TaskStore{Tasks: []Task{}}); err != nil {
			return err
		}
	}

	// Habits
	habitsPath := s.path("habits.json")
	if !fileExists(habitsPath) {
		if err := s.SaveHabits(&HabitStore{Habits: []Habit{}, Logs: []HabitLog{}}); err != nil {
			return err
		}
	}

	// Timer
	timerPath := s.path("timer.json")
	if !fileExists(timerPath) {
		if err := s.SaveTimer(&TimerStore{Entries: []TimerEntry{}}); err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) path(filename string) string {
	return filepath.Join(s.dataDir, filename)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func newID(prefix string) (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().UnixMilli(), hex.EncodeToString(b[:])), nil
}

func (s *Storage) writeJSONAtomic(filename string, v any) error {
	path := s.path(filename)
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("serialize %s: %w", filename, err)
	}

	// Keep a best-effort backup before overwriting.
	fsutil.BestEffortBackup(path, dataFilePerm)

	if err := fsutil.WriteFileAtomic(path, data, dataFilePerm); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}

	// Trigger legacy callback after successful write (for backward compatibility)
	if s.onSave != nil {
		s.onSave(filename)
	}

	return nil
}

// notifySaveWithContext triggers the context-aware callback if registered.
// This should be called after writeJSONAtomic when semantic context is available.
func (s *Storage) notifySaveWithContext(ctx SaveContext) {
	if s.onSaveWithContext != nil {
		s.onSaveWithContext(ctx)
	}
}

// truncateForCommit truncates a string for use in commit messages.
func truncateForCommit(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func (s *Storage) loadJSONWithRecovery(filename string, v any) error {
	path := s.path(filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := s.writeJSONAtomic(filename, v); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("read %s: %w", filename, err)
	}

	if len(bytes.TrimSpace(data)) == 0 {
		return s.recoverCorruptJSON(filename, v, fmt.Errorf("%s is empty", filename))
	}

	if err := json.Unmarshal(data, v); err == nil {
		return nil
	}
	return s.recoverCorruptJSON(filename, v, fmt.Errorf("parse %s: %w", filename, err))
}

func (s *Storage) recoverCorruptJSON(filename string, v any, cause error) error {
	path := s.path(filename)

	// Try backup first.
	bakData, bakErr := os.ReadFile(path + ".bak")
	if bakErr == nil && len(bytes.TrimSpace(bakData)) > 0 {
		if err := json.Unmarshal(bakData, v); err == nil {
			corruptPath := fmt.Sprintf("%s.corrupt.%s", path, time.Now().Format("20060102-150405"))
			_ = os.Rename(path, corruptPath)
			_ = s.writeJSONAtomic(filename, v)
			return fmt.Errorf("%s (recovered from %s.bak)", cause.Error(), filename)
		}
	}

	// No usable backup: preserve the broken file (best effort) and reset.
	corruptPath := fmt.Sprintf("%s.corrupt.%s", path, time.Now().Format("20060102-150405"))
	_ = os.Rename(path, corruptPath)
	_ = s.writeJSONAtomic(filename, v)
	return fmt.Errorf("%s (reset to defaults; original moved to %s)", cause.Error(), corruptPath)
}

// ============================================================================
// Tasks
// ============================================================================

// LoadTasks reads tasks from disk
func (s *Storage) LoadTasks() (*TaskStore, error) {
	store := TaskStore{Tasks: []Task{}}
	err := s.loadJSONWithRecovery("tasks.json", &store)
	return &store, err
}

// SaveTasks writes tasks to disk
func (s *Storage) SaveTasks(store *TaskStore) error {
	return s.writeJSONAtomic("tasks.json", store)
}

// AddTask adds a new task with optional priority and due date
func (s *Storage) AddTask(text, project string, priority Priority, dueDate *time.Time) (*Task, error) {
	text = strings.TrimSpace(text)
	project = strings.TrimSpace(project)

	if text == "" {
		return nil, fmt.Errorf("task text is required")
	}
	if len(text) > maxTaskTextLen {
		return nil, fmt.Errorf("task text too long (max %d)", maxTaskTextLen)
	}
	if len(project) > maxProjectLen {
		return nil, fmt.Errorf("project too long (max %d)", maxProjectLen)
	}

	// Validate priority if provided
	if priority != "" && priority != PriorityLow && priority != PriorityMedium && priority != PriorityHigh {
		return nil, fmt.Errorf("invalid priority: must be low, medium, or high")
	}

	store, err := s.LoadTasks()
	if err != nil {
		return nil, err
	}

	id, err := newID("t")
	if err != nil {
		return nil, err
	}

	task := Task{
		ID:        id,
		Text:      text,
		Project:   project,
		Priority:  priority,
		DueDate:   dueDate,
		Done:      false,
		CreatedAt: time.Now(),
	}

	store.Tasks = append(store.Tasks, task)

	if err := s.SaveTasks(store); err != nil {
		return nil, err
	}

	// Notify with semantic context for git commit
	s.notifySaveWithContext(SaveContext{
		Filename:  "tasks.json",
		Operation: "add",
		ItemType:  "task",
		ItemName:  truncateForCommit(task.Text, 50),
	})

	return &task, nil
}

// RestoreTask restores a previously existing task (used for undo/redo).
// It preserves the task ID and timestamps.
func (s *Storage) RestoreTask(task Task) error {
	task.Text = strings.TrimSpace(task.Text)
	task.Project = strings.TrimSpace(task.Project)

	if strings.TrimSpace(task.ID) == "" {
		return fmt.Errorf("task id is required")
	}
	if task.Text == "" {
		return fmt.Errorf("task text is required")
	}
	if len(task.Text) > maxTaskTextLen {
		return fmt.Errorf("task text too long (max %d)", maxTaskTextLen)
	}
	if len(task.Project) > maxProjectLen {
		return fmt.Errorf("project too long (max %d)", maxProjectLen)
	}
	if task.Priority != "" && task.Priority != PriorityLow && task.Priority != PriorityMedium && task.Priority != PriorityHigh {
		return fmt.Errorf("invalid priority: must be low, medium, or high")
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	if task.Done && task.CompletedAt == nil {
		now := time.Now()
		task.CompletedAt = &now
	}
	if !task.Done {
		task.CompletedAt = nil
	}

	store, err := s.LoadTasks()
	if err != nil {
		return err
	}
	for _, existing := range store.Tasks {
		if existing.ID == task.ID {
			return fmt.Errorf("task already exists: %s", task.ID)
		}
	}

	store.Tasks = append(store.Tasks, task)
	if err := s.SaveTasks(store); err != nil {
		return err
	}

	// Notify with semantic context for git commit
	s.notifySaveWithContext(SaveContext{
		Filename:  "tasks.json",
		Operation: "restore",
		ItemType:  "task",
		ItemName:  truncateForCommit(task.Text, 50),
	})

	return nil
}

// CompleteTask marks a task as done
func (s *Storage) CompleteTask(id string) error {
	store, err := s.LoadTasks()
	if err != nil {
		return err
	}

	for i := range store.Tasks {
		if store.Tasks[i].ID == id {
			taskText := store.Tasks[i].Text
			store.Tasks[i].Done = true
			now := time.Now()
			store.Tasks[i].CompletedAt = &now
			if err := s.SaveTasks(store); err != nil {
				return err
			}
			// Notify with semantic context for git commit
			s.notifySaveWithContext(SaveContext{
				Filename:  "tasks.json",
				Operation: "complete",
				ItemType:  "task",
				ItemName:  truncateForCommit(taskText, 50),
			})
			return nil
		}
	}

	return fmt.Errorf("task not found: %s", id)
}

// UncompleteTask marks a task as not done
func (s *Storage) UncompleteTask(id string) error {
	store, err := s.LoadTasks()
	if err != nil {
		return err
	}

	for i := range store.Tasks {
		if store.Tasks[i].ID == id {
			taskText := store.Tasks[i].Text
			store.Tasks[i].Done = false
			store.Tasks[i].CompletedAt = nil
			if err := s.SaveTasks(store); err != nil {
				return err
			}
			// Notify with semantic context for git commit
			s.notifySaveWithContext(SaveContext{
				Filename:  "tasks.json",
				Operation: "reopen",
				ItemType:  "task",
				ItemName:  truncateForCommit(taskText, 50),
			})
			return nil
		}
	}

	return fmt.Errorf("task not found: %s", id)
}

// DeleteTask removes a task
func (s *Storage) DeleteTask(id string) error {
	store, err := s.LoadTasks()
	if err != nil {
		return err
	}

	for i := range store.Tasks {
		if store.Tasks[i].ID == id {
			taskText := store.Tasks[i].Text
			store.Tasks = append(store.Tasks[:i], store.Tasks[i+1:]...)
			if err := s.SaveTasks(store); err != nil {
				return err
			}
			// Notify with semantic context for git commit
			s.notifySaveWithContext(SaveContext{
				Filename:  "tasks.json",
				Operation: "delete",
				ItemType:  "task",
				ItemName:  truncateForCommit(taskText, 50),
			})
			return nil
		}
	}

	return fmt.Errorf("task not found: %s", id)
}

// SortTasks sorts tasks by priority (high > medium > low) and due date (earliest first)
func (s *Storage) SortTasks(tasks []Task) []Task {
	sorted := make([]Task, len(tasks))
	copy(sorted, tasks)

	// Custom sort: completed last, then priority, due date, then creation time.
	sort.SliceStable(sorted, func(i, j int) bool {
		a := sorted[i]
		b := sorted[j]

		// Completed tasks always go to bottom.
		if a.Done != b.Done {
			return !a.Done
		}

		// Priority comparison (high=3, medium=2, low=1, none=0).
		aPrio := priorityValue(a.Priority)
		bPrio := priorityValue(b.Priority)
		if aPrio != bPrio {
			return aPrio > bPrio
		}

		// Due date comparison (tasks with due dates come before those without).
		if a.DueDate != nil && b.DueDate == nil {
			return true
		}
		if a.DueDate == nil && b.DueDate != nil {
			return false
		}
		if a.DueDate != nil && b.DueDate != nil && !a.DueDate.Equal(*b.DueDate) {
			return a.DueDate.Before(*b.DueDate)
		}

		// Finally, sort by creation time (newest first for pending tasks).
		return a.CreatedAt.After(b.CreatedAt)
	})

	return sorted
}

func priorityValue(p Priority) int {
	switch p {
	case PriorityHigh:
		return 3
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}

// ============================================================================
// Habits
// ============================================================================

// LoadHabits reads habits from disk
func (s *Storage) LoadHabits() (*HabitStore, error) {
	store := HabitStore{Habits: []Habit{}, Logs: []HabitLog{}}
	err := s.loadJSONWithRecovery("habits.json", &store)
	return &store, err
}

// SaveHabits writes habits to disk
func (s *Storage) SaveHabits(store *HabitStore) error {
	return s.writeJSONAtomic("habits.json", store)
}

// RestoreHabit restores a previously existing habit and its logs (used for undo/redo).
func (s *Storage) RestoreHabit(habit Habit, logs []HabitLog) error {
	habit.Name = strings.TrimSpace(habit.Name)
	habit.Icon = strings.TrimSpace(habit.Icon)

	if strings.TrimSpace(habit.ID) == "" {
		return fmt.Errorf("habit id is required")
	}
	if habit.Name == "" {
		return fmt.Errorf("habit name is required")
	}
	if len(habit.Name) > maxHabitNameLen {
		return fmt.Errorf("habit name too long (max %d)", maxHabitNameLen)
	}
	if habit.Icon == "" {
		return fmt.Errorf("habit icon is required")
	}
	if len(habit.Icon) > maxHabitIconLen {
		return fmt.Errorf("habit icon too long (max %d)", maxHabitIconLen)
	}
	if habit.CreatedAt.IsZero() {
		habit.CreatedAt = time.Now()
	}

	store, err := s.LoadHabits()
	if err != nil {
		return err
	}
	for _, existing := range store.Habits {
		if existing.ID == habit.ID {
			return fmt.Errorf("habit already exists: %s", habit.ID)
		}
	}

	store.Habits = append(store.Habits, habit)

	// Restore logs, deduping by (habit_id,date).
	seen := make(map[string]struct{}, len(store.Logs)+len(logs))
	for _, log := range store.Logs {
		seen[log.HabitID+"|"+log.Date] = struct{}{}
	}
	for _, log := range logs {
		date := strings.TrimSpace(log.Date)
		if date == "" {
			continue
		}
		key := habit.ID + "|" + date
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		store.Logs = append(store.Logs, HabitLog{HabitID: habit.ID, Date: date})
	}

	if err := s.SaveHabits(store); err != nil {
		return err
	}

	// Notify with semantic context for git commit
	s.notifySaveWithContext(SaveContext{
		Filename:  "habits.json",
		Operation: "restore",
		ItemType:  "habit",
		ItemName:  truncateForCommit(habit.Name, 50),
	})

	return nil
}

// SetHabitDoneOnDate sets a habit's completion for a specific YYYY-MM-DD date.
func (s *Storage) SetHabitDoneOnDate(habitID, date string, done bool) error {
	habitID = strings.TrimSpace(habitID)
	date = strings.TrimSpace(date)
	if habitID == "" {
		return fmt.Errorf("habit id is required")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return fmt.Errorf("invalid date %q: expected YYYY-MM-DD", date)
	}

	store, err := s.LoadHabits()
	if err != nil {
		return err
	}

	foundHabit := false
	for _, h := range store.Habits {
		if h.ID == habitID {
			foundHabit = true
			break
		}
	}
	if !foundHabit {
		return fmt.Errorf("habit not found: %s", habitID)
	}

	newLogs := store.Logs[:0]
	for _, log := range store.Logs {
		if log.HabitID == habitID && log.Date == date {
			continue
		}
		newLogs = append(newLogs, log)
	}
	store.Logs = newLogs

	if done {
		store.Logs = append(store.Logs, HabitLog{HabitID: habitID, Date: date})
	}
	return s.SaveHabits(store)
}

// AddHabit creates a new habit
func (s *Storage) AddHabit(name, icon string) (*Habit, error) {
	name = strings.TrimSpace(name)
	icon = strings.TrimSpace(icon)

	if name == "" {
		return nil, fmt.Errorf("habit name is required")
	}
	if len(name) > maxHabitNameLen {
		return nil, fmt.Errorf("habit name too long (max %d)", maxHabitNameLen)
	}
	if icon == "" {
		return nil, fmt.Errorf("habit icon is required")
	}
	if len(icon) > maxHabitIconLen {
		return nil, fmt.Errorf("habit icon too long (max %d)", maxHabitIconLen)
	}

	store, err := s.LoadHabits()
	if err != nil {
		return nil, err
	}

	id, err := newID("h")
	if err != nil {
		return nil, err
	}

	habit := Habit{
		ID:        id,
		Name:      name,
		Icon:      icon,
		CreatedAt: time.Now(),
	}

	store.Habits = append(store.Habits, habit)

	if err := s.SaveHabits(store); err != nil {
		return nil, err
	}

	// Notify with semantic context for git commit
	s.notifySaveWithContext(SaveContext{
		Filename:  "habits.json",
		Operation: "add",
		ItemType:  "habit",
		ItemName:  truncateForCommit(habit.Name, 50),
	})

	return &habit, nil
}

// ToggleHabitToday toggles a habit for today
func (s *Storage) ToggleHabitToday(habitID string) (bool, error) {
	today := s.Now().Format("2006-01-02")
	store, err := s.LoadHabits()
	if err != nil {
		return false, err
	}

	// Find habit name for context
	var habitName string
	for _, h := range store.Habits {
		if h.ID == habitID {
			habitName = h.Name
			break
		}
	}

	wasDone := s.IsHabitDoneOnDate(store, habitID, today)
	if err := s.SetHabitDoneOnDate(habitID, today, !wasDone); err != nil {
		return false, err
	}

	// Notify with semantic context for git commit
	s.notifySaveWithContext(SaveContext{
		Filename:  "habits.json",
		Operation: "toggle",
		ItemType:  "habit",
		ItemName:  truncateForCommit(habitName, 50),
	})

	return !wasDone, nil
}

// IsHabitDoneOnDate checks if a habit was completed on a specific date
func (s *Storage) IsHabitDoneOnDate(store *HabitStore, habitID, date string) bool {
	for _, log := range store.Logs {
		if log.HabitID == habitID && log.Date == date {
			return true
		}
	}
	return false
}

// GetHabitStreak calculates the current streak for a habit
func (s *Storage) GetHabitStreak(store *HabitStore, habitID string) int {
	return s.GetHabitStreakAt(store, habitID, s.Now())
}

// GetHabitStreakAt calculates the streak for a habit as of the given time.
// If the habit is not completed on that date, it counts back starting from the previous day.
func (s *Storage) GetHabitStreakAt(store *HabitStore, habitID string, at time.Time) int {
	streak := 0
	date := startOfDay(at)

	// If the target date isn't done, start from the previous day.
	if !s.IsHabitDoneOnDate(store, habitID, date.Format("2006-01-02")) {
		date = date.AddDate(0, 0, -1)
	}

	for {
		dateStr := date.Format("2006-01-02")
		if s.IsHabitDoneOnDate(store, habitID, dateStr) {
			streak++
			date = date.AddDate(0, 0, -1)
			continue
		}
		break
	}

	return streak
}

// GetHabitWeek returns the last 7 days of a habit (for display)
func (s *Storage) GetHabitWeek(store *HabitStore, habitID string) []bool {
	week := make([]bool, 7)
	today := s.Now()

	for i := 6; i >= 0; i-- {
		date := today.AddDate(0, 0, -(6 - i))
		dateStr := date.Format("2006-01-02")
		week[i] = s.IsHabitDoneOnDate(store, habitID, dateStr)
	}

	return week
}

// DeleteHabit removes a habit and its logs
func (s *Storage) DeleteHabit(id string) error {
	store, err := s.LoadHabits()
	if err != nil {
		return err
	}

	// Remove habit and capture name for context
	var habitName string
	found := false
	for i := range store.Habits {
		if store.Habits[i].ID == id {
			habitName = store.Habits[i].Name
			store.Habits = append(store.Habits[:i], store.Habits[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("habit not found: %s", id)
	}

	// Remove associated logs
	newLogs := []HabitLog{}
	for _, log := range store.Logs {
		if log.HabitID != id {
			newLogs = append(newLogs, log)
		}
	}
	store.Logs = newLogs

	if err := s.SaveHabits(store); err != nil {
		return err
	}

	// Notify with semantic context for git commit
	s.notifySaveWithContext(SaveContext{
		Filename:  "habits.json",
		Operation: "delete",
		ItemType:  "habit",
		ItemName:  truncateForCommit(habitName, 50),
	})

	return nil
}

// ============================================================================
// Timer
// ============================================================================

// LoadTimer reads timer state from disk
func (s *Storage) LoadTimer() (*TimerStore, error) {
	store := TimerStore{Entries: []TimerEntry{}}
	err := s.loadJSONWithRecovery("timer.json", &store)
	return &store, err
}

// SaveTimer writes timer state to disk
func (s *Storage) SaveTimer(store *TimerStore) error {
	return s.writeJSONAtomic("timer.json", store)
}

// StartTimer starts a new timer for a project
func (s *Storage) StartTimer(project string) error {
	project = strings.TrimSpace(project)

	if project == "" {
		return fmt.Errorf("project is required")
	}
	if len(project) > maxTimerProjLen {
		return fmt.Errorf("project too long (max %d)", maxTimerProjLen)
	}

	store, err := s.LoadTimer()
	if err != nil {
		return err
	}

	now := time.Now()

	// Stop any existing timer first
	if store.Current != nil {
		entry := TimerEntry{
			Project:   store.Current.Project,
			StartedAt: store.Current.StartedAt,
			EndedAt:   now,
		}
		store.Entries = append(store.Entries, entry)
	}

	store.Current = &CurrentTimer{
		Project:   project,
		StartedAt: now,
	}

	if err := s.SaveTimer(store); err != nil {
		return err
	}

	// Notify with semantic context for git commit
	s.notifySaveWithContext(SaveContext{
		Filename:  "timer.json",
		Operation: "start",
		ItemType:  "timer",
		ItemName:  truncateForCommit(project, 50),
	})

	return nil
}

// StopTimer stops the current timer
func (s *Storage) StopTimer() error {
	store, err := s.LoadTimer()
	if err != nil {
		return err
	}

	if store.Current == nil {
		return nil // Nothing to stop
	}

	projectName := store.Current.Project
	now := time.Now()
	entry := TimerEntry{
		Project:   store.Current.Project,
		StartedAt: store.Current.StartedAt,
		EndedAt:   now,
	}
	store.Entries = append(store.Entries, entry)
	store.Current = nil

	if err := s.SaveTimer(store); err != nil {
		return err
	}

	// Notify with semantic context for git commit
	s.notifySaveWithContext(SaveContext{
		Filename:  "timer.json",
		Operation: "stop",
		ItemType:  "timer",
		ItemName:  truncateForCommit(projectName, 50),
	})

	return nil
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func startOfWeekSunday(t time.Time) time.Time {
	dayStart := startOfDay(t)
	return dayStart.AddDate(0, 0, -int(dayStart.Weekday()))
}

func overlapDuration(aStart, aEnd, bStart, bEnd time.Time) time.Duration {
	if aEnd.Before(aStart) || bEnd.Before(bStart) {
		return 0
	}
	start := aStart
	if bStart.After(start) {
		start = bStart
	}
	end := aEnd
	if bEnd.Before(end) {
		end = bEnd
	}
	if end.After(start) {
		return end.Sub(start)
	}
	return 0
}

// GetTodayTotal returns total time tracked today
func (s *Storage) GetTodayTotal(store *TimerStore) time.Duration {
	return s.getTodayTotalAt(store, time.Now())
}

func (s *Storage) getTodayTotalAt(store *TimerStore, now time.Time) time.Duration {
	var total time.Duration
	dayStart := startOfDay(now)
	dayEnd := dayStart.AddDate(0, 0, 1)

	for _, entry := range store.Entries {
		total += overlapDuration(dayStart, dayEnd, entry.StartedAt, entry.EndedAt)
	}

	// Add current timer if running
	if store.Current != nil {
		total += overlapDuration(dayStart, dayEnd, store.Current.StartedAt, now)
	}

	return total
}

// GetWeekTotal returns total time tracked this week
func (s *Storage) GetWeekTotal(store *TimerStore) time.Duration {
	return s.getWeekTotalAt(store, time.Now())
}

func (s *Storage) getWeekTotalAt(store *TimerStore, now time.Time) time.Duration {
	var total time.Duration
	weekStart := startOfWeekSunday(now)
	weekEnd := weekStart.AddDate(0, 0, 7)

	for _, entry := range store.Entries {
		total += overlapDuration(weekStart, weekEnd, entry.StartedAt, entry.EndedAt)
	}

	// Add current timer if running
	if store.Current != nil {
		total += overlapDuration(weekStart, weekEnd, store.Current.StartedAt, now)
	}

	return total
}

// ============================================================================
// Timer Analytics
// ============================================================================

// ProjectTotal represents total time spent on a project
type ProjectTotal struct {
	Project string
	Total   time.Duration
}

// GetProjectTotals returns total time spent per project, sorted by time (descending)
func (s *Storage) GetProjectTotals(store *TimerStore) []ProjectTotal {
	projectTime := make(map[string]time.Duration)

	// Sum up all entries
	for _, entry := range store.Entries {
		duration := entry.EndedAt.Sub(entry.StartedAt)
		projectTime[entry.Project] += duration
	}

	// Add current timer if running
	if store.Current != nil {
		elapsed := time.Since(store.Current.StartedAt)
		projectTime[store.Current.Project] += elapsed
	}

	// Convert to slice and sort
	var totals []ProjectTotal
	for project, total := range projectTime {
		totals = append(totals, ProjectTotal{
			Project: project,
			Total:   total,
		})
	}

	// Sort by total time (descending)
	for i := 0; i < len(totals); i++ {
		for j := i + 1; j < len(totals); j++ {
			if totals[j].Total > totals[i].Total {
				totals[i], totals[j] = totals[j], totals[i]
			}
		}
	}

	return totals
}

// DayBreakdown represents time tracked for a specific day
type DayBreakdown struct {
	Date  string // YYYY-MM-DD format
	Total time.Duration
}

// GetDailyBreakdown returns time tracked per day for the last N days
func (s *Storage) GetDailyBreakdown(store *TimerStore, days int) []DayBreakdown {
	dayTotals := make(map[string]time.Duration)
	now := time.Now()

	// Process all entries
	for _, entry := range store.Entries {
		// Calculate overlap with each day
		for i := 0; i < days; i++ {
			dayStart := startOfDay(now.AddDate(0, 0, -i))
			dayEnd := dayStart.AddDate(0, 0, 1)
			dateKey := dayStart.Format("2006-01-02")

			overlap := overlapDuration(dayStart, dayEnd, entry.StartedAt, entry.EndedAt)
			if overlap > 0 {
				dayTotals[dateKey] += overlap
			}
		}
	}

	// Add current timer if running
	if store.Current != nil {
		for i := 0; i < days; i++ {
			dayStart := startOfDay(now.AddDate(0, 0, -i))
			dayEnd := dayStart.AddDate(0, 0, 1)
			dateKey := dayStart.Format("2006-01-02")

			overlap := overlapDuration(dayStart, dayEnd, store.Current.StartedAt, now)
			if overlap > 0 {
				dayTotals[dateKey] += overlap
			}
		}
	}

	// Convert to slice, sorted by date (descending)
	var breakdown []DayBreakdown
	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		dateKey := date.Format("2006-01-02")
		breakdown = append(breakdown, DayBreakdown{
			Date:  dateKey,
			Total: dayTotals[dateKey],
		})
	}

	return breakdown
}

// ============================================================================
// Data Export
// ============================================================================

// ExportTasksJSON exports tasks to JSON format
func (s *Storage) ExportTasksJSON() ([]byte, error) {
	store, err := s.LoadTasks()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(store, "", "  ")
}

// ExportHabitsJSON exports habits to JSON format
func (s *Storage) ExportHabitsJSON() ([]byte, error) {
	store, err := s.LoadHabits()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(store, "", "  ")
}

// ExportTimerJSON exports timer entries to JSON format
func (s *Storage) ExportTimerJSON() ([]byte, error) {
	store, err := s.LoadTimer()
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(store, "", "  ")
}

// ExportTasksCSV exports tasks to CSV format
func (s *Storage) ExportTasksCSV() (string, error) {
	store, err := s.LoadTasks()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("ID,Text,Project,Priority,DueDate,Done,CreatedAt,CompletedAt\n")

	for _, task := range store.Tasks {
		dueDate := ""
		if task.DueDate != nil {
			dueDate = task.DueDate.Format("2006-01-02")
		}
		completedAt := ""
		if task.CompletedAt != nil {
			completedAt = task.CompletedAt.Format("2006-01-02 15:04:05")
		}

		// CSV escape: wrap in quotes if contains comma
		text := task.Text
		if strings.Contains(text, ",") || strings.Contains(text, "\"") {
			text = "\"" + strings.ReplaceAll(text, "\"", "\"\"") + "\""
		}

		b.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%t,%s,%s\n",
			task.ID,
			text,
			task.Project,
			task.Priority,
			dueDate,
			task.Done,
			task.CreatedAt.Format("2006-01-02 15:04:05"),
			completedAt,
		))
	}

	return b.String(), nil
}

// ExportTimerCSV exports timer entries to CSV format
func (s *Storage) ExportTimerCSV() (string, error) {
	store, err := s.LoadTimer()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("Project,StartedAt,EndedAt,Duration\n")

	for _, entry := range store.Entries {
		duration := entry.EndedAt.Sub(entry.StartedAt)
		b.WriteString(fmt.Sprintf("%s,%s,%s,%s\n",
			entry.Project,
			entry.StartedAt.Format("2006-01-02 15:04:05"),
			entry.EndedAt.Format("2006-01-02 15:04:05"),
			duration,
		))
	}

	return b.String(), nil
}
