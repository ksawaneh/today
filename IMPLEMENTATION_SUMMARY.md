# Feature Implementation Summary

This document summarizes the feature additions implemented for the "today" terminal productivity app.

## Completed Backend Work

### 1. Task Priorities and Due Dates ✅

**Files Modified:**
- `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/storage/models.go`
- `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/storage/storage.go`

**Changes:**
1. Added `Priority` type with constants: `PriorityHigh`, `PriorityMedium`, `PriorityLow`, `PriorityNone`
2. Added fields to `Task` struct:
   - `Priority Priority` - Task priority level
   - `DueDate *time.Time` - Optional due date

3. Updated `AddTask()` function signature:
   ```go
   func (s *Storage) AddTask(text, project string, priority Priority, dueDate *time.Time) (*Task, error)
   ```

4. Added `SortTasks()` function that sorts tasks by:
   - Completion status (pending first)
   - Priority (high → medium → low → none)
   - Due date (earliest first)
   - Creation time (newest first)

**Backward Compatibility:** All fields are optional with `omitempty` JSON tags. Existing task data will continue to work.

---

### 2. Habit Frequency Options ✅

**Files Modified:**
- `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/storage/models.go`

**Changes:**
1. Added `HabitFrequency` type with constants:
   - `FrequencyDaily` - Every day (default)
   - `FrequencyWeekly` - Once per week
   - `FrequencyWeekdays` - Monday through Friday
   - `FrequencyCustom` - Specific days of week

2. Added fields to `Habit` struct:
   - `Frequency HabitFrequency` - How often the habit should be done
   - `CustomDays []int` - For custom frequency (0=Sunday, 1=Monday, etc.)

**Backward Compatibility:** `Frequency` field is optional and defaults to daily behavior for existing habits.

---

### 3. Timer Analytics ✅

**Files Modified:**
- `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/storage/storage.go`

**New Types:**
```go
type ProjectTotal struct {
    Project string
    Total   time.Duration
}

type DayBreakdown struct {
    Date  string        // YYYY-MM-DD format
    Total time.Duration
}
```

**New Functions:**

1. `GetProjectTotals(store *TimerStore) []ProjectTotal`
   - Returns total time spent per project
   - Sorted by time spent (descending)
   - Includes currently running timer

2. `GetDailyBreakdown(store *TimerStore, days int) []DayBreakdown`
   - Returns time tracked per day for last N days
   - Handles entries spanning multiple days
   - Includes currently running timer

---

### 4. Data Export (JSON/CSV) ✅

**Files Modified:**
- `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/storage/storage.go`

**New Functions:**

**JSON Export:**
- `ExportTasksJSON() ([]byte, error)` - Export all tasks
- `ExportHabitsJSON() ([]byte, error)` - Export habits and logs
- `ExportTimerJSON() ([]byte, error)` - Export timer entries

**CSV Export:**
- `ExportTasksCSV() (string, error)` - Export tasks with headers:
  ```
  ID,Text,Project,Priority,DueDate,Done,CreatedAt,CompletedAt
  ```

- `ExportTimerCSV() (string, error)` - Export timer entries with headers:
  ```
  Project,StartedAt,EndedAt,Duration
  ```

**Features:**
- Proper CSV escaping for fields containing commas or quotes
- ISO 8601 date formatting
- Human-readable duration strings

---

## Partially Completed UI Work

### Files Modified:
- `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/ui/commands.go`
- `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/ui/tasks.go`

**Changes:**
1. Updated `addTaskCmd()` signature to accept priority and due date parameters
2. Added `time` import to commands.go
3. Updated task pane to call storage `SortTasks()` when loading tasks
4. Modified task creation to pass default priority (`PriorityNone`) and nil due date

---

## Remaining UI Work

### 1. Fix Build Errors

**Issue:** `habits.go` is missing `NewHabitsPaneWithKeys()` function that matches the pattern in `tasks.go` and `timer.go`.

**Required Changes:**
1. Add config import to habits.go
2. Add `NewHabitsPaneWithKeys()` function:
   ```go
   func NewHabitsPaneWithKeys(store *storage.Storage, styles *Styles, keyCfg *config.KeysConfig) *HabitsPane {
       // ... implementation matching NewHabitsPane but using keyCfg
   }
   ```

