# Building a World-Class Terminal Productivity Dashboard

A comprehensive guide to architecture, UX, data layer, performance, and distribution patterns for exceptional Go TUI applications.

---

## Executive Summary

Your Go-based "today" app is entering a well-defined space where the difference between "good" and "exceptional" comes down to specific, implementable patterns. After analyzing successful TUI tools like lazygit (**56k+ GitHub stars**), k9s, and the Charm ecosystem, along with productivity leaders like Things 3 and Todoist, this document provides concrete guidance across architecture, UX, data, performance, distribution, and product differentiation.

**The core insight:** Terminal productivity tools fail not from lack of features, but from friction. Users abandon taskwarrior despite its power because configuration is mandatory; they abandon todo.txt because it lacks visual feedback. The exceptional TUI app works immediately with sensible defaults, feels responsive and beautiful, and provides unique value that GUI apps cannot—deep integration with the developer workflow.

---

## 1. Architecture Patterns from Production TUI Applications

The best Bubble Tea applications follow a consistent architectural pattern that separates concerns while maintaining the Elm architecture's benefits. Lazygit—the most successful Go TUI application—uses a layered structure that scales well.

### Recommended Directory Structure

```
today/
├── main.go                   # Entry point, program setup
├── internal/
│   ├── app/
│   │   └── app.go           # Application lifecycle, initialization
│   ├── tui/
│   │   ├── model.go         # Root model, state management
│   │   ├── update.go        # Update logic, message routing
│   │   ├── view.go          # View composition
│   │   ├── keys.go          # KeyMap definitions using key.Binding
│   │   ├── styles.go        # Centralized Lipgloss styles
│   │   ├── messages.go      # Custom message types
│   │   └── components/      # Reusable Bubble Tea components
│   │       ├── tasklist/
│   │       ├── timer/
│   │       └── habits/
│   ├── commands/            # tea.Cmd factories for I/O
│   └── storage/             # Persistence layer
├── migrations/              # Embedded SQL migrations
└── testdata/               # Golden files for teatest
```

### Critical Principle: Keep the Event Loop Fast

The `Update()` and `View()` methods block Bubble Tea's event loop, so expensive operations must be offloaded to `tea.Cmd`:

```go
// BAD - blocks event loop
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    data := db.LoadTasks()  // Blocks UI!
    return m, nil
}

// GOOD - offload to command
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m, func() tea.Msg {
        data, err := db.LoadTasks()
        if err != nil { return errMsg{err} }
        return tasksLoadedMsg{data}
    }
}
```

### Component Composition

Follow a tree-of-models pattern where the root model routes messages to child components:

```go
type Model struct {
    taskList    tasklist.Model
    timer       timer.Model
    habits      habits.Model
    activeView  ViewType
    common      *Common  // Shared deps: config, storage, logger
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    
    // Global messages first
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.taskList.SetSize(msg.Width, msg.Height-4)
    case tea.KeyMsg:
        if msg.String() == "?" { m.showHelp = !m.showHelp }
    }
    
    // Route to active view
    switch m.activeView {
    case TaskView:
        m.taskList, cmd = m.taskList.Update(msg)
        cmds = append(cmds, cmd)
    case TimerView:
        m.timer, cmd = m.timer.Update(msg)
        cmds = append(cmds, cmd)
    }
    return m, tea.Batch(cmds...)
}
```

The **Common struct pattern** (used by lazygit) eliminates prop drilling—embed a pointer to shared dependencies in every component so configuration changes propagate automatically.

### Testing Strategies

Use `teatest` for golden file testing of view output:

```go
func TestTaskListView(t *testing.T) {
    lipgloss.SetColorProfile(termenv.Ascii)  // Consistent colors
    
    tm := teatest.NewTestModel(t, initialModel(),
        teatest.WithInitialTermSize(80, 24))
    
    out, _ := io.ReadAll(tm.FinalOutput(t))
    teatest.RequireEqualOutput(t, out)  // Compares to .golden file
}
```

Run with `go test -update` to regenerate golden files after intentional changes.

---

## 2. UX Principles That Make TUI Tools Feel Exceptional

Lazygit's creator Jesse Duffield identified the core truth: **"Terminal users expect TUIs to be fast because they value speed more than other people."** Every interaction must feel instant.

### Keyboard Navigation That Builds on Muscle Memory

Implement vim-style bindings using Bubble Tea's key package:

