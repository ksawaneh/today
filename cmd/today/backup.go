// Package main is the entry point for the today application.
// This file contains the backup subcommand handler.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"today/internal/backup"
	"today/internal/config"
)

// backupHelpText is the help message for the backup subcommand.
const backupHelpText = `today backup - Create and manage backups

USAGE:
    today backup [OPTIONS]

OPTIONS:
    -l, --list     List available backups
    -h, --help     Show this help message

DESCRIPTION:
    Creates a timestamped backup of all your data files (tasks, habits, timer).
    Backups are stored in ~/.today/backups/ and can be restored later.

EXAMPLES:
    # Create a new backup
    today backup

    # List all available backups
    today backup --list
`

// runBackup handles the "today backup" subcommand.
func runBackup(args []string) {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)

	listFlag := fs.Bool("list", false, "list available backups")
	fs.BoolVar(listFlag, "l", false, "list available backups (shorthand)")

	helpFlag := fs.Bool("help", false, "show help message")
	fs.BoolVar(helpFlag, "h", false, "show help message (shorthand)")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, backupHelpText)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *helpFlag {
		fmt.Print(backupHelpText)
		os.Exit(0)
	}

	// Load config to get data directory
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	manager := backup.NewManager(cfg.GetDataDir(), version)

	if *listFlag {
		listBackups(manager)
	} else {
		createBackup(manager)
	}
}

// createBackup creates a new backup and displays the result.
func createBackup(manager *backup.Manager) {
	name, err := manager.Create()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating backup: %v\n", err)
		os.Exit(1)
	}

	info, err := manager.GetBackup(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading backup info: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Backup created: %s\n", name)
	fmt.Printf("  Tasks: %d, Habits: %d, Timer entries: %d\n",
		info.Stats["tasks"], info.Stats["habits"], info.Stats["timer_entries"])
	fmt.Printf("  Location: %s\n", info.Path)
}

// listBackups lists all available backups.
func listBackups(manager *backup.Manager) {
	backups, err := manager.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing backups: %v\n", err)
		os.Exit(1)
	}

	if len(backups) == 0 {
		fmt.Println("No backups available.")
		fmt.Println("Run 'today backup' to create one.")
		return
	}

	fmt.Println("Available backups:")
	for _, b := range backups {
		age := formatAge(b.CreatedAt)
		fmt.Printf("  %s  (%s)   Tasks: %d, Habits: %d\n",
			b.Name, age, b.Stats["tasks"], b.Stats["habits"])
	}
}

// formatAge returns a human-readable age string.
func formatAge(t time.Time) string {
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
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
}