3. Update `NewHabitsPane()` to call `NewHabitKeyMap(keyCfg)` instead of `DefaultHabitKeyMap()`

---

### 2. Add UI Visual Indicators for Priority and Due Date

**File:** `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/ui/tasks.go`

**Required Changes:**

In the `View()` method, add visual indicators when rendering tasks:

```go
// Priority indicator (using color circles or symbols)
var priorityIcon string
switch task.Priority {
case storage.PriorityHigh:
    priorityIcon = "●" // Red/high priority indicator
case storage.PriorityMedium:
    priorityIcon = "●" // Yellow/medium priority
case storage.PriorityLow:
    priorityIcon = "○" // Blue/low priority
default:
    priorityIcon = " " // No priority
}

// Due date indicator
var dueDateStr string
if task.DueDate != nil {
    dueIn := task.DueDate.Sub(time.Now())
    if dueIn < 0 {
        dueDateStr = " (overdue!)"
    } else if dueIn < 24*time.Hour {
        dueDateStr = " (today)"
    } else if dueIn < 48*time.Hour {
        dueDateStr = " (tomorrow)"
    }
}

// Update the line formatting
line := fmt.Sprintf("%s %s %s%s", checkbox, priorityIcon, text, dueDateStr)
```

**Color Styling:**
Use the styles system to color-code priorities:
- High: ColorDanger (red)
- Medium: ColorWarning (yellow)
- Low: ColorAccent (blue)

---

### 3. Add Task Priority/Due Date Input UI

Currently, tasks are created with default priority. Add UI flow for setting:

**Option A - Simple:** Add 'p' key to cycle priority on selected task
```go
case key.Matches(msg, p.keys.SetPriority):
    // Cycle through priorities
    // Call UpdateTaskPriority() storage function (needs to be added)
```

**Option B - Advanced:** Multi-step input like habits:
1. Step 0: Enter task text
2. Step 1: Select priority (h/m/l/skip)
3. Step 2: Enter due date (format: YYYY-MM-DD or skip)

---

### 4. Update Habit UI and Logic for Frequency

**Files:**
- `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/ui/habits.go`
- `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/storage/storage.go`

**Storage Changes:**
Update streak and week calculation to respect frequency:

```go
func (s *Storage) ShouldShowHabitToday(habit *Habit) bool {
    today := time.Now()
    switch habit.Frequency {
    case FrequencyWeekdays:
        weekday := int(today.Weekday())
        return weekday >= 1 && weekday <= 5 // Mon-Fri
    case FrequencyCustom:
        weekday := int(today.Weekday())
        for _, day := range habit.CustomDays {
            if day == weekday {
                return true
            }
        }
        return false
    default: // Daily or Weekly
        return true
    }
}
```

**UI Changes:**
1. Show frequency badge next to habit name: `📅 Daily`, `🗓️ M-F`, `📆 Custom`
2. Gray out habits that aren't scheduled for today
3. Update streak calculation to skip non-scheduled days

---

### 5. Create Analytics View Component

**New File:** `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/ui/analytics.go`

**Structure:**
```go
type AnalyticsPane struct {
    timerStore *storage.TimerStore
    storage    *storage.Storage
    styles     *Styles
    width      int
    height     int
}

func (p *AnalyticsPane) View() string {
    // Show:
    // 1. Project totals (top 10)
    // 2. Daily breakdown (last 7 days as bar chart)
    // 3. Weekly total
}
```

**Integration:**
- Add `PaneAnalytics` to app.go
- Add 'v' key binding to toggle analytics view
- Make it a modal overlay like help

---

### 6. Add Export Menu

**New File:** `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/ui/export.go`

**Structure:**
```go
type ExportMenu struct {
    selected int
    options  []string // ["Tasks (JSON)", "Tasks (CSV)", "Habits (JSON)", "Timer (CSV)", "Cancel"]
}
```

**Flow:**
1. Press 'e' to show export menu overlay
2. Use j/k to navigate options
3. Enter to select and save file to `~/Downloads/today-export-YYYYMMDD-HHMMSS.{json,csv}`
4. Show success message with file path

---

### 7. Implement Undo Functionality

**Approach:** Operation log with reverse operations

**New File:** `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/undo/undo.go`

