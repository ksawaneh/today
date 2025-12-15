package storage

import "time"

// Priority represents task priority levels
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityNone   Priority = "" // Default for backward compatibility
)

// Task represents a single todo item
type Task struct {
	ID          string     `json:"id"`
	Text        string     `json:"text"`
	Project     string     `json:"project,omitempty"`
	Priority    Priority   `json:"priority,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Done        bool       `json:"done"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// TaskStore holds all tasks
type TaskStore struct {
	Tasks []Task `json:"tasks"`
}

// HabitFrequency represents how often a habit should be done
type HabitFrequency string

const (
	FrequencyDaily   HabitFrequency = "daily"
	FrequencyWeekly  HabitFrequency = "weekly"
	FrequencyWeekdays HabitFrequency = "weekdays"  // Monday-Friday
	FrequencyCustom  HabitFrequency = "custom"      // Specific days of week
)

// Habit represents a trackable habit
type Habit struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Icon        string         `json:"icon"`
	Frequency   HabitFrequency `json:"frequency,omitempty"` // Default: daily for backward compatibility
	CustomDays  []int          `json:"custom_days,omitempty"` // 0=Sunday, 1=Monday, etc.
	CreatedAt   time.Time      `json:"created_at"`
}

// HabitLog represents a single habit completion
type HabitLog struct {
	HabitID string `json:"habit_id"`
	Date    string `json:"date"` // YYYY-MM-DD format
}

// HabitStore holds habits and their logs
type HabitStore struct {
	Habits []Habit    `json:"habits"`
	Logs   []HabitLog `json:"logs"`
}

// TimerEntry represents a completed time tracking entry
type TimerEntry struct {
	Project   string    `json:"project"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
}

// CurrentTimer represents the actively running timer (if any)
type CurrentTimer struct {
	Project   string    `json:"project"`
	StartedAt time.Time `json:"started_at"`
}

// TimerStore holds timer state and history
type TimerStore struct {
	Current *CurrentTimer `json:"current,omitempty"`
	Entries []TimerEntry  `json:"entries"`
}
