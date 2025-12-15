# today-demo.tape
# Creates demo.gif showing core workflow
# Run with: vhs today-demo.tape

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
Type "./today"
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
Type "Update API documentation"
Enter
Sleep 1s

# Add third task
Type "a"
Sleep 500ms
Type "Fix login bug"
Enter
Sleep 1s

# Navigate and complete first task
Type "gg"
Sleep 300ms
Type "d"
Sleep 1s

# Switch to timer pane
Tab
Sleep 800ms

# Start timer with project
Type "s"
Sleep 500ms
Type "Code Review"
Enter
Sleep 2s

# Switch to habits pane
Tab
Sleep 800ms

# Add a habit
Type "a"
Sleep 500ms
Type "Exercise"
Enter
Sleep 300ms
Type "🏃"
Enter
Sleep 1s

# Toggle habit for today
Space
Sleep 1s

# Add another habit
Type "a"
Sleep 500ms
Type "Reading"
Enter
Sleep 300ms
Type "📚"
Enter
Sleep 1s

# Show help overlay
Type "?"
Sleep 2.5s
Escape
Sleep 500ms

# Back to tasks
Type "1"
Sleep 500ms

# Final pause to show dashboard
Sleep 2s

# Quit
Type "q"
Sleep 1s
