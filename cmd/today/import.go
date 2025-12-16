// Package main is the entry point for the today application.
// This file contains the import subcommand handler.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"today/internal/config"
	"today/internal/importer"
	"today/internal/storage"
)

// importHelpText is the help message for the import subcommand.
const importHelpText = `today import - Import tasks from other apps

USAGE:
    today import <format> <file>
    today import [OPTIONS] <format> <file>

FORMATS:
    todoist      Import from Todoist CSV backup
    taskwarrior  Import from Taskwarrior JSON export

OPTIONS:
    --dry-run    Preview import without making changes
    -h, --help   Show this help message

DESCRIPTION:
    Import tasks from other productivity tools. Supported formats:

    TODOIST:
      Export your tasks from Todoist via Settings → Backups.
      The backup will be a CSV file that can be imported directly.

    TASKWARRIOR:
      Export your tasks using: task export > tasks.json
      Both JSON array and newline-delimited JSON formats are supported.

FIELD MAPPING:
    Todoist:
      - CONTENT → task text
      - PRIORITY: 1,2 → high, 3 → medium, 4 → low
      - DATE → due date
      - Notes are skipped

    Taskwarrior:
      - description → task text
      - project → project
      - priority: H → high, M → medium, L → low
      - due → due date
      - status: completed → marks task as done
      - Deleted tasks are skipped

EXAMPLES:
    # Import from Todoist
    today import todoist ~/Downloads/Todoist_backup.csv

    # Import from Taskwarrior
    task export > tasks.json
    today import taskwarrior tasks.json

    # Preview before importing
    today import --dry-run todoist backup.csv

    # Show help
    today import --help
`

// runImport handles the "today import" subcommand.
func runImport(args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)

	dryRunFlag := fs.Bool("dry-run", false, "preview import without making changes")
	helpFlag := fs.Bool("help", false, "show help message")
	fs.BoolVar(helpFlag, "h", false, "show help message (shorthand)")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, importHelpText)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *helpFlag {
		fmt.Print(importHelpText)
		os.Exit(0)
	}

	// Need at least format and file
	if fs.NArg() < 2 {
		fmt.Fprintf(os.Stderr, "Error: missing arguments\n\n")
		fmt.Fprintf(os.Stderr, "Usage: today import <format> <file>\n")
		fmt.Fprintf(os.Stderr, "Formats: %s\n", strings.Join(importer.SupportedFormats(), ", "))
		fmt.Fprintf(os.Stderr, "\nRun 'today import --help' for more information.\n")
		os.Exit(1)
	}

	format := strings.ToLower(fs.Arg(0))
	filePath := fs.Arg(1)

	// Get importer
	imp := importer.GetImporter(format)
	if imp == nil {
		fmt.Fprintf(os.Stderr, "Error: unknown format %q\n", format)
		fmt.Fprintf(os.Stderr, "Supported formats: %s\n", strings.Join(importer.SupportedFormats(), ", "))
		os.Exit(1)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if *dryRunFlag {
		runImportDryRun(imp, file)
	} else {
		runImportActual(imp, file)
	}
}

// runImportDryRun previews the import without making changes.
func runImportDryRun(imp importer.Importer, file *os.File) {
	tasks, err := imp.Preview(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing file: %v\n", err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found to import.")
		os.Exit(0)
	}

	fmt.Printf("Preview: %d tasks to import\n", len(tasks))
	fmt.Println("────────────────────────────")

	// Show first 20 tasks
	showCount := len(tasks)
	if showCount > 20 {
		showCount = 20
	}

	for i := 0; i < showCount; i++ {
		task := tasks[i]
		fmt.Printf("  %s", task.Text)

		var details []string
		if task.Project != "" {
			details = append(details, task.Project)
		}
		if task.Priority != "" {
			details = append(details, string(task.Priority))
		}
		if task.DueDate != nil {
			details = append(details, task.DueDate.Format("2006-01-02"))
		}
		if task.Done {
			details = append(details, "done")
		}

		if len(details) > 0 {
			fmt.Printf(" (%s)", strings.Join(details, ", "))
		}
		fmt.Println()
	}

	if len(tasks) > 20 {
		fmt.Printf("  ... and %d more\n", len(tasks)-20)
	}

	fmt.Println()
	fmt.Println("Run without --dry-run to import.")
}

// runImportActual performs the actual import.
func runImportActual(imp importer.Importer, file *os.File) {
	// Load config and storage
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	store, err := storage.New(cfg.GetDataDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	// Perform import
	result, err := imp.Import(file, store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error importing: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("Import complete!\n")
	fmt.Printf("  Imported: %d tasks\n", result.Imported)
	if result.Skipped > 0 {
		fmt.Printf("  Skipped:  %d items\n", result.Skipped)
	}
	if len(result.Errors) > 0 {
		fmt.Printf("  Errors:   %d\n", len(result.Errors))
		for _, e := range result.Errors {
			fmt.Printf("    - %s\n", e)
		}
	}
}