```go
type KeyMap struct {
    Up     key.Binding
    Down   key.Binding
    Select key.Binding
    Quit   key.Binding
}

var DefaultKeyMap = KeyMap{
    Up:     key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("↑/k", "up")),
    Down:   key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("↓/j", "down")),
    Select: key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter", "select")),
    Quit:   key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}
```

**Standard TUI conventions to follow:**

- `h/j/k/l` or arrows for navigation
- `q` or `Esc` always goes back/quits
- `/` for search/filter
- `?` for contextual help
- `Tab` cycles panels; `1-5` jumps to numbered panels
- `g g` goes to top; `G` goes to bottom

### Animation and Theming

Bubble Tea's framerate-based renderer (default ~60fps) enables smooth animations. Use Charmbracelet's **harmonica** library for physics-based spring animations:

```go
// Critically-damped spring (ratio = 1.0) reaches target fastest without bounce
spring := harmonica.NewSpring(harmonica.FPS(60), 6.0, 1.0)
pos, velocity = spring.Update(pos, velocity, targetPos)
```

For **theming with graceful degradation**, use Lipgloss's adaptive colors:

```go
type Theme struct {
    Primary   lipgloss.AdaptiveColor
    Success   lipgloss.AdaptiveColor
    Error     lipgloss.AdaptiveColor
}

var DefaultTheme = Theme{
    Primary: lipgloss.AdaptiveColor{Light: "63", Dark: "212"},
    Success: lipgloss.AdaptiveColor{Light: "34", Dark: "78"},
    Error:   lipgloss.AdaptiveColor{Light: "160", Dark: "196"},
}

// Complete fallback chain for all terminal types
style := lipgloss.NewStyle().Foreground(lipgloss.CompleteColor{
    TrueColor: "#0000FF",
    ANSI256:   "86",
    ANSI:      "5",
})
```

**Always respect `NO_COLOR`** environment variable—check on startup and disable all colors if present.

### Accessibility Checklist

Screen readers struggle with TUI's constant redraws. Provide an `--accessible` flag that:

- Replaces animated spinners with static "Working..." text
- Uses clear text indicators instead of visual-only progress
- Exports long outputs to accessible formats (CSV, HTML with headings)

---

## 3. Data Architecture for Reliability and Trust

After analyzing taskwarrior's evolution (which moved from JSON to SQLite in v3 due to "slow and buggy" performance with thousands of tasks), **SQLite is the right choice** for your productivity app.

### Recommended Data Layer Setup

```go
import (
    "database/sql"
    _ "modernc.org/sqlite"  // Pure Go, no CGO = easy cross-compilation
    "github.com/adrg/xdg"
    "github.com/gofrs/flock"
    "github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func NewApp() (*App, error) {
    // XDG-compliant paths
    dataDir, _ := xdg.DataFile("today/")
    dbPath := filepath.Join(dataDir, "today.db")
    
    // Single-instance enforcement
    fileLock := flock.New(dbPath + ".lock")
    locked, _ := fileLock.TryLock()
    if !locked {
        return nil, fmt.Errorf("another instance is running")
    }
    
    db, _ := sql.Open("sqlite", dbPath)
    
    // Optimal SQLite settings
    db.Exec("PRAGMA journal_mode=WAL")      // Better concurrency
    db.Exec("PRAGMA synchronous=NORMAL")    // Speed vs safety balance
    db.Exec("PRAGMA foreign_keys=ON")
    db.Exec("PRAGMA busy_timeout=5000")
    
    // Run embedded migrations
    goose.SetBaseFS(migrationsFS)
    goose.SetDialect("sqlite3")
    goose.Up(db, "migrations")
    
    return &App{db: db, lock: fileLock}, nil
}
```

### Automatic Backups with Rotation

```go
func (bm *BackupManager) CreateBackup() error {
    timestamp := time.Now().Format("2006-01-02-150405")
    backupPath := filepath.Join(bm.backupDir, fmt.Sprintf("backup-%s.db", timestamp))
    
    // SQLite's VACUUM INTO creates consistent backup
    _, err := bm.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupPath))
    if err != nil { return err }
    
    bm.rotateOldBackups(maxBackups)  // Keep last 7
    return nil
}
```

For **sync**, follow the `pass` password-store pattern: store data files in a git repository, with automatic commits on changes and simple push/pull commands.

---

## 4. Performance Optimization for Responsive Feel

