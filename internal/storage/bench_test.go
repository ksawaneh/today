package storage

import (
	"fmt"
	"testing"
	"time"
)

// BenchmarkAddTask measures task creation performance
func BenchmarkAddTask(b *testing.B) {
	store := createBenchStorage(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.AddTask(fmt.Sprintf("Task %d", i), "", PriorityNone, nil)
		if err != nil {
			b.Fatalf("AddTask failed: %v", err)
		}
	}
}

// BenchmarkLoadTasks measures task loading performance with varying sizes
func BenchmarkLoadTasks(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			store := createBenchStorage(b)

			// Populate with tasks
			for i := 0; i < size; i++ {
				store.AddTask(fmt.Sprintf("Task %d", i), "project", PriorityNone, nil)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := store.LoadTasks()
				if err != nil {
					b.Fatalf("LoadTasks failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkCompleteTask measures task completion performance
func BenchmarkCompleteTask(b *testing.B) {
	store := createBenchStorage(b)

	// Create tasks
	tasks := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		task, _ := store.AddTask(fmt.Sprintf("Task %d", i), "", PriorityNone, nil)
		tasks[i] = task.ID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := store.CompleteTask(tasks[i])
		if err != nil {
			b.Fatalf("CompleteTask failed: %v", err)
		}
	}
}

// BenchmarkDeleteTask measures task deletion performance
func BenchmarkDeleteTask(b *testing.B) {
	store := createBenchStorage(b)

	// Create tasks
	tasks := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		task, _ := store.AddTask(fmt.Sprintf("Task %d", i), "", PriorityNone, nil)
		tasks[i] = task.ID
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := store.DeleteTask(tasks[i])
		if err != nil {
			b.Fatalf("DeleteTask failed: %v", err)
		}
	}
}

// BenchmarkSaveTasks measures task persistence performance
func BenchmarkSaveTasks(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			store := createBenchStorage(b)

			// Create task store
			taskStore := &TaskStore{Tasks: make([]Task, size)}
			for i := 0; i < size; i++ {
				taskStore.Tasks[i] = Task{
					ID:        fmt.Sprintf("t_%d", i),
					Text:      fmt.Sprintf("Task %d", i),
					Project:   "benchmark",
					Done:      i%2 == 0,
					CreatedAt: time.Now(),
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := store.SaveTasks(taskStore)
				if err != nil {
					b.Fatalf("SaveTasks failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkAddHabit measures habit creation performance
func BenchmarkAddHabit(b *testing.B) {
	store := createBenchStorage(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.AddHabit(fmt.Sprintf("Habit %d", i), "🎯")
		if err != nil {
			b.Fatalf("AddHabit failed: %v", err)
		}
	}
}

// BenchmarkLoadHabits measures habit loading performance
func BenchmarkLoadHabits(b *testing.B) {
	sizes := []int{10, 50, 100}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			store := createBenchStorage(b)

			// Populate with habits
			for i := 0; i < size; i++ {
				store.AddHabit(fmt.Sprintf("Habit %d", i), "🔥")
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := store.LoadHabits()
				if err != nil {
					b.Fatalf("LoadHabits failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkToggleHabitToday measures habit toggle performance
func BenchmarkToggleHabitToday(b *testing.B) {
	store := createBenchStorage(b)

	habit, _ := store.AddHabit("Benchmark Habit", "⚡")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.ToggleHabitToday(habit.ID)
		if err != nil {
			b.Fatalf("ToggleHabitToday failed: %v", err)
		}
	}
}

// BenchmarkGetHabitStreak measures streak calculation performance
func BenchmarkGetHabitStreak(b *testing.B) {
	store := createBenchStorage(b)

	habit, _ := store.AddHabit("Streak Habit", "🔥")

	// Create a 30-day streak
	hs, _ := store.LoadHabits()
	for i := 0; i < 30; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		hs.Logs = append(hs.Logs, HabitLog{
			HabitID: habit.ID,
			Date:    date,
		})
	}
	store.SaveHabits(hs)

	hs, _ = store.LoadHabits()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		streak := store.GetHabitStreak(hs, habit.ID)
		if streak == 0 {
			b.Fatal("streak should not be 0")
		}
	}
}

// BenchmarkGetHabitWeek measures week data retrieval performance
func BenchmarkGetHabitWeek(b *testing.B) {
	store := createBenchStorage(b)

	habit, _ := store.AddHabit("Week Habit", "📅")

	// Add logs for the past 7 days
	hs, _ := store.LoadHabits()
	for i := 0; i < 7; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		hs.Logs = append(hs.Logs, HabitLog{
			HabitID: habit.ID,
			Date:    date,
		})
	}
	store.SaveHabits(hs)

	hs, _ = store.LoadHabits()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		week := store.GetHabitWeek(hs, habit.ID)
		if len(week) != 7 {
			b.Fatal("week should have 7 days")
		}
	}
}

// BenchmarkLoadHabitsWithManyLogs tests performance with many logs
func BenchmarkLoadHabitsWithManyLogs(b *testing.B) {
	logCounts := []int{100, 1000, 10000}

	for _, count := range logCounts {
		b.Run(fmt.Sprintf("logs_%d", count), func(b *testing.B) {
			store := createBenchStorage(b)

			// Create habits
			habits := make([]string, 10)
			for i := 0; i < 10; i++ {
				h, _ := store.AddHabit(fmt.Sprintf("Habit %d", i), "🎯")
				habits[i] = h.ID
			}

			// Create many logs
			hs, _ := store.LoadHabits()
			for i := 0; i < count; i++ {
				habitID := habits[i%len(habits)]
				date := time.Now().AddDate(0, 0, -(i/len(habits))).Format("2006-01-02")
				hs.Logs = append(hs.Logs, HabitLog{
					HabitID: habitID,
					Date:    date,
				})
			}
			store.SaveHabits(hs)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := store.LoadHabits()
				if err != nil {
					b.Fatalf("LoadHabits failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkStartTimer measures timer start performance
func BenchmarkStartTimer(b *testing.B) {
	store := createBenchStorage(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := store.StartTimer(fmt.Sprintf("Project %d", i))
		if err != nil {
			b.Fatalf("StartTimer failed: %v", err)
		}
	}
}

// BenchmarkStopTimer measures timer stop performance
func BenchmarkStopTimer(b *testing.B) {
	store := createBenchStorage(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		store.StartTimer("Project")
		b.StartTimer()

		err := store.StopTimer()
		if err != nil {
			b.Fatalf("StopTimer failed: %v", err)
		}
	}
}

// BenchmarkGetTodayTotal measures today total calculation performance
func BenchmarkGetTodayTotal(b *testing.B) {
	entryCounts := []int{10, 100, 1000}

	for _, count := range entryCounts {
		b.Run(fmt.Sprintf("entries_%d", count), func(b *testing.B) {
			store := createBenchStorage(b)

			// Create many entries for today
			ts := &TimerStore{
				Entries: make([]TimerEntry, count),
			}

			now := time.Now()
			for i := 0; i < count; i++ {
				start := now.Add(-time.Duration(i+1) * time.Hour)
				end := start.Add(30 * time.Minute)
				ts.Entries[i] = TimerEntry{
					Project:   fmt.Sprintf("Project %d", i),
					StartedAt: start,
					EndedAt:   end,
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				total := store.GetTodayTotal(ts)
				if total == 0 {
					b.Fatal("total should not be 0")
				}
			}
		})
	}
}

// BenchmarkGetWeekTotal measures week total calculation performance
func BenchmarkGetWeekTotal(b *testing.B) {
	entryCounts := []int{10, 100, 1000}

	for _, count := range entryCounts {
		b.Run(fmt.Sprintf("entries_%d", count), func(b *testing.B) {
			store := createBenchStorage(b)

			// Create entries across the week
			ts := &TimerStore{
				Entries: make([]TimerEntry, count),
			}

			now := time.Now()
			for i := 0; i < count; i++ {
				start := now.Add(-time.Duration(i) * time.Hour)
				end := start.Add(30 * time.Minute)
				ts.Entries[i] = TimerEntry{
					Project:   fmt.Sprintf("Project %d", i),
					StartedAt: start,
					EndedAt:   end,
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				total := store.GetWeekTotal(ts)
				if total == 0 {
					b.Fatal("total should not be 0")
				}
			}
		})
	}
}

// BenchmarkTimerCalculations measures complex timer boundary calculations
func BenchmarkTimerCalculations(b *testing.B) {
	store := createBenchStorage(b)

	// Create a scenario with entries spanning multiple days
	ts := &TimerStore{
		Entries: make([]TimerEntry, 100),
	}

	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for i := 0; i < 100; i++ {
		// Some entries today, some yesterday, some spanning midnight
		offset := time.Duration(i) * time.Hour
		if i%3 == 0 {
			// Span midnight
			start := dayStart.Add(-2 * time.Hour)
			end := dayStart.Add(2 * time.Hour)
			ts.Entries[i] = TimerEntry{
				Project:   "Midnight",
				StartedAt: start,
				EndedAt:   end,
			}
		} else {
			start := dayStart.Add(-offset)
			end := start.Add(30 * time.Minute)
			ts.Entries[i] = TimerEntry{
				Project:   fmt.Sprintf("Project %d", i),
				StartedAt: start,
				EndedAt:   end,
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		todayTotal := store.GetTodayTotal(ts)
		weekTotal := store.GetWeekTotal(ts)

		if todayTotal > weekTotal {
			b.Fatal("today total cannot exceed week total")
		}
	}
}

// BenchmarkLoadTimer measures timer state loading performance
func BenchmarkLoadTimer(b *testing.B) {
	entryCounts := []int{10, 100, 1000}

	for _, count := range entryCounts {
		b.Run(fmt.Sprintf("entries_%d", count), func(b *testing.B) {
			store := createBenchStorage(b)

			// Create many timer entries
			for i := 0; i < count; i++ {
				store.StartTimer(fmt.Sprintf("Project %d", i))
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := store.LoadTimer()
				if err != nil {
					b.Fatalf("LoadTimer failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkConcurrentReads measures read performance under concurrent access
func BenchmarkConcurrentReads(b *testing.B) {
	store := createBenchStorage(b)

	// Populate with data
	for i := 0; i < 100; i++ {
		store.AddTask(fmt.Sprintf("Task %d", i), "project", PriorityNone, nil)
		store.AddHabit(fmt.Sprintf("Habit %d", i), "🎯")
	}
	store.StartTimer("Active Project")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Simulate concurrent reads from different panes
			_, _ = store.LoadTasks()
			_, _ = store.LoadHabits()
			_, _ = store.LoadTimer()
		}
	})
}

// BenchmarkSortTasks measures task sorting performance
func BenchmarkSortTasks(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			store := createBenchStorage(b)

			// Create tasks with mixed states
			taskStore := &TaskStore{Tasks: make([]Task, size)}
			now := time.Now()
			for i := 0; i < size; i++ {
				task := Task{
					ID:        fmt.Sprintf("t_%d", i),
					Text:      fmt.Sprintf("Task %d", i),
					Done:      i%3 == 0, // Every third task is done
					CreatedAt: now.Add(-time.Duration(i) * time.Minute),
				}
				if task.Done {
					completedAt := now.Add(-time.Duration(i) * time.Second)
					task.CompletedAt = &completedAt
				}
				// Add project to some tasks
				if i%4 == 0 {
					task.Project = "important"
				}
				taskStore.Tasks[i] = task
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				store.SortTasks(taskStore.Tasks)
			}
		})
	}
}

// createBenchStorage creates a storage instance for benchmarks
func createBenchStorage(b *testing.B) *Storage {
	b.Helper()
	store, err := New(b.TempDir())
	if err != nil {
		b.Fatalf("failed to create bench storage: %v", err)
	}
	return store
}
