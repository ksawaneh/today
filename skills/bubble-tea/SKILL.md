# Bubble Tea TUI Development Skill

A comprehensive guide for building terminal user interfaces with Charmbracelet's Bubble Tea framework.

## When to Use This Skill

- Building terminal user interfaces (TUIs)
- Creating interactive CLI applications
- Developing dashboard-style terminal apps
- Any project using Bubble Tea, Bubbles, or Lipgloss

---

## Core Architecture: The Elm Pattern

Bubble Tea implements the Elm architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                         Program                              │
│  ┌─────────┐    ┌──────────┐    ┌──────────┐    ┌────────┐ │
│  │  Init   │───▶│  Model   │───▶│   View   │───▶│ Screen │ │
│  └─────────┘    └────┬─────┘    └──────────┘    └────────┘ │
│                      │                                      │
│                      ▼                                      │
│                 ┌──────────┐                                │
│                 │  Update  │◀─── Messages (keyboard, etc)   │
│                 └──────────┘                                │
└─────────────────────────────────────────────────────────────┘
```

### The Three Functions

```go
// Model holds all application state
type Model struct {
    items    []string
    cursor   int
    selected map[int]struct{}
}

// Init returns the initial command (or nil)
func (m Model) Init() tea.Cmd {
    return nil  // Or tea.Batch(loadData, startTimer)
}

// Update handles messages and returns updated model + commands
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "up", "k":
            if m.cursor > 0 {
                m.cursor--
            }
        case "down", "j":
            if m.cursor < len(m.items)-1 {
                m.cursor++
            }
        case "enter", " ":
            if _, ok := m.selected[m.cursor]; ok {
                delete(m.selected, m.cursor)
            } else {
                m.selected[m.cursor] = struct{}{}
            }
        }
    }
    return m, nil
}

// View renders the UI as a string
func (m Model) View() string {
    var b strings.Builder
    
    for i, item := range m.items {
        cursor := " "
        if m.cursor == i {
            cursor = ">"
        }
        
        checked := " "
        if _, ok := m.selected[i]; ok {
            checked = "x"
        }
        
        b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, checked, item))
    }
    
    b.WriteString("\nPress q to quit.\n")
    return b.String()
}
```

---

## Critical Rules

### 1. Never Block in Update or View

```go
// BAD - Blocks the event loop
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    data := fetchFromAPI()  // BLOCKS!
    m.data = data
    return m, nil
}

// GOOD - Use tea.Cmd for I/O
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m, func() tea.Msg {
        data := fetchFromAPI()  // Runs in background
        return dataLoadedMsg{data}
    }
}
```

### 2. Use strings.Builder in View

```go
// BAD - String concatenation is slow
func (m Model) View() string {
    s := ""
    for _, item := range m.items {
        s += item + "\n"  // Creates new string each time
    }
    return s
}

// GOOD - Use strings.Builder (80-250x faster)
func (m Model) View() string {
    var b strings.Builder
    b.Grow(len(m.items) * 50)  // Pre-allocate if known
    for _, item := range m.items {
        b.WriteString(item)
        b.WriteString("\n")
    }
    return b.String()
}
```

### 3. Handle Window Size

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // Resize components
        m.list.SetSize(msg.Width, msg.Height-4)
    }
    return m, nil
}
```

---

## Component Patterns

### Parent-Child Composition

```go
type Model struct {
    tabs     tabs.Model      // Child component
    list     list.Model      // Child component
    input    textinput.Model // Child component
    focused  string          // Which child has focus
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd
    var cmd tea.Cmd
    
    // Global key handling
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            m.focused = nextFocus(m.focused)
            return m, nil
        case "ctrl+c":
            return m, tea.Quit
        }
    }
    
    // Route to focused component
    switch m.focused {
    case "list":
        m.list, cmd = m.list.Update(msg)
        cmds = append(cmds, cmd)
    case "input":
        m.input, cmd = m.input.Update(msg)
        cmds = append(cmds, cmd)
    }
    
    return m, tea.Batch(cmds...)
}
```

### Shared Dependencies (Common Pattern)

