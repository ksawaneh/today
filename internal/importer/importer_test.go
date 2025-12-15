// Package importer provides import functionality for the today app.
// This file contains tests for the import functionality.
package importer

import (
	"fmt"
	"strings"
	"testing"

	"today/internal/storage"
)

// TestTodoist_ParseCSV tests parsing valid Todoist CSV.
func TestTodoist_ParseCSV(t *testing.T) {
	csv := `TYPE,CONTENT,PRIORITY,INDENT,AUTHOR,RESPONSIBLE,DATE,DATE_LANG,TIMEZONE
task,Buy groceries,4,1,,,2025-12-20,en,America/New_York
task,Review PR,1,1,,,,,
note,This is a note,4,1,,,,,
task,Call mom,3,1,,,,,`

	importer := &TodoistImporter{}
	tasks, err := importer.Preview(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Preview() error: %v", err)
	}

	// Should have 3 tasks (note is skipped)
	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// Check first task
	if tasks[0].Text != "Buy groceries" {
		t.Errorf("Expected 'Buy groceries', got %q", tasks[0].Text)
	}
	if tasks[0].DueDate == nil {
		t.Error("Expected due date for first task")
	}
}

// TestTodoist_PriorityMapping tests Todoist priority conversion.
func TestTodoist_PriorityMapping(t *testing.T) {
	tests := []struct {
		input    string
		expected storage.Priority
	}{
		{"1", storage.PriorityHigh},   // Urgent
		{"2", storage.PriorityHigh},   // High
		{"3", storage.PriorityMedium}, // Medium
		{"4", storage.PriorityLow},    // Normal
		{"", storage.PriorityNone},
		{"5", storage.PriorityNone},
	}

	for _, tc := range tests {
		result := mapTodoistPriority(tc.input)
		if result != tc.expected {
			t.Errorf("mapTodoistPriority(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// TestTodoist_DateParsing tests various date formats.
func TestTodoist_DateParsing(t *testing.T) {
	tests := []struct {
		input    string
		hasDate  bool
		expected string // YYYY-MM-DD format
	}{
		{"2025-12-20", true, "2025-12-20"},
		{"Jan 2 2025", true, "2025-01-02"},
		{"", false, ""},
		{"invalid", false, ""},
	}

	for _, tc := range tests {
		result := parseTodoistDate(tc.input)
		if tc.hasDate {
			if result == nil {
				t.Errorf("parseTodoistDate(%q) = nil, want date", tc.input)
			} else if result.Format("2006-01-02") != tc.expected {
				t.Errorf("parseTodoistDate(%q) = %s, want %s", tc.input, result.Format("2006-01-02"), tc.expected)
			}
		} else {
			if result != nil {
				t.Errorf("parseTodoistDate(%q) = %v, want nil", tc.input, result)
			}
		}
	}
}

// TestTodoist_EmptyFile tests handling of empty CSV.
func TestTodoist_EmptyFile(t *testing.T) {
	importer := &TodoistImporter{}
	_, err := importer.Preview(strings.NewReader(""))
	if err == nil {
		t.Error("Expected error for empty CSV")
	}
}

// TestTodoist_MissingColumns tests handling of missing required columns.
func TestTodoist_MissingColumns(t *testing.T) {
	csv := `CONTENT,PRIORITY
Buy groceries,4`

	importer := &TodoistImporter{}
	_, err := importer.Preview(strings.NewReader(csv))
	if err == nil {
		t.Error("Expected error for missing TYPE column")
	}
}

func TestTodoist_HeaderBOM(t *testing.T) {
	csv := "\ufeffTYPE,CONTENT,PRIORITY\n" +
		"task,With BOM,4\n"

	importer := &TodoistImporter{}
	tasks, err := importer.Preview(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Preview() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Text != "With BOM" {
		t.Errorf("Expected 'With BOM', got %q", tasks[0].Text)
	}
}

func TestTodoist_RaggedRows(t *testing.T) {
	csv := `TYPE,CONTENT,PRIORITY
task,One,4,EXTRA,EXTRA2
task,Two,1`

	importer := &TodoistImporter{}
	tasks, err := importer.Preview(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("Preview() error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("Expected 2 tasks, got %d", len(tasks))
	}
}

// TestTaskwarrior_ParseJSON tests parsing Taskwarrior JSON array.
func TestTaskwarrior_ParseJSON(t *testing.T) {
	json := `[
		{"description":"Buy milk","status":"pending","project":"Home","priority":"H"},
		{"description":"Review code","status":"completed","project":"Work"},
		{"description":"Deleted task","status":"deleted"}
	]`

	importer := &TaskwarriorImporter{}
	tasks, err := importer.Preview(strings.NewReader(json))
	if err != nil {
		t.Fatalf("Preview() error: %v", err)
	}

	// Should have 2 tasks (deleted is skipped)
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}

	// Check first task
	if tasks[0].Text != "Buy milk" {
		t.Errorf("Expected 'Buy milk', got %q", tasks[0].Text)
	}
	if tasks[0].Priority != storage.PriorityHigh {
		t.Errorf("Expected high priority, got %q", tasks[0].Priority)
	}
	if tasks[0].Project != "Home" {
		t.Errorf("Expected project 'Home', got %q", tasks[0].Project)
	}

	// Check completed task
	if !tasks[1].Done {
		t.Error("Expected second task to be done")
	}
}

// TestTaskwarrior_ParseNDJSON tests parsing newline-delimited JSON.
func TestTaskwarrior_ParseNDJSON(t *testing.T) {
	ndjson := `{"description":"Task 1","status":"pending"}
{"description":"Task 2","status":"pending","priority":"M"}
{"description":"Task 3","status":"completed"}`

	importer := &TaskwarriorImporter{}
	tasks, err := importer.Preview(strings.NewReader(ndjson))
	if err != nil {
		t.Fatalf("Preview() error: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// Check priority mapping
	if tasks[1].Priority != storage.PriorityMedium {
		t.Errorf("Expected medium priority, got %q", tasks[1].Priority)
	}
}

// TestTaskwarrior_StatusMapping tests status to done conversion.
func TestTaskwarrior_StatusMapping(t *testing.T) {
	json := `[
		{"description":"Pending","status":"pending"},
		{"description":"Completed","status":"completed"},
		{"description":"Waiting","status":"waiting"},
		{"description":"Deleted","status":"deleted"}
	]`

	importer := &TaskwarriorImporter{}
	tasks, err := importer.Preview(strings.NewReader(json))
	if err != nil {
		t.Fatalf("Preview() error: %v", err)
	}

	// Should have 3 tasks (deleted is skipped)
	if len(tasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(tasks))
	}

	// Check done status
	if tasks[0].Done {
		t.Error("Pending task should not be done")
	}
	if !tasks[1].Done {
		t.Error("Completed task should be done")
	}
	if tasks[2].Done {
		t.Error("Waiting task should not be done")
	}
}

// TestTaskwarrior_PriorityMapping tests Taskwarrior priority conversion.
func TestTaskwarrior_PriorityMapping(t *testing.T) {
	tests := []struct {
		input    string
		expected storage.Priority
	}{
		{"H", storage.PriorityHigh},
		{"h", storage.PriorityHigh},
		{"M", storage.PriorityMedium},
		{"m", storage.PriorityMedium},
		{"L", storage.PriorityLow},
		{"l", storage.PriorityLow},
		{"", storage.PriorityNone},
	}

	for _, tc := range tests {
		result := mapTaskwarriorPriority(tc.input)
		if result != tc.expected {
			t.Errorf("mapTaskwarriorPriority(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// TestTaskwarrior_DateParsing tests Taskwarrior date format parsing.
func TestTaskwarrior_DateParsing(t *testing.T) {
	tests := []struct {
		input   string
		hasDate bool
	}{
		{"20251220T000000Z", true},
		{"20251220T120000", true},
		{"2025-12-20T00:00:00Z", true},
		{"2025-12-20", true},
		{"", false},
		{"invalid", false},
	}

	for _, tc := range tests {
		result := parseTaskwarriorDate(tc.input)
		if tc.hasDate && result == nil {
			t.Errorf("parseTaskwarriorDate(%q) = nil, want date", tc.input)
		}
		if !tc.hasDate && result != nil {
			t.Errorf("parseTaskwarriorDate(%q) = %v, want nil", tc.input, result)
		}
	}
}

// TestTaskwarrior_EmptyInput tests handling of empty input.
func TestTaskwarrior_EmptyInput(t *testing.T) {
	importer := &TaskwarriorImporter{}
	_, err := importer.Preview(strings.NewReader(""))
	if err == nil {
		t.Error("Expected error for empty input")
	}
}

func TestTaskwarrior_LongNDJSONLine(t *testing.T) {
	desc := strings.Repeat("a", 70_000)
	ndjson := fmt.Sprintf("{\"description\":%q,\"status\":\"pending\"}\n", desc)

	importer := &TaskwarriorImporter{}
	tasks, err := importer.Preview(strings.NewReader(ndjson))
	if err != nil {
		t.Fatalf("Preview() error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Text != desc {
		t.Errorf("Unexpected task text length: got %d, want %d", len(tasks[0].Text), len(desc))
	}
}

func TestTaskwarrior_InvalidNDJSONReturnsError(t *testing.T) {
	ndjson := `{"description":"Task 1","status":"pending"}
{invalid json}
{"description":"Task 2","status":"pending"}`

	importer := &TaskwarriorImporter{}
	_, err := importer.Preview(strings.NewReader(ndjson))
	if err == nil {
		t.Fatal("Expected error for invalid NDJSON")
	}
}

// TestGetImporter tests the importer factory function.
func TestGetImporter(t *testing.T) {
	tests := []struct {
		format   string
		expected string
	}{
		{"todoist", "todoist"},
		{"taskwarrior", "taskwarrior"},
		{"unknown", ""},
	}

	for _, tc := range tests {
		importer := GetImporter(tc.format)
		if tc.expected == "" {
			if importer != nil {
				t.Errorf("GetImporter(%q) should return nil", tc.format)
			}
		} else {
			if importer == nil {
				t.Errorf("GetImporter(%q) should not return nil", tc.format)
			} else if importer.Name() != tc.expected {
				t.Errorf("GetImporter(%q).Name() = %q, want %q", tc.format, importer.Name(), tc.expected)
			}
		}
	}
}

// TestSupportedFormats tests the supported formats list.
func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	if len(formats) < 2 {
		t.Errorf("Expected at least 2 formats, got %d", len(formats))
	}

	// Check that todoist and taskwarrior are included
	found := map[string]bool{"todoist": false, "taskwarrior": false}
	for _, f := range formats {
		found[f] = true
	}

	for format, ok := range found {
		if !ok {
			t.Errorf("Expected %q in supported formats", format)
		}
	}
}

// TestImport_Integration tests actual import to storage.
func TestImport_Integration(t *testing.T) {
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Import from Todoist
	csv := `TYPE,CONTENT,PRIORITY,INDENT,AUTHOR,RESPONSIBLE,DATE,DATE_LANG,TIMEZONE
task,Test task 1,1,1,,,,,
task,Test task 2,4,1,,,,,`

	importer := &TodoistImporter{}
	result, err := importer.Import(strings.NewReader(csv), store)
	if err != nil {
		t.Fatalf("Import() error: %v", err)
	}

	if result.Imported != 2 {
		t.Errorf("Expected 2 imported, got %d", result.Imported)
	}

	// Verify tasks exist in storage
	tasks, err := store.LoadTasks()
	if err != nil {
		t.Fatalf("LoadTasks() error: %v", err)
	}

	if len(tasks.Tasks) != 2 {
		t.Errorf("Expected 2 tasks in storage, got %d", len(tasks.Tasks))
	}
}
