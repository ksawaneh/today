# Terminal Recording Skill

A comprehensive guide for creating high-quality terminal recordings and demo GIFs using VHS and other tools.

## When to Use This Skill

- Creating demo GIFs for README files
- Recording terminal sessions for documentation
- Making tutorial videos
- Showcasing CLI/TUI applications

---

## VHS (Charmbracelet)

VHS is a tool for creating terminal GIFs from a simple scripting language.

### Installation

```bash
# macOS
brew install vhs

# Go
go install github.com/charmbracelet/vhs@latest

# Arch Linux
yay -S vhs
```

### Basic Tape File

```tape
# demo.tape
Output demo.gif

# Terminal settings
Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 600
Set Framerate 60

# Theme
Set Theme "Dracula"

# Recording
Type "today"
Enter
Sleep 2s

Type "a"
Sleep 500ms
Type "Review pull request"
Enter
Sleep 1s

Type "a"
Sleep 500ms
Type "Write documentation"
Enter
Sleep 1s

Type "j"
Sleep 500ms
Type "d"
Sleep 1s

Type "q"
Sleep 500ms
```

### Run Recording

```bash
vhs demo.tape
```

---

## VHS Command Reference

### Output Settings

```tape
# Output formats
Output demo.gif           # Animated GIF
Output demo.mp4           # MP4 video
Output demo.webm          # WebM video
Output frames/            # PNG frames

# Multiple outputs
Output demo.gif
Output demo.mp4
```

### Terminal Configuration

```tape
# Dimensions
Set Width 1200            # Terminal width in pixels
Set Height 600            # Terminal height in pixels

# Font
Set FontSize 14           # Font size
Set FontFamily "JetBrains Mono"  # Font family

# Timing
Set Framerate 60          # Frames per second
Set PlaybackSpeed 1.0     # 0.5 = half speed, 2.0 = double speed

# Shell
Set Shell "bash"          # bash, zsh, fish, powershell
Set Shell "zsh"

# Theme (built-in themes)
Set Theme "Dracula"
Set Theme "Catppuccin Mocha"
Set Theme "Nord"
Set Theme "Tokyo Night"
Set Theme "One Dark"
Set Theme "Gruvbox"
```

### Input Commands

```tape
# Type text (with realistic typing delay)
Type "hello world"
Type@100ms "fast typing"   # Custom typing speed
Type@500ms "slow typing"

# Special keys
Enter                      # Press Enter
Backspace                  # Press Backspace
Delete                     # Press Delete
Tab                        # Press Tab
Space                      # Press Space
Escape                     # Press Escape

# Arrow keys
Up
Down
Left
Right

# Modifier combinations
Ctrl+C                     # Ctrl+C
Ctrl+D                     # Ctrl+D
Alt+Enter                  # Alt+Enter

# Function keys
F1
F12
```

### Timing Commands

```tape
# Wait
Sleep 1s                   # Sleep 1 second
Sleep 500ms                # Sleep 500 milliseconds
Sleep 2.5s                 # Sleep 2.5 seconds

# Pause for screenshot
Screenshot screenshot.png
```

### Control Commands

```tape
# Hide commands (don't show in output)
Hide
Type "secret-setup-command"
Enter
Show

# Source another tape
Source setup.tape
```

---

## Best Practices for Demo GIFs

### 1. Keep It Short

```tape
# Bad: 60+ second demo
# Good: 10-15 seconds showing core workflow

# Aim for:
# - One clear feature per GIF
# - Maximum 15 seconds
# - 3-5 key interactions
```

### 2. Start with Clear State

```tape
# Clear terminal first
Hide
Type "clear"
Enter
Show
Sleep 500ms
```

### 3. Use Appropriate Delays

```tape
# Give viewers time to read
Type "today"
Enter
Sleep 2s                   # Time to see the UI

# Faster for repetitive actions
Type "jjj"                 # Navigate down
Sleep 500ms                # Brief pause

# Slow down for important moments
Type "a"                   # Add task
Sleep 1s                   # Time to notice input mode
Type "Important task"
Sleep 500ms
Enter
Sleep 1.5s                 # Time to see result
```

### 4. Highlight Key Actions

```tape
# Pause before important actions
Sleep 500ms
Type "d"                   # Mark done (important!)
Sleep 1.5s                 # Show the result

# Visual break between sections
Sleep 1s
```

### 5. End Cleanly

```tape
# Show final state
Sleep 2s

# Or exit gracefully
Type "q"
Sleep 1s
```

---

## Complete Example: Today App Demo

