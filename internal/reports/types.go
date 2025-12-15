// Package reports provides daily and weekly report generation for the today app.
// Reports aggregate data from tasks, habits, and time tracking.
package reports

import (
	"time"

	"today/internal/storage"
)

// DailyReport contains aggregated data for a single day.
type DailyReport struct {
	Date        time.Time     `json:"date"`
	Tasks       TaskSummary   `json:"tasks"`
	Time        TimeSummary   `json:"time"`
	Habits      HabitSummary  `json:"habits"`
	GeneratedAt time.Time     `json:"generated_at"`
}

// WeeklyReport contains aggregated data for a week.
type WeeklyReport struct {
	StartDate     time.Time        `json:"start_date"`
	EndDate       time.Time        `json:"end_date"`
	Tasks         WeeklyTasks      `json:"tasks"`
	Time          WeeklyTime       `json:"time"`
	Habits        WeeklyHabits     `json:"habits"`
	DailyBreakdown []DailySummary  `json:"daily_breakdown"`
	GeneratedAt   time.Time        `json:"generated_at"`
}

// TaskSummary contains task statistics for a period.
type TaskSummary struct {
	Completed      []storage.Task `json:"completed"`
	Pending        []storage.Task `json:"pending"`
	CompletedCount int            `json:"completed_count"`
	PendingCount   int            `json:"pending_count"`
	AddedCount     int            `json:"added_count"`
	ByProject      []ProjectCount `json:"by_project"`
}

// ProjectCount represents a count grouped by project.
type ProjectCount struct {
	Project string `json:"project"`
	Count   int    `json:"count"`
}

// TimeSummary contains time tracking statistics for a period.
type TimeSummary struct {
	Total     time.Duration `json:"total"`
	ByProject []ProjectTime `json:"by_project"`
}

// ProjectTime represents time tracked for a specific project.
type ProjectTime struct {
	Project    string        `json:"project"`
	Duration   time.Duration `json:"duration"`
	Percentage float64       `json:"percentage"`
}

// HabitSummary contains habit statistics for a period.
type HabitSummary struct {
	Habits         []HabitStatus `json:"habits"`
	CompletedCount int           `json:"completed_count"`
	TotalCount     int           `json:"total_count"`
	CompletionRate float64       `json:"completion_rate"`
}

// HabitStatus represents a habit and its completion status.
type HabitStatus struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
	Done   bool   `json:"done"`
	Streak int    `json:"streak"`
}

// WeeklyTasks contains task statistics for a week.
type WeeklyTasks struct {
	TotalCompleted int            `json:"total_completed"`
	TotalAdded     int            `json:"total_added"`
	ByProject      []ProjectCount `json:"by_project"`
	ByDay          []DayTaskCount `json:"by_day"`
}

// DayTaskCount represents task counts for a specific day.
type DayTaskCount struct {
	Date      string `json:"date"`
	DayOfWeek string `json:"day_of_week"`
	Completed int    `json:"completed"`
	Added     int    `json:"added"`
}

// WeeklyTime contains time tracking statistics for a week.
type WeeklyTime struct {
	Total         time.Duration `json:"total"`
	DailyAverage  time.Duration `json:"daily_average"`
	ByProject     []ProjectTime `json:"by_project"`
	ByDay         []DayTime     `json:"by_day"`
}

// DayTime represents time tracked for a specific day.
type DayTime struct {
	Date      string        `json:"date"`
	DayOfWeek string        `json:"day_of_week"`
	Total     time.Duration `json:"total"`
}

// WeeklyHabits contains habit statistics for a week.
type WeeklyHabits struct {
	Habits         []WeeklyHabitStatus `json:"habits"`
	OverallRate    float64             `json:"overall_rate"`
	TotalCompleted int                 `json:"total_completed"`
	TotalExpected  int                 `json:"total_expected"`
}

// WeeklyHabitStatus represents a habit's completion over a week.
type WeeklyHabitStatus struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Icon           string   `json:"icon"`
	DaysCompleted  []bool   `json:"days_completed"` // 7 bools for each day
	CompletedCount int      `json:"completed_count"`
	CompletionRate float64  `json:"completion_rate"`
	Streak         int      `json:"streak"`
}

// DailySummary provides a quick overview of a single day within a week.
type DailySummary struct {
	Date           string        `json:"date"`
	DayOfWeek      string        `json:"day_of_week"`
	TasksCompleted int           `json:"tasks_completed"`
	TimeTracked    time.Duration `json:"time_tracked"`
	HabitsComplete int           `json:"habits_complete"`
	HabitsTotal    int           `json:"habits_total"`
}