```go
// Common holds shared dependencies - pass pointer to all components
type Common struct {
    Config  *Config
    Storage *Storage
    Logger  *slog.Logger
    Width   int
    Height  int
}

type TaskList struct {
    common *Common  // Pointer to shared state
    tasks  []Task
    cursor int
}

func NewTaskList(common *Common) *TaskList {
    return &TaskList{common: common}
}
```

---

## KeyMap Pattern

Use `key.Binding` for discoverable, customizable keybindings:

```go
import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
    Up     key.Binding
    Down   key.Binding
    Select key.Binding
    Delete key.Binding
    Help   key.Binding
    Quit   key.Binding
}

var DefaultKeyMap = KeyMap{
    Up: key.NewBinding(
        key.WithKeys("up", "k"),
        key.WithHelp("↑/k", "move up"),
    ),
    Down: key.NewBinding(
        key.WithKeys("down", "j"),
        key.WithHelp("↓/j", "move down"),
    ),
    Select: key.NewBinding(
        key.WithKeys("enter", " "),
        key.WithHelp("enter/space", "select"),
    ),
    Delete: key.NewBinding(
        key.WithKeys("x", "d"),
        key.WithHelp("x/d", "delete"),
    ),
    Help: key.NewBinding(
        key.WithKeys("?"),
        key.WithHelp("?", "toggle help"),
    ),
    Quit: key.NewBinding(
        key.WithKeys("q", "ctrl+c"),
        key.WithHelp("q", "quit"),
    ),
}

// In Update:
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Up):
            m.cursor--
        case key.Matches(msg, m.keys.Down):
            m.cursor++
        case key.Matches(msg, m.keys.Quit):
            return m, tea.Quit
        }
    }
    return m, nil
}
```

---

## Lipgloss Styling

### Basic Styles

```go
import "github.com/charmbracelet/lipgloss"

var (
    // Colors
    purple    = lipgloss.Color("99")
    gray      = lipgloss.Color("245")
    darkGray  = lipgloss.Color("236")
    
    // Styles
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(purple).
        MarginLeft(2)
    
    itemStyle = lipgloss.NewStyle().
        PaddingLeft(4)
    
    selectedStyle = lipgloss.NewStyle().
        PaddingLeft(2).
        Foreground(purple).
        Bold(true)
    
    paneStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(gray).
        Padding(1, 2)
)
```

### Adaptive Colors (Light/Dark Mode)

```go
var (
    primary = lipgloss.AdaptiveColor{Light: "235", Dark: "252"}
    subtle  = lipgloss.AdaptiveColor{Light: "250", Dark: "238"}
)

// Complete color fallback for all terminal types
style := lipgloss.NewStyle().Foreground(lipgloss.CompleteColor{
    TrueColor: "#7C3AED",  // 24-bit color
    ANSI256:   "99",       // 256-color fallback
    ANSI:      "5",        // 16-color fallback
})
```

### Layout

```go
// Horizontal join
row := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

// Vertical join
column := lipgloss.JoinVertical(lipgloss.Left, header, content, footer)

// Place content (centering)
centered := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)

// Fixed width/height
box := style.Width(40).Height(10).Render(content)
```

---

## Common Bubbles Components

### List

```go
import "github.com/charmbracelet/bubbles/list"

// Item interface
type item struct {
    title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// Initialize
items := []list.Item{
    item{title: "Task 1", desc: "Description"},
    item{title: "Task 2", desc: "Description"},
}
l := list.New(items, list.NewDefaultDelegate(), 0, 0)
l.Title = "My Tasks"
```

### Text Input

```go
import "github.com/charmbracelet/bubbles/textinput"

ti := textinput.New()
ti.Placeholder = "Enter task..."
ti.Focus()
ti.CharLimit = 156
ti.Width = 40

// In Update
ti, cmd = ti.Update(msg)

// Get value
value := ti.Value()
```

### Viewport (Scrollable Content)

```go
import "github.com/charmbracelet/bubbles/viewport"

vp := viewport.New(80, 24)
vp.SetContent(longContent)

// In Update
vp, cmd = vp.Update(msg)

// In View
vp.View()
```

### Spinner

