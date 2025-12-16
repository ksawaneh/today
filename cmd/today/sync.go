// Package main is the entry point for the today application.
// This file contains the sync subcommand handler.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"today/internal/config"
	"today/internal/sync"
)

// syncHelpText is the help message for the sync subcommand.
const syncHelpText = `today sync - Git synchronization for your data

USAGE:
    today sync [OPTIONS]

OPTIONS:
    --setup        Interactive setup wizard (recommended for first-time setup)
    --init         Initialize git repository in data directory
    --status       Show sync status
    --pull         Pull latest changes from remote
    --push         Push local changes to remote
    -h, --help     Show this help message

DESCRIPTION:
    Manages git synchronization for your today data. Your tasks, habits, and
    timer data can be automatically committed to a git repository for backup
    and sync across machines.

SETUP:
    1. Initialize the repository:
       today sync --init

    2. Add a remote (manually):
       cd ~/.today && git remote add origin <your-repo-url>

    3. Enable sync in config (~/.config/today/config.yaml):
       sync:
         enabled: true
         auto_commit: true
         auto_push: false
         pull_on_startup: false

EXAMPLES:
    # Initialize git repo in data directory
    today sync --init

    # Check sync status
    today sync --status

    # Manual sync (commit + push)
    today sync

    # Pull latest changes
    today sync --pull

    # Push local commits
    today sync --push

CONFIGURATION:
    sync:
      enabled: false           # Enable/disable git sync
      auto_commit: true        # Automatically commit after changes
      auto_push: false         # Automatically push after commits
      pull_on_startup: false   # Pull when starting the app
      commit_message: "auto"   # "auto" or custom message template
`

// runSync handles the "today sync" subcommand.
func runSync(args []string) {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)

	setupFlag := fs.Bool("setup", false, "interactive setup wizard")
	initFlag := fs.Bool("init", false, "initialize git repository")
	statusFlag := fs.Bool("status", false, "show sync status")
	pullFlag := fs.Bool("pull", false, "pull latest changes")
	pushFlag := fs.Bool("push", false, "push local changes")
	helpFlag := fs.Bool("help", false, "show help message")
	fs.BoolVar(helpFlag, "h", false, "show help message (shorthand)")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, syncHelpText)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *helpFlag {
		fmt.Print(syncHelpText)
		os.Exit(0)
	}

	// Check if git is installed
	if !sync.IsGitInstalled() {
		fmt.Fprintf(os.Stderr, "Error: git is not installed. Please install git to use sync.\n")
		os.Exit(1)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create sync config from app config
	syncCfg := &sync.Config{
		Enabled:       cfg.Sync.Enabled,
		AutoCommit:    cfg.Sync.AutoCommit,
		AutoPush:      cfg.Sync.AutoPush,
		PullOnStartup: cfg.Sync.PullOnStartup,
		CommitMessage: cfg.Sync.CommitMessage,
	}

	gs := sync.New(cfg.GetDataDir(), syncCfg)

	// Handle flags
	switch {
	case *setupFlag:
		runSyncSetup(gs, cfg)
	case *initFlag:
		runSyncInit(gs, cfg.GetDataDir())
	case *statusFlag:
		runSyncStatus(gs, cfg)
	case *pullFlag:
		runSyncPull(gs)
	case *pushFlag:
		runSyncPush(gs)
	default:
		// Default: commit all + push
		runSyncDefault(gs)
	}
}

// runSyncInit initializes the git repository.
func runSyncInit(gs *sync.GitSync, dataDir string) {
	if gs.IsRepo() {
		fmt.Printf("Git repository already initialized in %s\n", dataDir)
		os.Exit(0)
	}

	fmt.Printf("Initializing git repository in %s...\n", dataDir)
	if err := gs.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Repository initialized successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Add a remote repository:")
	fmt.Printf("     cd %s && git remote add origin <your-repo-url>\n", dataDir)
	fmt.Println()
	fmt.Println("  2. Enable sync in your config (~/.config/today/config.yaml):")
	fmt.Println("     sync:")
	fmt.Println("       enabled: true")
}

// runSyncStatus shows the sync status.
func runSyncStatus(gs *sync.GitSync, cfg *config.Config) {
	status, err := gs.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Git Sync Status")
	fmt.Println("───────────────")

	if cfg.Sync.Enabled {
		fmt.Println("Sync:       enabled")
	} else {
		fmt.Println("Sync:       disabled")
	}

	fmt.Printf("Data dir:   %s\n", cfg.GetDataDir())

	if !status.IsRepo {
		fmt.Println("Repository: not initialized")
		fmt.Println()
		fmt.Println("Run 'today sync --init' to initialize.")
		return
	}

	fmt.Printf("Repository: initialized\n")
	fmt.Printf("Branch:     %s\n", status.Branch)

	if status.HasRemote {
		fmt.Printf("Remote:     %s (%s)\n", status.RemoteName, status.RemoteURL)
		if status.Ahead > 0 || status.Behind > 0 {
			fmt.Printf("Status:     %d ahead, %d behind\n", status.Ahead, status.Behind)
		} else {
			fmt.Println("Status:     up to date")
		}
	} else {
		fmt.Println("Remote:     not configured")
	}

	if status.HasChanges {
		fmt.Println("Changes:    uncommitted changes present")
	} else {
		fmt.Println("Changes:    clean")
	}

	if status.LastCommitAt != nil {
		fmt.Printf("Last commit: %s\n", formatTimeAgo(*status.LastCommitAt))
	}
}