### The Three Critical Performance Rules

**1. Use strings.Builder for view rendering** (80-250x faster than concatenation):

```go
func (m model) View() string {
    var b strings.Builder
    b.Grow(m.estimatedSize)  // Pre-allocate if known
    b.WriteString(m.header)
    b.WriteString(m.timerView())
    b.WriteString(m.taskListView())
    return b.String()
}
```

**2. Debounce search input** to avoid filtering on every keystroke:

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        m.searchInput += msg.String()
        return m, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
            return debounceMsg{}
        })
    case debounceMsg:
        m.filteredItems = filterItems(m.allItems, m.searchInput)
    }
}
```

**3. Virtual scrolling for large lists** (1000+ items):

```go
func (v *VirtualList) visibleRange() (start, end int) {
    start = max(0, v.scrollTop/v.itemHeight - v.buffer)
    visibleCount := v.viewHeight/v.itemHeight + 2*v.buffer
    end = min(len(v.items), start+visibleCount)
    return
}
```

Use Bubbles' built-in `list` component for lists under 500 items—it handles pagination and fuzzy filtering efficiently.

### Startup Time Optimization

Target **<100ms startup**. Key patterns:

- Use `sync.OnceValue` (Go 1.21+) for lazy-loaded resources
- Avoid large map literals in source files (load from embedded JSON instead)
- Initialize database and load assets in parallel with `sync.WaitGroup`

---

## 5. Distribution Strategy for Maximum Adoption

The most successful CLI tools share common launch patterns. GoReleaser automates cross-platform builds and package manager integration.

### Essential GoReleaser Configuration

```yaml
# .goreleaser.yaml
version: 2
builds:
  - env: [CGO_ENABLED=0]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags: [-s -w -X main.version={{.Version}}]

brews:
  - repository:
      owner: your-username
      name: homebrew-tap
    description: "Terminal productivity dashboard"

nfpms:
  - formats: [deb, rpm, apk]
```

### Packaging Checklist

| Platform | Action |
|----------|--------|
| **Homebrew** | Create `homebrew-tap` repo, GoReleaser auto-generates formula |
| **AUR** | Create PKGBUILD, run `makepkg --printsrcinfo > .SRCINFO`, push to AUR |
| **Scoop** | Create bucket repo with JSON manifest |
| **Nix** | Include `flake.nix` in repository root |

### README Structure That Converts

```markdown
# today

> Your productivity dashboard in the terminal

![Demo GIF - 5-10 seconds showing core workflow]

## Features
- Task management with natural language input
- Pomodoro timer with desktop notifications  
- Habit tracking with streak visualization

## Installation
brew install your-username/tap/today

## Quick Start
today              # Open dashboard
today add "Review PR by tomorrow 3pm"
today timer start  # Start focus session
```

The **demo GIF is critical**—use VHS (Charmbracelet's scripted terminal recorder) to create consistent, high-quality recordings.

### Hacker News Launch Strategy

- Title format: `Show HN: Today – Terminal productivity dashboard for developers`
- First comment: Introduce yourself, explain the problem, share technical decisions
- Be available to respond for 4-6 hours post-launch
- Link to GitHub repo (signals working product, easy to try)

---

## 6. Product Differentiation That Matters

The gap in terminal productivity tools is clear: taskwarrior is powerful but requires extensive configuration; todo.txt is simple but lacks features; neither integrates with developer workflows. GUI apps like Things 3 and Todoist offer beautiful UX but force context-switching away from the terminal.

### Features That Would Make "today" Exceptional

**1. Start date + due date** (Things 3's killer feature): Separate "when to work on it" from "when it's due" to prevent today-list overwhelm.

**2. Natural language input**: Parse "tomorrow 3pm review PR #123 high priority +work" into structured data. This is Todoist's most-praised feature.

**3. Unified dashboard**: Single TUI showing today's tasks, active timer, habit streaks—no more switching between taskwarrior, timewarrior, and habit scripts.

**4. Developer-native integrations**:

- Auto-detect project from current directory
- Parse git branch for context
- Auto-log completed pomodoros with task association
- Generate standup reports from yesterday's activity

**5. Beautiful defaults that work immediately**: Zero required configuration, sensible keybindings, adaptive theming.

### Competitive Positioning

| Feature | Taskwarrior | todo.txt | Todoist | **today** |
|---------|-------------|----------|---------|-----------|
| Start + Due dates | ❌ | ❌ | ❌ | ✅ |
| Habit tracking | ❌ | ❌ | ❌ | ✅ |
| Pomodoro timer | Via plugin | ❌ | ❌ | ✅ |
| NLP input | ❌ | ❌ | ✅ | ✅ |
| Local-first | ✅ | ✅ | ❌ | ✅ |
| Beautiful TUI | ❌ | ❌ | N/A | ✅ |
| Zero config | ❌ | ✅ | ✅ | ✅ |
| Git integration | ❌ | ❌ | ❌ | ✅ |

### Streak and Progress Visualization

```
Habits This Week:
  Reading    ████████░░ 8/10 days 🔥 Current: 8 | Best: 23
  Exercise   ██████░░░░ 6/10 days
  Meditation ██████████ 10/10 ✓

