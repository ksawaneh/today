// Package main is the entry point for the today application.
// This file contains the export subcommand handler.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"today/internal/config"
	"today/internal/fsutil"
	"today/internal/reports"
	"today/internal/storage"
)

// exportHelpText is the help message for the export subcommand.
const exportHelpText = `today export - Generate productivity reports

USAGE:
    today export [OPTIONS] [DATE]

OPTIONS:
    -d, --daily        Generate daily report (default)
    -w, --weekly       Generate weekly report
    -f, --format FMT   Output format: markdown (default) or json
    -o, --output FILE  Write to file instead of stdout
    -h, --help         Show this help message

ARGUMENTS:
    DATE               Date for report (YYYY-MM-DD). Defaults to today.
                       For weekly reports, this is the start of the week.

DESCRIPTION:
    Generates reports summarizing your tasks, time tracking, and habits.
    Reports can be output as Markdown (human-readable) or JSON (machine-readable).

EXAMPLES:
    # Today's report in Markdown
    today export

    # Specific date
    today export 2025-12-14

    # Weekly report
    today export --weekly

    # JSON format
    today export --format json

    # Save to file
    today export --output report.md

    # Weekly JSON report to file
    today export --weekly --format json --output weekly.json
`

// runExport handles the "today export" subcommand.
func runExport(args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)

	dailyFlag := fs.Bool("daily", false, "generate daily report")
	fs.BoolVar(dailyFlag, "d", false, "generate daily report (shorthand)")

	weeklyFlag := fs.Bool("weekly", false, "generate weekly report")
	fs.BoolVar(weeklyFlag, "w", false, "generate weekly report (shorthand)")

	formatFlag := fs.String("format", "markdown", "output format: markdown or json")
	fs.StringVar(formatFlag, "f", "markdown", "output format (shorthand)")

	outputFlag := fs.String("output", "", "write to file instead of stdout")
	fs.StringVar(outputFlag, "o", "", "write to file (shorthand)")

	helpFlag := fs.Bool("help", false, "show help message")
	fs.BoolVar(helpFlag, "h", false, "show help message (shorthand)")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, exportHelpText)
	}

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *helpFlag {
		fmt.Print(exportHelpText)
		os.Exit(0)
	}

	// Validate format
	format := *formatFlag
	if format != "markdown" && format != "json" && format != "md" {
		fmt.Fprintf(os.Stderr, "Error: invalid format %q. Use 'markdown' or 'json'.\n", format)
		os.Exit(1)
	}
	if format == "md" {
		format = "markdown"
	}

	// Parse date argument
	date := time.Now()
	if fs.NArg() > 0 {
		parsedDate, err := time.ParseInLocation("2006-01-02", fs.Arg(0), time.Local)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid date %q. Use YYYY-MM-DD format.\n", fs.Arg(0))
			os.Exit(1)
		}
		date = parsedDate
	}

	// Determine report type (default to daily)
	isWeekly := *weeklyFlag
	// If neither flag is set, default to daily
	if !*dailyFlag && !*weeklyFlag {
		isWeekly = false
	}

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

	gen := reports.NewGenerator(store)

	// Generate report
	var output string
	if isWeekly {
		report, err := gen.GenerateWeekly(date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating weekly report: %v\n", err)
			os.Exit(1)
		}

		if format == "json" {
			data, err := reports.FormatWeeklyJSON(report)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
				os.Exit(1)
			}
			output = string(data)
		} else {
			output = reports.FormatWeeklyMarkdown(report)
		}
	} else {
		report, err := gen.GenerateDaily(date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating daily report: %v\n", err)
			os.Exit(1)
		}

		if format == "json" {
			data, err := reports.FormatDailyJSON(report)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
				os.Exit(1)
			}
			output = string(data)
		} else {
			output = reports.FormatDailyMarkdown(report)
		}
	}

	// Write output
	if *outputFlag != "" {
		if err := os.MkdirAll(filepath.Dir(*outputFlag), 0700); err != nil && filepath.Dir(*outputFlag) != "." {
			fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
			os.Exit(1)
		}
		if err := fsutil.WriteFileAtomic(*outputFlag, []byte(output), 0600); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Report written to %s\n", *outputFlag)
	} else {
		fmt.Print(output)
	}
}