// runSyncPull pulls from the remote.
func runSyncPull(gs *sync.GitSync) {
	if !gs.IsRepo() {
		fmt.Fprintf(os.Stderr, "Error: not a git repository. Run 'today sync --init' first.\n")
		os.Exit(1)
	}

	fmt.Println("Pulling latest changes...")
	if err := gs.Pull(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Pull complete.")
}

// runSyncPush pushes to the remote.
func runSyncPush(gs *sync.GitSync) {
	if !gs.IsRepo() {
		fmt.Fprintf(os.Stderr, "Error: not a git repository. Run 'today sync --init' first.\n")
		os.Exit(1)
	}

	fmt.Println("Pushing local changes...")
	if err := gs.Push(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Push complete.")
}

// runSyncDefault performs a manual sync (commit all + push).
func runSyncDefault(gs *sync.GitSync) {
	if !gs.IsRepo() {
		fmt.Fprintf(os.Stderr, "Error: not a git repository. Run 'today sync --init' first.\n")
		os.Exit(1)
	}

	// Commit all changes
	fmt.Println("Committing changes...")
	if err := gs.CommitAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Error committing: %v\n", err)
		os.Exit(1)
	}

	// Check if there's a remote before trying to push
	status, err := gs.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
		os.Exit(1)
	}

	if status.HasRemote {
		fmt.Println("Pushing to remote...")
		if err := gs.Push(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: push failed: %v\n", err)
			fmt.Println("Changes committed locally.")
			os.Exit(0)
		}
		fmt.Println("Sync complete.")
	} else {
		fmt.Println("Changes committed locally.")
		fmt.Println("(No remote configured - add one with 'git remote add origin <url>')")
	}
}

// runSyncSetup runs the interactive setup wizard.
func runSyncSetup(gs *sync.GitSync, cfg *config.Config) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	fmt.Println("Git Sync Setup")
	fmt.Println("══════════════")
	fmt.Println()
	fmt.Println("This wizard will help you set up git sync for your today data.")
	fmt.Println()

	// Step 1: Check/Initialize repository
	fmt.Printf("Data directory: %s\n", cfg.GetDataDir())
	fmt.Println()

	if !gs.IsRepo() {
		fmt.Print("Initialize git repository? [Y/n] ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "" || response == "y" || response == "yes" {
			fmt.Println("Initializing repository...")
			if err := gs.Init(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("✓ Repository initialized")
		} else {
			fmt.Println("Setup canceled.")
			os.Exit(0)
		}
	} else {
		fmt.Println("✓ Repository already initialized")
	}
	fmt.Println()

	// Step 2: Check/Add remote
	status, err := gs.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
		os.Exit(1)
	}

	if !status.HasRemote {
		fmt.Print("Add a remote repository? [y/N] ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "y" || response == "yes" {
			fmt.Print("Remote URL (e.g., git@github.com:user/today-data.git): ")
			remoteURL, _ := reader.ReadString('\n')
			remoteURL = strings.TrimSpace(remoteURL)

			if remoteURL != "" {
				if err := gs.AddRemote("origin", remoteURL); err != nil {
					fmt.Fprintf(os.Stderr, "Error adding remote: %v\n", err)
				} else {
					fmt.Println("✓ Remote 'origin' added")
				}
			} else {
				fmt.Println("Skipped (no URL provided)")
			}
		} else {
			fmt.Println("Skipped")
		}
	} else {
		fmt.Printf("✓ Remote configured: %s (%s)\n", status.RemoteName, status.RemoteURL)
	}
	fmt.Println()

	// Step 3: Configure sync options
	fmt.Println("Configuration Options")
	fmt.Println("─────────────────────")
	fmt.Println()

	// Auto-commit
	autoCommit := cfg.Sync.AutoCommit
	fmt.Printf("Enable auto-commit (commits after each change)? [%s] ", yesNoDefault(autoCommit))
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "" {
		autoCommit = response == "y" || response == "yes"
	}

	// Auto-push
	autoPush := cfg.Sync.AutoPush
	fmt.Printf("Enable auto-push (pushes after each commit)? [%s] ", yesNoDefault(autoPush))
	response, _ = reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "" {
		autoPush = response == "y" || response == "yes"
	}

	// Pull on startup
	pullOnStartup := cfg.Sync.PullOnStartup
	fmt.Printf("Pull on startup (sync before opening app)? [%s] ", yesNoDefault(pullOnStartup))
	response, _ = reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "" {
		pullOnStartup = response == "y" || response == "yes"
	}

	fmt.Println()

	// Step 4: Save config
	cfg.Sync.Enabled = true
	cfg.Sync.AutoCommit = autoCommit
	cfg.Sync.AutoPush = autoPush
	cfg.Sync.PullOnStartup = pullOnStartup

	if err := cfg.Save(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not save config: %v\n", err)
		fmt.Println()
		fmt.Println("Add this to your config file (~/.config/today/config.yaml):")
		fmt.Println()
		fmt.Println("sync:")
		fmt.Println("  enabled: true")
		fmt.Printf("  auto_commit: %v\n", autoCommit)
		fmt.Printf("  auto_push: %v\n", autoPush)
		fmt.Printf("  pull_on_startup: %v\n", pullOnStartup)
	} else {
		fmt.Println("✓ Configuration saved")
	}

	fmt.Println()
	fmt.Println("Setup complete! Git sync is now enabled.")
	fmt.Println()
	fmt.Println("Your data will be automatically committed to git.")
	if autoPush {
		fmt.Println("Changes will be pushed automatically after each commit.")
	} else {
		fmt.Println("Use 'today sync' to push changes to your remote.")
	}
}

// yesNoDefault returns "Y/n" or "y/N" based on the default value.
func yesNoDefault(defaultYes bool) string {
	if defaultYes {
		return "Y/n"
	}
	return "y/N"
}

// formatTimeAgo formats a time as a human-readable "time ago" string.
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)

	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "yesterday"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		return t.Format("Jan 2, 2006")
	}
}
