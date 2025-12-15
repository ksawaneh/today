// Package reports provides daily and weekly report generation for the today app.
package reports

import (
	"sort"
	"time"

	"today/internal/storage"
)

// Generator creates reports from storage data.
type Generator struct {
	store *storage.Storage
}

// NewGenerator creates a new report generator.
func NewGenerator(store *storage.Storage) *Generator {
	return &Generator{store: store}
}

// GenerateDaily generates a report for a specific date.
func (g *Generator) GenerateDaily(date time.Time) (*DailyReport, error) {
	date = startOfDay(date)
	end := date.AddDate(0, 0, 1)

	tasks, err := g.getTaskSummary(date, end)
	if err != nil {
		return nil, err
	}

	timeSummary, err := g.getTimeSummary(date, end)
	if err != nil {
		return nil, err
	}

	habits, err := g.getHabitSummary(date)
	if err != nil {
		return nil, err
	}

	return &DailyReport{
		Date:        date,
		Tasks:       tasks,
		Time:        timeSummary,
		Habits:      habits,
		GeneratedAt: time.Now(),
	}, nil
}

// GenerateWeekly generates a report for a week starting on the given date.
func (g *Generator) GenerateWeekly(startDate time.Time) (*WeeklyReport, error) {
	// Align to start of week (Sunday)
	startDate = startOfWeekSunday(startDate)
	endDate := startDate.AddDate(0, 0, 7)

	weeklyTasks, err := g.getWeeklyTasks(startDate, endDate)
	if err != nil {
		return nil, err
	}

	weeklyTime, err := g.getWeeklyTime(startDate, endDate)
	if err != nil {
		return nil, err
	}

	weeklyHabits, err := g.getWeeklyHabits(startDate, endDate)
	if err != nil {
		return nil, err
	}

	dailyBreakdown, err := g.getDailyBreakdown(startDate, endDate)
	if err != nil {
		return nil, err
	}

	return &WeeklyReport{
		StartDate:      startDate,
		EndDate:        endDate.Add(-time.Nanosecond), // End of last day
		Tasks:          weeklyTasks,
		Time:           weeklyTime,
		Habits:         weeklyHabits,
		DailyBreakdown: dailyBreakdown,
		GeneratedAt:    time.Now(),
	}, nil
}

// getTaskSummary returns task statistics for a date range.
func (g *Generator) getTaskSummary(start, end time.Time) (TaskSummary, error) {
	taskStore, err := g.store.LoadTasks()
	if err != nil {
		return TaskSummary{}, err
	}

	var completed, pending []storage.Task
	projectCounts := make(map[string]int)
	addedCount := 0

	for _, task := range taskStore.Tasks {
		// Check if task was added in this period
		if !task.CreatedAt.Before(start) && task.CreatedAt.Before(end) {
			addedCount++
		}

		// Check if task was completed in this period
		if task.Done && task.CompletedAt != nil {
			if !task.CompletedAt.Before(start) && task.CompletedAt.Before(end) {
				completed = append(completed, task)
				project := task.Project
				if project == "" {
					project = "General"
				}
				projectCounts[project]++
			}
		} else if !task.Done {
			pending = append(pending, task)
		}
	}

	// Convert project counts to sorted slice
	byProject := make([]ProjectCount, 0, len(projectCounts))
	for project, count := range projectCounts {
		byProject = append(byProject, ProjectCount{
			Project: project,
			Count:   count,
		})
	}
	sort.Slice(byProject, func(i, j int) bool {
		return byProject[i].Count > byProject[j].Count
	})

	return TaskSummary{
		Completed:      completed,
		Pending:        pending,
		CompletedCount: len(completed),
		PendingCount:   len(pending),
		AddedCount:     addedCount,
		ByProject:      byProject,
	}, nil
}