```go
type Operation struct {
    Type     string    // "task_complete", "task_delete", "habit_toggle", etc.
    Data     any       // Original state needed to reverse
    Timestamp time.Time
}

type UndoStack struct {
    operations []Operation
    maxSize    int // e.g., 20
}

func (u *UndoStack) Push(op Operation)
func (u *UndoStack) Pop() (Operation, error)
func (u *UndoStack) Undo(storage *storage.Storage) error
```

**Integration:**
- Add UndoStack to app.go
- Record operations before executing them
- Add Ctrl+Z key binding to pop and execute undo
- Show "Undid: [operation]" status message

---

## Testing Checklist

### Backward Compatibility Tests

1. **Existing Data:**
   - [ ] Old tasks.json loads correctly (no Priority or DueDate fields)
   - [ ] Old habits.json loads correctly (no Frequency fields)
   - [ ] Old timer.json loads correctly

2. **Migration:**
   - [ ] Tasks without priority sort to bottom (after prioritized tasks)
   - [ ] Habits without frequency behave as daily habits

### Feature Tests

3. **Task Priority:**
   - [ ] High priority tasks appear before medium
   - [ ] Medium before low
   - [ ] Low before unprioritized
   - [ ] Completed tasks always at bottom regardless of priority

4. **Task Due Date:**
   - [ ] Tasks due today show indicator
   - [ ] Overdue tasks show warning
   - [ ] Tasks sort by due date within same priority

5. **Habit Frequency:**
   - [ ] Weekday habits don't show on weekends
   - [ ] Custom habits only show on specified days
   - [ ] Streak calculation respects frequency

6. **Analytics:**
   - [ ] Project totals sum correctly
   - [ ] Daily breakdown handles entries spanning midnight
   - [ ] Currently running timer included in calculations

7. **Export:**
   - [ ] JSON exports are valid and parseable
   - [ ] CSV exports open correctly in spreadsheet apps
   - [ ] Special characters (commas, quotes) escaped properly

8. **Undo:**
   - [ ] Can undo task completion
   - [ ] Can undo task deletion
   - [ ] Can undo habit toggle
   - [ ] Undo stack limited to last N operations
   - [ ] Can't undo if stack is empty

---

## Build Instructions

After completing the remaining UI work:

```bash
# Fix any remaining build errors
go build ./...

# Run tests
go test ./...

# Build the binary
go build -o today cmd/today/main.go

# Test run
./today
```

## Known Issues

1. **Build Error:** `NewHabitsPaneWithKeys` not found - needs implementation
2. **Missing:** Priority/due date input UI - currently uses defaults
3. **Missing:** Visual indicators for priority in task list (prepared but not styled)
4. **Missing:** Analytics view component
5. **Missing:** Export menu UI
6. **Missing:** Undo infrastructure

## Data Format Examples

### Task with Priority and Due Date
```json
{
  "id": "t_1702834567890_abc123",
  "text": "Review pull request",
  "project": "work",
  "priority": "high",
  "due_date": "2025-12-15T17:00:00Z",
  "done": false,
  "created_at": "2025-12-14T10:00:00Z"
}
```

### Habit with Custom Frequency
```json
{
  "id": "h_1702834567890_def456",
  "name": "Gym",
  "icon": "💪",
  "frequency": "custom",
  "custom_days": [1, 3, 5],
  "created_at": "2025-12-01T08:00:00Z"
}
```

### CSV Export Sample
```csv
ID,Text,Project,Priority,DueDate,Done,CreatedAt,CompletedAt
t_123,Fix bug,work,high,2025-12-15,false,2025-12-14 10:00:00,
t_124,"Review ""new"" feature",work,medium,2025-12-16,true,2025-12-13 09:00:00,2025-12-14 11:30:00
```

---

## Architecture Notes

The implementation follows the project's established patterns:

1. **Separation of Concerns:**
   - Models in `storage/models.go`
   - Business logic in `storage/storage.go`
   - UI in separate pane files

2. **Backward Compatibility:**
   - All new fields are optional (`omitempty`)
   - Default values preserve old behavior
   - No breaking changes to existing APIs

3. **Type Safety:**
   - Custom types for Priority and Frequency
   - Validation in storage layer
   - Errors bubble up to UI for user feedback

4. **Testing:**
   - Storage functions are pure and testable
   - Golden file tests can verify UI output
   - Table-driven tests for sorting logic

---

*Last Updated: 2025-12-14*
