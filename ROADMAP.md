# today — Roadmap

This document outlines planned enhancements for `today`. Features are organized by priority and estimated effort.

---

## ✅ v1.0 (Complete)

- [x] Task management with vim-style navigation
- [x] Time tracking with project tagging
- [x] Habit tracking with streaks and week view
- [x] Three-pane responsive layout
- [x] Help overlay
- [x] Local JSON storage

---

## ✅ v1.1 — Quality of Life (Complete)

| Feature | Description | Status |
|---------|-------------|--------|
| **Config file** | User preferences (theme, keybinds, data dir) | ✅ Done |
| **Task projects/tags** | Organize tasks by project, filter by tag | ✅ Done |
| **Task priorities** | High/medium/low priority with visual indicators | ✅ Done |
| **Compact/responsive mode** | Stacked layout for smaller terminals | ✅ Done |
| **Mouse support** | Click to select, scroll to navigate | ✅ Done |
| **Undo/Redo** | Reversible destructive actions (ctrl+z/y) | ✅ Bonus |

### Config File Spec

```yaml
# ~/.config/today/config.yaml
theme: default  # default, minimal, nord
data_dir: ~/.today
keybinds:
  quit: q
  help: "?"
  add: a
timer:
  default_project: "General"
  show_seconds: true
habits:
  week_start: monday  # monday or sunday
```

---

## 🔄 v1.2 — Sync & Persistence

**Target: 2-3 weeks**

| Feature | Description | Effort |
|---------|-------------|--------|
| **Git sync** | Auto-commit changes to a git repo | 1-2 days |
| **Export reports** | Daily/weekly summaries as Markdown or JSON | 1 day |
| **Import from Todoist/Taskwarrior** | Migration tools for existing users | 2 days |
| **Backup/restore** | One-command backup and restore | 0.5 days |
| **Desktop notifications** | Notify on habit reminders, timer completion | 1-2 days |

### Git Sync Flow

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  User makes  │ ──▶ │  today auto  │ ──▶ │  Push to     │
│  a change    │     │  commits     │     │  remote      │
└──────────────┘     └──────────────┘     └──────────────┘
                            │
                            ▼
                     On startup, pull
                     latest changes
```

---

## 📅 v1.3 — Calendar Integration

**Target: 3-4 weeks**

| Feature | Description | Effort |
|---------|-------------|--------|
| **Calendar pane** | Fourth pane showing today's events | 2-3 days |
| **ICS import** | Import .ics files for local calendar | 1-2 days |
| **CalDAV sync** | Sync with Google Calendar, iCloud, etc. | 2-3 weeks |
| **Event creation** | Add events from the TUI | 1-2 days |
| **Conflict detection** | Warn about overlapping events | 1 day |

### Calendar Pane Mockup

```
┌────────────────────────┐
│ 📅 SCHEDULE            │
│ ──────────────────     │
│  09:00  Team standup   │
│  11:00  Dentist ✓      │
│  14:00  Call with Mo   │
│  16:30  Gym            │
│                        │
│  Tomorrow: 3 events    │
└────────────────────────┘
```

---

## 📊 v1.4 — Analytics & Insights

**Target: 2-3 weeks**

| Feature | Description | Effort |
|---------|-------------|--------|
| **Statistics view** | Dedicated stats pane/overlay | 2 days |
| **Time reports** | Where did my time go? Breakdown by project | 1-2 days |
| **Habit analytics** | Completion rates, best/worst days | 1-2 days |
| **Task velocity** | Tasks completed per day/week trends | 1 day |
| **Focus score** | Daily productivity score based on activity | 2 days |

### Stats Overlay Mockup

```
┌─────────────────────────────────────────────────────────────┐
│  📊 This Week's Stats                                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  TIME TRACKED                     TASKS                     │
│  ─────────────                    ─────                     │
│  Total: 32h 45m                   Completed: 23             │
│  Avg/day: 4h 41m                  Added: 31                 │
│                                   Velocity: +8%             │
│  By Project:                                                │
│    LexEdge      ████████████ 18h                           │
│    Writing      ██████ 8h                                   │
│    Admin        ████ 6h                                     │
│                                                             │
│  HABITS                           STREAKS                   │
│  ──────                           ───────                   │
│  Completion: 85%                  Exercise: 12 days 🔥      │
│  Best day: Tuesday                Writing: 5 days           │
│  Worst day: Sunday                Reading: 3 days           │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## 🌐 v2.0 — Multi-Device & Collaboration

**Target: 6-8 weeks**

| Feature | Description | Effort |
|---------|-------------|--------|
| **Cloud sync** | Optional Anthropic/self-hosted sync server | 3-4 weeks |
| **Mobile companion** | Quick capture app (iOS/Android) | 4-6 weeks |
| **Shared dashboards** | Team view for accountability partners | 2-3 weeks |
| **API** | REST API for integrations | 1-2 weeks |
| **Webhooks** | Trigger external actions on events | 1 week |

---

## 🎨 v2.1 — Customization & Themes

**Target: 2-3 weeks**

| Feature | Description | Effort |
|---------|-------------|--------|
| **Built-in themes** | Nord, Dracula, Solarized, Gruvbox | 2-3 days |
| **Custom themes** | User-defined color schemes | 1-2 days |
| **Layout presets** | Different pane arrangements | 1-2 days |
| **Custom keybinds** | Fully remappable keys | 2 days |
| **Plugins** | Lua/Go plugin system | 2-3 weeks |

---

## 💡 Ideas Backlog

Lower priority or speculative features:

| Idea | Description |
|------|-------------|
| **Pomodoro mode** | Built-in pomodoro timer with breaks |
| **Focus mode** | Hide everything except current task + timer |
| **Daily review** | End-of-day guided review flow |
| **Templates** | Recurring task templates |
| **Natural language** | "Add task meeting tomorrow 3pm" |
| **Voice notes** | Attach voice memos to tasks |
| **AI suggestions** | Smart task prioritization |
| **Offline PWA** | Web version that works offline |
| **VS Code extension** | Today sidebar in your editor |
| **Alfred/Raycast** | Quick capture integrations |

---

## Contributing

Contributions welcome! If you want to tackle any of these:

1. Open an issue to discuss the approach
2. Fork the repo
3. Create a feature branch
4. Submit a PR

### Development Setup

```bash
git clone https://github.com/yourusername/today
cd today
go mod tidy
go build -o today ./cmd/today
./today
```

---

## Versioning

We use [SemVer](https://semver.org/):

- **MAJOR** — Breaking changes to data format or CLI
- **MINOR** — New features, backward compatible
- **PATCH** — Bug fixes, minor improvements

---

*Last updated: December 2025*