// getTimeSummary returns time tracking statistics for a date range.
func (g *Generator) getTimeSummary(start, end time.Time) (TimeSummary, error) {
	timerStore, err := g.store.LoadTimer()
	if err != nil {
		return TimeSummary{}, err
	}

	projectDurations := make(map[string]time.Duration)
	var total time.Duration

	for _, entry := range timerStore.Entries {
		// Calculate overlap with the date range
		overlap := overlapDuration(entry.StartedAt, entry.EndedAt, start, end)
		if overlap > 0 {
			total += overlap
			projectDurations[entry.Project] += overlap
		}
	}

	// Include current timer if running
	if timerStore.Current != nil {
		overlap := overlapDuration(timerStore.Current.StartedAt, time.Now(), start, end)
		if overlap > 0 {
			total += overlap
			projectDurations[timerStore.Current.Project] += overlap
		}
	}

	// Convert to sorted slice with percentages
	byProject := make([]ProjectTime, 0, len(projectDurations))
	for project, duration := range projectDurations {
		pct := 0.0
		if total > 0 {
			pct = float64(duration) / float64(total) * 100
		}
		byProject = append(byProject, ProjectTime{
			Project:    project,
			Duration:   duration,
			Percentage: pct,
		})
	}
	sort.Slice(byProject, func(i, j int) bool {
		return byProject[i].Duration > byProject[j].Duration
	})

	return TimeSummary{
		Total:     total,
		ByProject: byProject,
	}, nil
}

// getHabitSummary returns habit statistics for a specific date.
func (g *Generator) getHabitSummary(date time.Time) (HabitSummary, error) {
	habitStore, err := g.store.LoadHabits()
	if err != nil {
		return HabitSummary{}, err
	}

	dateStr := date.Format("2006-01-02")
	var statuses []HabitStatus
	completedCount := 0

	for _, habit := range habitStore.Habits {
		done := g.store.IsHabitDoneOnDate(habitStore, habit.ID, dateStr)
		streak := g.store.GetHabitStreakAt(habitStore, habit.ID, date)

		if done {
			completedCount++
		}

		statuses = append(statuses, HabitStatus{
			ID:     habit.ID,
			Name:   habit.Name,
			Icon:   habit.Icon,
			Done:   done,
			Streak: streak,
		})
	}

	rate := 0.0
	if len(statuses) > 0 {
		rate = float64(completedCount) / float64(len(statuses)) * 100
	}

	return HabitSummary{
		Habits:         statuses,
		CompletedCount: completedCount,
		TotalCount:     len(statuses),
		CompletionRate: rate,
	}, nil
}

// getWeeklyTasks returns task statistics for a week.
func (g *Generator) getWeeklyTasks(start, end time.Time) (WeeklyTasks, error) {
	taskStore, err := g.store.LoadTasks()
	if err != nil {
		return WeeklyTasks{}, err
	}

	projectCounts := make(map[string]int)
	totalCompleted := 0
	totalAdded := 0
	byDay := make([]DayTaskCount, 7)

	// Initialize days
	for i := 0; i < 7; i++ {
		day := start.AddDate(0, 0, i)
		byDay[i] = DayTaskCount{
			Date:      day.Format("2006-01-02"),
			DayOfWeek: day.Format("Mon"),
		}
	}

	for _, task := range taskStore.Tasks {
		// Count added tasks
		if !task.CreatedAt.Before(start) && task.CreatedAt.Before(end) {
			totalAdded++
			dayIdx := dayIndexInRange(task.CreatedAt, start, 7)
			if dayIdx >= 0 && dayIdx < 7 {
				byDay[dayIdx].Added++
			}
		}

		// Count completed tasks
		if task.Done && task.CompletedAt != nil {
			if !task.CompletedAt.Before(start) && task.CompletedAt.Before(end) {
				totalCompleted++

				project := task.Project
				if project == "" {
					project = "General"
				}
				projectCounts[project]++

				dayIdx := dayIndexInRange(*task.CompletedAt, start, 7)
				if dayIdx >= 0 && dayIdx < 7 {
					byDay[dayIdx].Completed++
				}
			}
		}
	}

	// Convert project counts
	byProject := make([]ProjectCount, 0, len(projectCounts))
	for project, count := range projectCounts {
		byProject = append(byProject, ProjectCount{
			Project: project,
			Count:   count,
		})
	}
	sort.Slice(byProject, func(i, j int) bool {
		return byProject[i].Count > byProject[j].Count
	})

	return WeeklyTasks{
		TotalCompleted: totalCompleted,
		TotalAdded:     totalAdded,
		ByProject:      byProject,
		ByDay:          byDay,
	}, nil
}