```go
import "github.com/charmbracelet/bubbles/spinner"

s := spinner.New()
s.Spinner = spinner.Dot  // or spinner.Line, spinner.MiniDot, etc.
s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

// Must tick the spinner
func (m Model) Init() tea.Cmd {
    return m.spinner.Tick
}
```

---

## Async Commands

### Loading Data

```go
type dataLoadedMsg struct {
    data []Item
    err  error
}

func loadDataCmd() tea.Msg {
    data, err := fetchData()
    return dataLoadedMsg{data, err}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case dataLoadedMsg:
        if msg.err != nil {
            m.err = msg.err
            return m, nil
        }
        m.data = msg.data
        m.loading = false
    }
    return m, nil
}
```

### Ticking (Timers, Clocks)

```go
type tickMsg time.Time

func tickCmd() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tickMsg:
        m.elapsed++
        return m, tickCmd()  // Keep ticking
    }
    return m, nil
}
```

### Debouncing Input

```go
type debounceMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        m.input += msg.String()
        m.debounceTimer++
        return m, tea.Tick(300*time.Millisecond, func(t time.Time) tea.Msg {
            return debounceMsg{}
        })
    case debounceMsg:
        // Only filter after user stops typing
        m.filtered = filter(m.items, m.input)
    }
    return m, nil
}
```

---

## Testing with teatest

```go
import (
    "testing"
    "github.com/charmbracelet/x/exp/teatest"
    "github.com/muesli/termenv"
)

func TestView(t *testing.T) {
    // Ensure consistent colors across environments
    lipgloss.SetColorProfile(termenv.Ascii)
    
    m := NewModel()
    tm := teatest.NewTestModel(t, m,
        teatest.WithInitialTermSize(80, 24),
    )
    
    // Simulate key presses
    tm.Send(tea.KeyMsg{Type: tea.KeyDown})
    tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
    
    // Get final output
    out := tm.FinalModel(t).(Model)
    
    // Assert state
    if out.cursor != 1 {
        t.Errorf("cursor = %d, want 1", out.cursor)
    }
}

// Golden file test
func TestViewGolden(t *testing.T) {
    lipgloss.SetColorProfile(termenv.Ascii)
    
    tm := teatest.NewTestModel(t, NewModel())
    
    out, _ := io.ReadAll(tm.FinalOutput(t))
    teatest.RequireEqualOutput(t, out)  // Compares to testdata/TestViewGolden.golden
}
```

Run `go test -update` to regenerate golden files.

---

## Performance Tips

### 1. Virtual Scrolling for Large Lists

```go
func (m Model) visibleItems() []Item {
    start := m.scrollOffset
    end := min(start+m.visibleCount, len(m.items))
    return m.items[start:end]
}
```

### 2. Memoize Expensive Renders

```go
type Model struct {
    cachedView    string
    cacheInvalid  bool
}

func (m Model) View() string {
    if !m.cacheInvalid && m.cachedView != "" {
        return m.cachedView
    }
    // Expensive render
    m.cachedView = expensiveRender()
    m.cacheInvalid = false
    return m.cachedView
}
```

### 3. Batch State Updates

```go
// BAD - Multiple re-renders
m.cursor++
m.selected = append(m.selected, m.cursor)
m.dirty = true

// GOOD - Single state update
m = Model{
    cursor:   m.cursor + 1,
    selected: append(m.selected, m.cursor),
    dirty:    true,
}
```

---

## Program Options

```go
func main() {
    p := tea.NewProgram(
        NewModel(),
        tea.WithAltScreen(),        // Use alternate screen buffer
        tea.WithMouseCellMotion(),  // Enable mouse support
        tea.WithOutput(os.Stderr),  // Output to stderr (for piping)
    )
    
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

---

## Quick Reference

```go
// Essential imports
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/lipgloss"
)

// Common tea.Cmd returns
tea.Quit                           // Exit program
tea.ClearScreen                    // Clear terminal
tea.Batch(cmd1, cmd2)              // Run multiple commands
tea.Tick(duration, func)           // Timer tick
textinput.Blink                    // Cursor blink (for inputs)

// Common message types
tea.KeyMsg                         // Keyboard input
tea.MouseMsg                       // Mouse input  
tea.WindowSizeMsg                  // Terminal resize
```
