# today

A unified productivity dashboard for your terminal. Tasks, habits, and time tracking in one view.

```
 today   Tasks: 2/5  Habits: 3/4           ▶ LexEdge 01:23:45       Sat Dec 13 · 15:45
┌─────────────────────────┐ ┌──────────────────────┐ ┌──────────────────────────────┐
│ ✅ TASKS                │ │ ⏱️  TIMER            │ │ 🔥 HABITS                    │
│ ───────────────────     │ │ ──────────────────   │ │ ────────────────────         │
│  [✓] Review PR          │ │                      │ │                              │
│ ▶[ ] Finish spec        │ │   ▶ LexEdge          │ │ ▶🏃 Exercise  ○●●○●●● 5/7 🔥3│
│  [ ] Write blog         │ │     01:23:45         │ │  ✍️ Writing   ○●○○●●○ 3/7    │
│                         │ │                      │ │  📚 Reading   ○○○○○●● 2/7    │
│   1/3 complete          │ │   Today: 04:12:33    │ │                              │
│                         │ │   Week:  22:45:00    │ │   Best streak: 18 days 🔥    │
└─────────────────────────┘ └──────────────────────┘ └──────────────────────────────┘
[a] add  [d] done  [x] del  [j/k] nav  [tab] pane  [?] help
```

<!--
TODO: Add demo GIF here
![Demo](docs/demo.gif)
-->

## Features

- **📋 Tasks** — Add, complete, delete tasks with vim-style navigation
- **⏱️ Timer** — Track time by project, see daily/weekly totals
- **🔥 Habits** — Daily tracking with week view and streak counting
- **🎨 Beautiful TUI** — Modern terminal UI built with Bubble Tea
- **💾 Local Storage** — Plain JSON files, easy to backup or git sync
- **⌨️ Keyboard-first** — Vim keybindings, no mouse needed
- **📱 Responsive** — Adapts to terminal size
- **❓ Help Overlay** — Press `?` for full keyboard reference

## Installation

### Homebrew (macOS/Linux)

```bash
# Add the tap (first time only)
brew tap yourusername/tap

# Install
brew install today

# Run
today
```

### Binary Download (Recommended)

Download the latest release for your platform from the [releases page](https://github.com/yourusername/today/releases/latest):

**Linux (amd64):**
```bash
curl -LO https://github.com/yourusername/today/releases/latest/download/today_linux_amd64.tar.gz
tar -xzf today_linux_amd64.tar.gz
sudo mv today /usr/local/bin/
```

**Linux (arm64):**
```bash
curl -LO https://github.com/yourusername/today/releases/latest/download/today_linux_arm64.tar.gz
tar -xzf today_linux_arm64.tar.gz
sudo mv today /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -LO https://github.com/yourusername/today/releases/latest/download/today_darwin_amd64.tar.gz
tar -xzf today_darwin_amd64.tar.gz
sudo mv today /usr/local/bin/
```

**macOS (Apple Silicon):**
```bash
curl -LO https://github.com/yourusername/today/releases/latest/download/today_darwin_arm64.tar.gz
tar -xzf today_darwin_arm64.tar.gz
sudo mv today /usr/local/bin/
```

**Windows:**
Download `today_windows_amd64.zip` from the releases page and extract it to a directory in your PATH.

### Go Install

If you have Go 1.22 or later installed:

```bash
go install github.com/yourusername/today/cmd/today@latest
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/today.git
cd today

# Download dependencies
go mod download

# Build
go build -o today ./cmd/today

# Run
./today

# Optional: Install globally
sudo mv today /usr/local/bin/
```

### Verify Installation

```bash
today --version
```

## Usage

### Keybindings

**Global**

| Key | Action |
|-----|--------|
| `Tab` | Switch between panes |
| `1` | Focus tasks pane |
| `2` | Focus timer pane |
| `3` | Focus habits pane |
| `?` | Show help overlay |
| `q` | Quit |

**Tasks Pane**

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `a` | Add new task |
| `d` / `Enter` / `Space` | Toggle task done |
| `x` | Delete task |
| `g` | Go to top |
| `G` | Go to bottom |

**Timer Pane**

| Key | Action |
|-----|--------|
| `Space` / `Enter` | Start/stop timer |
| `s` | Switch project (starts new timer) |
| `x` | Stop timer |

**Habits Pane**

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `a` | Add new habit (name → icon) |
| `Space` / `Enter` / `d` | Toggle habit for today |
| `x` | Delete habit |

### When In Input Mode

| Key | Action |
|-----|--------|
| `Enter` | Save |
| `Esc` | Cancel |

## Data Storage

All data is stored in `~/.today/`:

```
~/.today/
├── tasks.json    # Your tasks
├── habits.json   # Habits and completion logs
└── timer.json    # Time tracking entries
```

Data is plain JSON — easy to backup, sync with git, or edit manually.

### Configuration

Optional configuration file: `~/.config/today/config.yaml`

```yaml
# Override default data directory
data_dir: ~/Documents/today-data
```

### Backup Your Data

Since everything is plain JSON, backing up is simple:

```bash
# Copy to Dropbox/iCloud/etc
cp -r ~/.today ~/Dropbox/backups/

# Or use git for version control
cd ~/.today
git init
git add .
git commit -m "Initial backup"
```

## Development Status

- [x] **Day 1:** Storage layer (complete)
- [x] **Day 2:** Task management TUI (complete)
- [x] **Day 3:** Timer functionality (complete)
- [x] **Day 4:** Habit tracking (complete)
- [x] **Day 5:** Layout + polish (complete)

**✅ v1.0 Ready!**

## Project Structure

```
today/
├── cmd/today/
│   └── main.go              # Entrypoint
├── internal/
│   ├── ui/
│   │   ├── app.go           # Main Bubble Tea app
│   │   ├── tasks.go         # Task pane component
│   │   ├── timer.go         # Timer pane component
│   │   ├── habits.go        # Habits pane component
│   │   ├── help.go          # Help overlay
│   │   └── styles.go        # Lipgloss styling
│   └── storage/
│       ├── models.go        # Data structures
│       └── storage.go       # JSON file operations
├── go.mod
└── README.md
```

## Contributing

Contributions are welcome! Here's how you can help:

### Reporting Bugs

1. Check if the issue already exists in [GitHub Issues](https://github.com/yourusername/today/issues)
2. If not, create a new issue with:
   - Clear description of the bug
   - Steps to reproduce
   - Expected vs actual behavior
   - Your OS and terminal type
   - Output of `today --version`

### Suggesting Features

1. Check [ROADMAP.md](ROADMAP.md) and existing issues first
2. Create a feature request issue with:
   - Clear use case
   - Proposed behavior
   - Any UI/UX considerations

### Submitting Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`go test ./...`)
5. Commit with descriptive messages
6. Push to your fork
7. Open a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/today.git
cd today

# Install dependencies
go mod download

# Run tests
go test ./...

# Build and run
go build -o today ./cmd/today
./today

# Run with race detector (recommended during development)
go run -race ./cmd/today
```

### Code Style

- Follow standard Go conventions (`gofmt`, `golint`)
- Add comments for exported functions
- Write tests for new features
- Keep commits atomic and well-described

## License

MIT

---

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features and future direction.

## Documentation

- [Man Page](docs/today.1) - Full command reference
- [Architecture](docs/ARCHITECTURE.md) - Technical documentation (if available)

## Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