Focus Today: ████████████░░ 4h 23m / 6h goal
```

---

## 7. What I Would Do Differently

Based on this research, here are the specific changes I would make to the current `today` implementation:

### Architecture Changes

| Current | Recommended | Why |
|---------|-------------|-----|
| JSON file storage | SQLite with WAL mode | Performance at scale, atomic operations, migrations |
| Single `storage.go` | Repository pattern per entity | Testability, separation of concerns |
| Inline styles | Centralized theme system | Easy theming, accessibility support |
| Direct file I/O in Update | tea.Cmd for all I/O | Non-blocking event loop |

### Data Model Changes

| Current | Recommended | Why |
|---------|-------------|-----|
| Simple task text | Task with start date, due date, project, tags | Things 3-style deferred tasks |
| Basic habit toggle | Flexible frequency (daily, weekdays, 3x/week) | Real habit patterns |
| Timer entries only | Pomodoro mode with focus sessions | Structured work intervals |

### UX Changes

| Current | Recommended | Why |
|---------|-------------|-----|
| Fixed color scheme | Adaptive colors with NO_COLOR support | Accessibility, terminal compatibility |
| Basic help overlay | Contextual help + searchable command palette | Discoverability |
| Manual text input | Natural language parsing | Faster capture |
| Exit message | Exit with progress summary + tip | Encouragement loop |

### Code Quality Changes

| Current | Recommended | Why |
|---------|-------------|-----|
| No tests | Golden file tests with teatest | Regression prevention |
| No error handling in UI | Error boundary pattern with user-friendly messages | Graceful degradation |
| Hardcoded keybindings | KeyMap with user override | Customization |

### Distribution Changes

| Current | Recommended | Why |
|---------|-------------|-----|
| Manual build | GoReleaser + GitHub Actions | One-command releases |
| README only | README + demo GIF + man page | Discoverability |
| No versioning | SemVer with embedded version string | User trust |

---

## 8. Implementation Priority

If rebuilding from scratch with limited time:

### Week 1: Foundation
- [ ] SQLite storage with migrations
- [ ] Proper tea.Cmd pattern for all I/O
- [ ] KeyMap-based keybindings
- [ ] Adaptive color theme

### Week 2: Core Features
- [ ] Tasks with start/due dates and projects
- [ ] Pomodoro timer mode
- [ ] Flexible habit frequencies
- [ ] Natural language task input

### Week 3: Polish
- [ ] Golden file tests for all views
- [ ] GoReleaser configuration
- [ ] Demo GIF with VHS
- [ ] Homebrew formula

### Week 4: Launch
- [ ] README optimization
- [ ] HN post draft
- [ ] r/commandline post
- [ ] Product Hunt preparation

---

## Conclusion

The path from "good" to "exceptional" TUI productivity app requires:

1. **Architectural discipline**: Bubble Tea's Elm architecture with fast Update/View, offloading I/O to commands
2. **Obsessive UX polish**: Sub-100ms response times, vim-style navigation, adaptive theming
3. **Data reliability**: SQLite with WAL mode, automatic backups, embedded migrations
4. **Zero-friction adoption**: Works immediately, sensible defaults, cross-platform packaging
5. **Unique developer value**: Git integration, NLP input, unified dashboard

The terminal productivity space has an opening for an app that combines the power of taskwarrior with the beauty of modern TUIs and the workflow integration that only terminal-native tools can provide. Your "today" app is positioned to fill that gap—execution on these patterns will determine whether it becomes the tool developers recommend to each other.

---

*Generated December 2025*
