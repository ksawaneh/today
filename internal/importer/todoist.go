// Package importer provides import functionality for the today app.
// This file implements Todoist CSV import.
package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"today/internal/storage"
)

// TodoistImporter handles importing from Todoist CSV exports.
type TodoistImporter struct{}

// Name returns the importer name.
func (t *TodoistImporter) Name() string {
	return "todoist"
}

// Import reads tasks from Todoist CSV and adds them to storage.
func (t *TodoistImporter) Import(reader io.Reader, store *storage.Storage) (*ImportResult, error) {
	tasks, err := t.parseTasks(reader)
	if err != nil {
		return nil, err
	}

	result := &ImportResult{}

	for _, task := range tasks {
		_, err := store.AddTask(task.Text, task.Project, task.Priority, task.DueDate)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", task.Text, err))
			continue
		}
		result.Imported++
	}

	return result, nil
}

// Preview returns a list of tasks that would be imported.
func (t *TodoistImporter) Preview(reader io.Reader) ([]PreviewTask, error) {
	return t.parseTasks(reader)
}

// parseTasks reads and parses the Todoist CSV format.
func (t *TodoistImporter) parseTasks(reader io.Reader) ([]PreviewTask, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.ReuseRecord = true

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Find column indices
	colIndex := make(map[string]int)
	for i, col := range header {
		if i == 0 {
			col = strings.TrimPrefix(col, "\ufeff") // UTF-8 BOM (common in some exports)
		}
		colIndex[strings.ToUpper(strings.TrimSpace(col))] = i
	}

	// Verify required columns
	requiredCols := []string{"TYPE", "CONTENT"}
	for _, col := range requiredCols {
		if _, ok := colIndex[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	var tasks []PreviewTask

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}
		if len(record) == 0 {
			continue
		}

		// Skip non-task rows
		typeIdx := colIndex["TYPE"]
		if typeIdx >= len(record) || strings.ToLower(record[typeIdx]) != "task" {
			continue
		}

		task := PreviewTask{}

		// Content (task text)
		if idx, ok := colIndex["CONTENT"]; ok && idx < len(record) {
			task.Text = strings.TrimSpace(record[idx])
		}

		// Skip empty tasks
		if task.Text == "" {
			continue
		}

		// Priority (Todoist: 1=urgent/high, 2=high, 3=medium, 4=normal/none)
		if idx, ok := colIndex["PRIORITY"]; ok && idx < len(record) {
			task.Priority = mapTodoistPriority(record[idx])
		}

		// Project (from INDENT hierarchy or PROJECT column if exists)
		if idx, ok := colIndex["PROJECT"]; ok && idx < len(record) {
			task.Project = strings.TrimSpace(record[idx])
		}

		// Due date
		if idx, ok := colIndex["DATE"]; ok && idx < len(record) {
			if dueDate := parseTodoistDate(record[idx]); dueDate != nil {
				task.DueDate = dueDate
			}
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// mapTodoistPriority converts Todoist priority to our priority system.
// Todoist: 1 = urgent (highest), 2 = high, 3 = medium, 4 = normal (lowest)
// Our system: high, medium, low, none
func mapTodoistPriority(priority string) storage.Priority {
	switch strings.TrimSpace(priority) {
	case "1":
		return storage.PriorityHigh
	case "2":
		return storage.PriorityHigh
	case "3":
		return storage.PriorityMedium
	case "4":
		return storage.PriorityLow
	default:
		return storage.PriorityNone
	}
}

// parseTodoistDate parses various Todoist date formats.
func parseTodoistDate(dateStr string) *time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return nil
	}

	// Try various formats
	formats := []string{
		"2006-01-02",
		"Jan 2 2006",
		"Jan 2, 2006",
		"2 Jan 2006",
		"January 2, 2006",
		"01/02/2006",
		"02/01/2006",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateStr, time.Local); err == nil {
			return &t
		}
	}

	return nil
}