```tape
# today-demo.tape
# Creates demo.gif showing core workflow

Output demo.gif
Output demo.mp4

# Terminal setup
Set Shell "bash"
Set FontSize 16
Set Width 1400
Set Height 800
Set Framerate 60
Set Theme "Catppuccin Mocha"

# Clear and start
Hide
Type "cd ~ && clear"
Enter
Show
Sleep 500ms

# Launch app
Type "today"
Enter
Sleep 2s

# Add first task
Type "a"
Sleep 500ms
Type "Review pull request #42"
Enter
Sleep 1s

# Add second task
Type "a"
Sleep 500ms
Type "Update documentation"
Enter
Sleep 1s

# Navigate and complete
Type "k"
Sleep 300ms
Type "d"
Sleep 1s

# Switch to timer pane
Tab
Sleep 500ms

# Start timer
Type "s"
Sleep 500ms
Type "Code review"
Enter
Sleep 2s

# Switch to habits pane
Tab
Sleep 500ms

# Toggle habit
Space
Sleep 1s

# Show help
Type "?"
Sleep 2s
Escape
Sleep 500ms

# Quit
Type "q"
Sleep 1s
```

---

## Alternative: asciinema

For longer recordings or interactive playback:

### Record

```bash
asciinema rec demo.cast
# ... do your thing ...
# Ctrl+D to stop

# With specific settings
asciinema rec -t "My Demo" -i 2 demo.cast
```

### Upload and Share

```bash
asciinema upload demo.cast
# Returns URL like: https://asciinema.org/a/123456
```

### Embed in README

```markdown
[![asciicast](https://asciinema.org/a/123456.svg)](https://asciinema.org/a/123456)
```

### Convert to GIF

```bash
# Using agg (asciinema gif generator)
agg demo.cast demo.gif

# Using asciicast2gif
asciicast2gif demo.cast demo.gif
```

---

## GIF Optimization

### Reduce File Size

```bash
# Using gifsicle
gifsicle -O3 --colors 256 demo.gif -o demo-optimized.gif

# Reduce colors more aggressively
gifsicle -O3 --colors 64 demo.gif -o demo-small.gif

# Resize
gifsicle --resize-width 800 demo.gif -o demo-resized.gif
```

### Convert to WebP (Smaller)

```bash
# Using gif2webp
gif2webp demo.gif -o demo.webp

# Or with ffmpeg
ffmpeg -i demo.gif -c:v libwebp demo.webp
```

---

## README Integration

### Basic Embed

```markdown
# My CLI Tool

![Demo](./demo.gif)
```

### With Link to Larger Version

```markdown
# My CLI Tool

<a href="./demo.mp4">
  <img src="./demo.gif" alt="Demo" width="600">
</a>

*Click for full quality video*
```

### Multiple Demos

```markdown
## Features

### Task Management
![Tasks Demo](./demos/tasks.gif)

### Time Tracking
![Timer Demo](./demos/timer.gif)

### Habit Tracking
![Habits Demo](./demos/habits.gif)
```

---

## CI Integration

### Generate GIF in GitHub Actions

```yaml
# .github/workflows/demo.yml
name: Generate Demo GIF

on:
  push:
    paths:
      - 'demo.tape'
  workflow_dispatch:

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install VHS
        run: |
          sudo mkdir -p /etc/apt/keyrings
          curl -fsSL https://repo.charm.sh/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/charm.gpg
          echo "deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *" | sudo tee /etc/apt/sources.list.d/charm.list
          sudo apt update && sudo apt install vhs
      
      - name: Build app
        run: go build -o today ./cmd/today
      
      - name: Generate GIF
        run: vhs demo.tape
      
      - name: Commit GIF
        run: |
          git config user.name github-actions
          git config user.email github-actions@github.com
          git add demo.gif
          git commit -m "Update demo GIF" || exit 0
          git push
```

---

## Troubleshooting

### Font Issues

```tape
# Use a common monospace font
Set FontFamily "Menlo"           # macOS
Set FontFamily "Consolas"        # Windows
Set FontFamily "DejaVu Sans Mono" # Linux

# Or use a Nerd Font for icons
Set FontFamily "JetBrainsMono Nerd Font"
```

### Color Issues

```tape
# Force true color
Set Shell "bash"
Hide
Type "export COLORTERM=truecolor"
Enter
Show
```

### TTY Issues

```tape
# Some apps need TTY
Set Shell "script -q /dev/null bash"
```

### Slow Recording

```tape
# Reduce framerate for smaller files
Set Framerate 30

# Use lower resolution
Set Width 800
Set Height 400
```

---

## Quick Reference

```tape
# Essential settings
Output demo.gif
Set Width 1200
Set Height 600
Set FontSize 14
Set Theme "Dracula"

# Common commands
Type "text"          # Type text
Enter                # Press enter
Sleep 1s             # Wait
Tab                  # Tab key
Ctrl+C               # Ctrl+C

# Run
vhs demo.tape

# Optimize output
gifsicle -O3 demo.gif -o demo-opt.gif
```