// getWeeklyTime returns time statistics for a week.
func (g *Generator) getWeeklyTime(start, end time.Time) (WeeklyTime, error) {
	timerStore, err := g.store.LoadTimer()
	if err != nil {
		return WeeklyTime{}, err
	}

	projectDurations := make(map[string]time.Duration)
	var total time.Duration
	byDay := make([]DayTime, 7)

	// Initialize days
	for i := 0; i < 7; i++ {
		day := start.AddDate(0, 0, i)
		byDay[i] = DayTime{
			Date:      day.Format("2006-01-02"),
			DayOfWeek: day.Format("Mon"),
		}
	}

	// Process entries
	for _, entry := range timerStore.Entries {
		overlap := overlapDuration(entry.StartedAt, entry.EndedAt, start, end)
		if overlap > 0 {
			total += overlap
			projectDurations[entry.Project] += overlap

			// Split across days.
			for i := 0; i < 7; i++ {
				dayStart := start.AddDate(0, 0, i)
				dayEnd := start.AddDate(0, 0, i+1)
				dayOverlap := overlapDuration(entry.StartedAt, entry.EndedAt, dayStart, dayEnd)
				if dayOverlap > 0 {
					byDay[i].Total += dayOverlap
				}
			}
		}
	}

	// Include current timer
	if timerStore.Current != nil {
		overlap := overlapDuration(timerStore.Current.StartedAt, time.Now(), start, end)
		if overlap > 0 {
			total += overlap
			projectDurations[timerStore.Current.Project] += overlap

			for i := 0; i < 7; i++ {
				dayStart := start.AddDate(0, 0, i)
				dayEnd := start.AddDate(0, 0, i+1)
				dayOverlap := overlapDuration(timerStore.Current.StartedAt, time.Now(), dayStart, dayEnd)
				if dayOverlap > 0 {
					byDay[i].Total += dayOverlap
				}
			}
		}
	}

	// Convert to sorted slice
	byProject := make([]ProjectTime, 0, len(projectDurations))
	for project, duration := range projectDurations {
		pct := 0.0
		if total > 0 {
			pct = float64(duration) / float64(total) * 100
		}
		byProject = append(byProject, ProjectTime{
			Project:    project,
			Duration:   duration,
			Percentage: pct,
		})
	}
	sort.Slice(byProject, func(i, j int) bool {
		return byProject[i].Duration > byProject[j].Duration
	})

	// Calculate daily average
	dailyAvg := time.Duration(0)
	if total > 0 {
		dailyAvg = total / 7
	}

	return WeeklyTime{
		Total:        total,
		DailyAverage: dailyAvg,
		ByProject:    byProject,
		ByDay:        byDay,
	}, nil
}

// getWeeklyHabits returns habit statistics for a week.
func (g *Generator) getWeeklyHabits(start, end time.Time) (WeeklyHabits, error) {
	habitStore, err := g.store.LoadHabits()
	if err != nil {
		return WeeklyHabits{}, err
	}

	var statuses []WeeklyHabitStatus
	totalCompleted := 0
	totalExpected := 0
	weekEnd := end.Add(-time.Nanosecond)

	for _, habit := range habitStore.Habits {
		daysCompleted := make([]bool, 7)

		for i := 0; i < 7; i++ {
			day := start.AddDate(0, 0, i)
			dateStr := day.Format("2006-01-02")
			done := g.store.IsHabitDoneOnDate(habitStore, habit.ID, dateStr)
			daysCompleted[i] = done
		}

		expectedCount := expectedCountForWeek(habit, start)
		totalExpected += expectedCount

		completedCount := completedCountForWeek(habit, daysCompleted)
		totalCompleted += completedCount

		rate := 0.0
		if expectedCount > 0 {
			rate = float64(completedCount) / float64(expectedCount) * 100
		}
		streak := g.store.GetHabitStreakAt(habitStore, habit.ID, weekEnd)

		statuses = append(statuses, WeeklyHabitStatus{
			ID:             habit.ID,
			Name:           habit.Name,
			Icon:           habit.Icon,
			DaysCompleted:  daysCompleted,
			CompletedCount: completedCount,
			CompletionRate: rate,
			Streak:         streak,
		})
	}

	overallRate := 0.0
	if totalExpected > 0 {
		overallRate = float64(totalCompleted) / float64(totalExpected) * 100
	}

	return WeeklyHabits{
		Habits:         statuses,
		OverallRate:    overallRate,
		TotalCompleted: totalCompleted,
		TotalExpected:  totalExpected,
	}, nil
}

