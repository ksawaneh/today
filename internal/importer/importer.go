// Package importer provides import functionality for migrating tasks from
// other productivity tools like Todoist and Taskwarrior.
package importer

import (
	"io"
	"time"

	"today/internal/storage"
)

// ImportResult contains statistics about an import operation.
type ImportResult struct {
	Imported int      // Number of successfully imported tasks
	Skipped  int      // Number of skipped items (duplicates, notes, etc.)
	Errors   []string // Error messages for failed imports
}

// PreviewTask represents a task preview before import.
type PreviewTask struct {
	Text     string
	Project  string
	Priority storage.Priority
	DueDate  *time.Time
	Done     bool
}

// Importer defines the interface for import implementations.
type Importer interface {
	// Import reads tasks from the reader and adds them to storage.
	Import(reader io.Reader, store *storage.Storage) (*ImportResult, error)

	// Preview reads tasks from the reader without importing.
	Preview(reader io.Reader) ([]PreviewTask, error)

	// Name returns the importer name (e.g., "todoist", "taskwarrior").
	Name() string
}

// GetImporter returns the appropriate importer for the given format.
func GetImporter(format string) Importer {
	switch format {
	case "todoist":
		return &TodoistImporter{}
	case "taskwarrior":
		return &TaskwarriorImporter{}
	default:
		return nil
	}
}

// SupportedFormats returns the list of supported import formats.
func SupportedFormats() []string {
	return []string{"todoist", "taskwarrior"}
}
