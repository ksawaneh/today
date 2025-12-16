// Package main is the entry point for the today application.
// This file contains the restore subcommand handler.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"today/internal/backup"
	"today/internal/config"
)

// restoreHelpText is the help message for the restore subcommand.
const restoreHelpText = `today restore - Restore data from a backup

USAGE:
    today restore [OPTIONS] [BACKUP_NAME]

OPTIONS:
    --latest       Restore from the most recent backup
    --force, -f    Skip confirmation prompt
    -h, --help     Show this help message

ARGUMENTS:
    BACKUP_NAME    Name of the backup to restore (e.g., 2025-12-15_143022_000)
                   Use 'today backup --list' to see available backups.

DESCRIPTION:
    Restores all data files (tasks, habits, timer) from a specific backup.
    A safety backup is automatically created before restoring.

EXAMPLES:
    # Restore from a specific backup
    today restore 2025-12-15_143022_000

    # Restore from the most recent backup
    today restore --latest

    # Restore without confirmation prompt
    today restore --force 2025-12-15_143022_000
`

// runRestore handles the "today restore" subcommand.
func runRestore(args []string) {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)

	latestFlag := fs.Bool("latest", false, "restore from most recent backup")
	forceFlag := fs.Bool("force", false, "skip confirmation prompt")
	fs.BoolVar(forceFlag, "f", false, "skip confirmation prompt (shorthand)")

	helpFlag := fs.Bool("help", false, "show help message")
	fs.BoolVar(helpFlag, "h", false, "show help message (shorthand)")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, restoreHelpText)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *helpFlag {
		fmt.Print(restoreHelpText)
		os.Exit(0)
	}

	// Load config to get data directory
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	manager := backup.NewManager(cfg.GetDataDir(), version)

	// Determine which backup to restore
	var backupName string
	if *latestFlag {
		backups, err := manager.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing backups: %v\n", err)
			os.Exit(1)
		}
		if len(backups) == 0 {
			fmt.Fprintln(os.Stderr, "No backups available.")
			os.Exit(1)
		}
		backupName = backups[0].Name
	} else if fs.NArg() > 0 {
		backupName = fs.Arg(0)
	} else {
		fmt.Fprintln(os.Stderr, "Error: no backup specified")
		fmt.Fprintln(os.Stderr, "Use 'today restore BACKUP_NAME' or 'today restore --latest'")
		fmt.Fprintln(os.Stderr, "Run 'today backup --list' to see available backups.")
		os.Exit(1)
	}

	// Get backup info
	info, err := manager.GetBackup(backupName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Display backup info
	fmt.Printf("Restoring from backup: %s\n", info.Name)
	fmt.Printf("  Created: %s\n", info.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Tasks: %d, Habits: %d, Timer entries: %d\n",
		info.Stats["tasks"], info.Stats["habits"], info.Stats["timer_entries"])
	fmt.Println()

	// Confirm unless --force is set
	if !*forceFlag {
		fmt.Println("⚠ This will overwrite your current data.")
		fmt.Print("Continue? [y/N] ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Restore cancelled.")
			os.Exit(0)
		}
	}

	// Perform restore
	fmt.Println("✓ Creating safety backup first...")
	if err := manager.Restore(backupName); err != nil {
		fmt.Fprintf(os.Stderr, "Error restoring backup: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Restored successfully from %s\n", backupName)
}