// getDailyBreakdown returns a summary for each day in the period.
func (g *Generator) getDailyBreakdown(start, end time.Time) ([]DailySummary, error) {
	days := daysBetween(start, end)
	breakdown := make([]DailySummary, 0, days)

	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i)
		dayEnd := day.AddDate(0, 0, 1)

		tasks, err := g.getTaskSummary(day, dayEnd)
		if err != nil {
			return nil, err
		}

		timeSummary, err := g.getTimeSummary(day, dayEnd)
		if err != nil {
			return nil, err
		}

		habits, err := g.getHabitSummary(day)
		if err != nil {
			return nil, err
		}

		breakdown = append(breakdown, DailySummary{
			Date:           day.Format("2006-01-02"),
			DayOfWeek:      day.Format("Mon"),
			TasksCompleted: tasks.CompletedCount,
			TimeTracked:    timeSummary.Total,
			HabitsComplete: habits.CompletedCount,
			HabitsTotal:    habits.TotalCount,
		})
	}

	return breakdown, nil
}

// Helper functions

// startOfDay returns the start of the day (midnight).
func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// startOfWeekSunday returns the start of the week (Sunday).
func startOfWeekSunday(t time.Time) time.Time {
	t = startOfDay(t)
	weekday := int(t.Weekday())
	return t.AddDate(0, 0, -weekday)
}

func daysBetween(start, end time.Time) int {
	if end.Before(start) || end.Equal(start) {
		return 0
	}
	count := 0
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		count++
		if count > 3660 {
			break
		}
	}
	return count
}

func dayIndexInRange(t time.Time, start time.Time, days int) int {
	for i := 0; i < days; i++ {
		dayStart := start.AddDate(0, 0, i)
		dayEnd := start.AddDate(0, 0, i+1)
		if !t.Before(dayStart) && t.Before(dayEnd) {
			return i
		}
	}
	return -1
}

func expectedCountForWeek(h storage.Habit, weekStart time.Time) int {
	freq := h.Frequency
	if freq == "" {
		freq = storage.FrequencyDaily
	}

	switch freq {
	case storage.FrequencyWeekly:
		return 1
	case storage.FrequencyWeekdays:
		count := 0
		for i := 0; i < 7; i++ {
			day := weekStart.AddDate(0, 0, i)
			if day.Weekday() >= time.Monday && day.Weekday() <= time.Friday {
				count++
			}
		}
		return count
	case storage.FrequencyCustom:
		allowed := make(map[int]struct{}, len(h.CustomDays))
		for _, d := range h.CustomDays {
			if d >= 0 && d <= 6 {
				allowed[d] = struct{}{}
			}
		}
		count := 0
		for i := 0; i < 7; i++ {
			day := weekStart.AddDate(0, 0, i)
			if _, ok := allowed[int(day.Weekday())]; ok {
				count++
			}
		}
		return count
	default:
		return 7
	}
}

func completedCountForWeek(h storage.Habit, daysCompleted []bool) int {
	freq := h.Frequency
	if freq == "" {
		freq = storage.FrequencyDaily
	}

	switch freq {
	case storage.FrequencyWeekly:
		for _, done := range daysCompleted {
			if done {
				return 1
			}
		}
		return 0
	case storage.FrequencyWeekdays:
		count := 0
		for i, done := range daysCompleted {
			day := time.Weekday(i) // i is offset from Sunday
			if day >= time.Monday && day <= time.Friday && done {
				count++
			}
		}
		return count
	case storage.FrequencyCustom:
		allowed := make(map[int]struct{}, len(h.CustomDays))
		for _, d := range h.CustomDays {
			if d >= 0 && d <= 6 {
				allowed[d] = struct{}{}
			}
		}
		count := 0
		for i, done := range daysCompleted {
			if !done {
				continue
			}
			if _, ok := allowed[i]; ok {
				count++
			}
		}
		return count
	default:
		count := 0
		for _, done := range daysCompleted {
			if done {
				count++
			}
		}
		return count
	}
}

// overlapDuration calculates how much of [entryStart, entryEnd] overlaps with [rangeStart, rangeEnd].
func overlapDuration(entryStart, entryEnd, rangeStart, rangeEnd time.Time) time.Duration {
	// Find the overlap
	overlapStart := entryStart
	if rangeStart.After(overlapStart) {
		overlapStart = rangeStart
	}

	overlapEnd := entryEnd
	if rangeEnd.Before(overlapEnd) {
		overlapEnd = rangeEnd
	}

	// Check if there's any overlap
	if overlapEnd.Before(overlapStart) || overlapEnd.Equal(overlapStart) {
		return 0
	}

	return overlapEnd.Sub(overlapStart)
}
