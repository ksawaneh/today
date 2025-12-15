# Quick Fix Guide

To get the code building again after the feature additions, apply these fixes:

## Fix 1: Add Missing Function to habits.go

Add this function after the existing `NewHabitsPane()` function in `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/ui/habits.go`:

```go
// NewHabitsPaneWithKeys creates a new habits pane with custom key bindings.
func NewHabitsPaneWithKeys(store *storage.Storage, styles *Styles, keyCfg *config.KeysConfig) *HabitsPane {
	ti := textinput.New()
	ti.Placeholder = "Habit name (e.g., Exercise)"
	ti.CharLimit = 30
	ti.Width = 30

	return &HabitsPane{
		habitStore: &storage.HabitStore{},
		cursor:     0,
		focused:    false,
		input:      ti,
		storage:    store,
		styles:     styles,
		keys:       NewHabitKeyMap(keyCfg),
		inputKeys:  NewInputKeyMap(keyCfg),
	}
}
```

Also update the NewHabitsPane function to use the WithKeys variant:

```go
// NewHabitsPane creates a new habits pane.
func NewHabitsPane(store *storage.Storage, styles *Styles) *HabitsPane {
	return NewHabitsPaneWithKeys(store, styles, &config.KeysConfig{})
}
```

And add the config import:

```go
import (
	"fmt"
	"strings"
	"time"

	"today/internal/config"
	"today/internal/storage"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)
```

## Fix 2: Verify keys.go has NewHabitKeyMap

Check that `/Users/karimsawaneh/kodehut/dev/kodehut/personal/today/internal/ui/keys.go` has:

```go
func NewHabitKeyMap(cfg *config.KeysConfig) HabitKeyMap {
	// Implementation should be similar to NewTaskKeyMap
}
```

If it doesn't exist, add it following the pattern of `NewTaskKeyMap`.

## Fix 3: Update View() methods

The View() methods in tasks.go, timer.go, and habits.go are using old style references. They should use `p.styles.XYZ` instead of bare `XYZ`.

Example - in tasks.go around line 199-210, change:
```go
// Before
title := PaneTitleStyle.Render("✅ TASKS")

// After
title := p.styles.PaneTitleStyle.Render("✅ TASKS")
```

Do this for all style references in View() methods.

## Verification

After applying fixes:

```bash
# Should compile without errors
go build ./...

# Run tests
go test ./internal/storage/...

# Try running the app
go run cmd/today/main.go
```

## If You Get "undefined: NewTaskKeyMap" etc.

The keys.go file needs these constructor functions. Add them if missing:

```go
func NewTaskKeyMap(cfg *config.KeysConfig) TaskKeyMap {
	base := DefaultTaskKeyMap()
	// Apply config overrides if needed
	return base
}

func NewTimerKeyMap(cfg *config.KeysConfig) TimerKeyMap {
	base := DefaultTimerKeyMap()
	return base
}

func NewHabitKeyMap(cfg *config.KeysConfig) HabitKeyMap {
	base := DefaultHabitKeyMap()
	return base
}

func NewInputKeyMap(cfg *config.KeysConfig) InputKeyMap {
	base := DefaultInputKeyMap()
	return base
}
```

---

## Next Steps After Build Fixes

Once the code builds successfully:

1. **Test backward compatibility** with existing data files
2. **Add visual indicators** for task priority (see IMPLEMENTATION_SUMMARY.md)
3. **Implement analytics view** for timer reports
4. **Add export menu UI**
5. **Implement undo functionality**

See `IMPLEMENTATION_SUMMARY.md` for complete details on each remaining feature.
