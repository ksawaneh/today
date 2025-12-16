// Package main is the entry point for the today application.
// It loads configuration, initializes storage, and starts the TUI.
package main

import (
	"flag"
	"fmt"
	"os"

	"today/internal/config"
	"today/internal/storage"
	"today/internal/sync"
	"today/internal/ui"
)

// Version information - set by GoReleaser during build
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const helpText = `today - A unified productivity dashboard for your terminal

USAGE:
    today [OPTIONS]
    today <command> [ARGS]

COMMANDS:
    backup           Create a backup of all data
    backup --list    List available backups
    restore NAME     Restore from a specific backup
    restore --latest Restore from the most recent backup
    export           Generate a daily report (Markdown)
    export --weekly  Generate a weekly report
    export -f json   Output report as JSON
    sync             Sync data with git (commit + push)
    sync --init      Initialize git repo in data directory
    sync --status    Show git sync status
    import           Import tasks from other apps
    import todoist   Import from Todoist CSV backup
    import taskwarrior  Import from Taskwarrior JSON

OPTIONS:
    -h, --help       Show this help message
    -v, --version    Show version information

DESCRIPTION:
    today is a terminal-based productivity app that combines tasks, habits,
    and time tracking in a single, keyboard-driven interface.

FEATURES:
    • Tasks      - Add, complete, delete tasks with vim-style navigation
    • Timer      - Track time by project, see daily/weekly totals
    • Habits     - Daily tracking with week view and streak counting
    • Local Data - Plain JSON files in ~/.today/

KEYBINDINGS:
    Global:
        Tab          Switch between panes
        1, 2, 3      Jump to specific pane
        ?            Show help overlay
        Ctrl+Z       Undo last action
        Ctrl+Y       Redo
        q            Quit

    Tasks Pane:
        j/k, ↓/↑     Navigate
        a            Add task
        d/Space      Toggle done
        x            Delete task
        g/G          Go to top/bottom

    Timer Pane:
        Space        Start/stop timer
        s            Switch project
        x            Stop timer

    Habits Pane:
        j/k, ↓/↑     Navigate
        a            Add habit
        d/Space      Toggle today's completion
        x            Delete habit

DATA STORAGE:
    All data is stored in ~/.today/ as plain JSON files:
        tasks.json   - Your tasks
        habits.json  - Habits and completion logs
        timer.json   - Time tracking entries

CONFIGURATION:
    Optional config file: ~/.config/today/config.yaml
    See documentation for configuration options.

EXAMPLES:
    # Start the app
    today

    # Create a backup
    today backup

    # Restore from a backup
    today restore --latest

    # Generate today's report
    today export

    # Generate weekly report as JSON
    today export --weekly --format json

    # Show version
    today --version

    # Show this help
    today --help

For more information, visit: https://github.com/yourusername/today
`

func main() {
	// Check for subcommands first (before flag parsing)
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "backup":
			runBackup(os.Args[2:])
			return
		case "restore":
			runRestore(os.Args[2:])
			return
		case "export":
			runExport(os.Args[2:])
			return
		case "sync":
			runSync(os.Args[2:])
			return
		case "import":
			runImport(os.Args[2:])
			return
		}
	}

	// Define flags
	showVersion := flag.Bool("version", false, "show version information")
	flag.BoolVar(showVersion, "v", false, "show version information (shorthand)")

	showHelp := flag.Bool("help", false, "show help message")
	flag.BoolVar(showHelp, "h", false, "show help message (shorthand)")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, helpText)
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("today version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	// Handle help flag
	if *showHelp {
		fmt.Print(helpText)
		os.Exit(0)
	}

	// Reject unknown arguments
	if flag.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "Error: unknown arguments: %v\n\n", flag.Args())
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration (from ~/.config/today/config.yaml or defaults)
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize storage with configured data directory
	store, err := storage.New(cfg.GetDataDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	// Set up git sync if enabled
	var gitSync *sync.GitSync
	if cfg.Sync.Enabled && sync.IsGitInstalled() {
		syncCfg := &sync.Config{
			Enabled:       cfg.Sync.Enabled,
			AutoCommit:    cfg.Sync.AutoCommit,
			AutoPush:      cfg.Sync.AutoPush,
			PullOnStartup: cfg.Sync.PullOnStartup,
			CommitMessage: cfg.Sync.CommitMessage,
		}
		gitSync = sync.New(cfg.GetDataDir(), syncCfg)

		// Pull on startup if configured and repo exists
		if cfg.Sync.PullOnStartup && gitSync.IsRepo() {
			if err := gitSync.Pull(); err != nil {
				// Log warning but continue - local data is still valid
				fmt.Fprintf(os.Stderr, "Warning: sync pull failed: %v\n", err)
			}
		}

		// Register auto-commit hook if enabled (with semantic context for meaningful messages)
		if cfg.Sync.AutoCommit && gitSync.IsRepo() {
			store.SetOnSaveWithContext(gitSync.OnFileSavedWithContext)
		}
	}

	// Create styles from theme config
	styles := ui.NewStylesFromTheme(&cfg.Theme)

	// Create app config with keys and UX settings
	appCfg := &ui.AppConfig{
		Keys:                  &cfg.Keys,
		ConfirmDeletions:      cfg.UX.ConfirmDeletions,
		ShowOnboarding:        cfg.UX.ShowOnboarding,
		NarrowLayoutThreshold: cfg.UX.NarrowLayoutThreshold,
	}

	// Run the TUI with optional GitSync for status display
	if err := ui.RunWithSync(store, styles, appCfg, gitSync); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}

	// Flush any pending git commits before exit
	if gitSync != nil {
		gitSync.Flush()
	}
}
