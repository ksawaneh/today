// Package importer provides import functionality for the today app.
// This file implements Taskwarrior JSON import.
package importer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"today/internal/storage"
)

// TaskwarriorImporter handles importing from Taskwarrior JSON exports.
type TaskwarriorImporter struct{}

// taskwarriorTask represents a task in Taskwarrior's JSON format.
type taskwarriorTask struct {
	Description string `json:"description"`
	Status      string `json:"status"`
	Project     string `json:"project"`
	Priority    string `json:"priority"`
	Due         string `json:"due"`
	Entry       string `json:"entry"`
	End         string `json:"end"`
	UUID        string `json:"uuid"`
}

// Name returns the importer name.
func (t *TaskwarriorImporter) Name() string {
	return "taskwarrior"
}

// Import reads tasks from Taskwarrior JSON and adds them to storage.
func (t *TaskwarriorImporter) Import(reader io.Reader, store *storage.Storage) (*ImportResult, error) {
	tasks, err := t.parseTasks(reader)
	if err != nil {
		return nil, err
	}

	result := &ImportResult{}

	for _, task := range tasks {
		addedTask, err := store.AddTask(task.Text, task.Project, task.Priority, task.DueDate)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", task.Text, err))
			continue
		}

		// Mark as complete if it was completed in Taskwarrior
		if task.Done {
			if err := store.CompleteTask(addedTask.ID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to mark %s as complete: %v", task.Text, err))
			}
		}

		result.Imported++
	}

	return result, nil
}

// Preview returns a list of tasks that would be imported.
func (t *TaskwarriorImporter) Preview(reader io.Reader) ([]PreviewTask, error) {
	return t.parseTasks(reader)
}

// parseTasks reads and parses Taskwarrior JSON format.
// Supports both JSON array format and newline-delimited JSON (NDJSON).
func (t *TaskwarriorImporter) parseTasks(reader io.Reader) ([]PreviewTask, error) {
	br := bufio.NewReader(reader)
	prefix, first, err := readFirstNonSpaceByte(br)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty input")
		}
		return nil, fmt.Errorf("failed to read input: %w", err)
	}

	r := io.MultiReader(bytes.NewReader(prefix), br)
	if first == '[' {
		return parseTaskwarriorJSONArray(r)
	}
	return parseTaskwarriorNDJSON(r)
}

const maxTaskwarriorNDJSONLineBytes = 4 << 20 // 4MiB

func readFirstNonSpaceByte(r *bufio.Reader) ([]byte, byte, error) {
	var prefix []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF && len(prefix) == 0 {
				return nil, 0, io.EOF
			}
			return prefix, 0, err
		}
		prefix = append(prefix, b)
		if !isSpaceByte(b) {
			return prefix, b, nil
		}
	}
}

func isSpaceByte(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}

func parseTaskwarriorJSONArray(r io.Reader) ([]PreviewTask, error) {
	dec := json.NewDecoder(r)
	tok, err := dec.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON array: %w", err)
	}
	if d, ok := tok.(json.Delim); !ok || d != '[' {
		return nil, fmt.Errorf("failed to parse JSON array: expected '['")
	}

	var tasks []PreviewTask
	var idx int
	for dec.More() {
		idx++
		var tw taskwarriorTask
		if err := dec.Decode(&tw); err != nil {
			return nil, fmt.Errorf("failed to decode task %d: %w", idx, err)
		}
		if task, ok := previewFromTaskwarrior(tw); ok {
			tasks = append(tasks, task)
		}
	}

	// Consume closing ']'
	if _, err := dec.Token(); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array: %w", err)
	}

	return tasks, nil
}

func parseTaskwarriorNDJSON(r io.Reader) ([]PreviewTask, error) {
	br := bufio.NewReader(r)
	var tasks []PreviewTask
	var lineNo int
	for {
		line, err := br.ReadBytes('\n')
		if len(line) > maxTaskwarriorNDJSONLineBytes {
			return nil, fmt.Errorf("taskwarrior NDJSON line %d exceeds %d bytes", lineNo+1, maxTaskwarriorNDJSONLineBytes)
		}
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to read NDJSON: %w", err)
		}
		if len(line) == 0 && err == io.EOF {
			break
		}

		lineNo++
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			if err == io.EOF {
				break
			}
			continue
		}

		var tw taskwarriorTask
		if uerr := json.Unmarshal(line, &tw); uerr != nil {
			return nil, fmt.Errorf("invalid JSON on line %d: %w", lineNo, uerr)
		}
		if task, ok := previewFromTaskwarrior(tw); ok {
			tasks = append(tasks, task)
		}

		if err == io.EOF {
			break
		}
	}

	if lineNo == 0 {
		return nil, fmt.Errorf("empty input")
	}

	return tasks, nil
}

func previewFromTaskwarrior(tw taskwarriorTask) (PreviewTask, bool) {
	// Skip deleted tasks
	if tw.Status == "deleted" {
		return PreviewTask{}, false
	}

	task := PreviewTask{
		Text:     tw.Description,
		Project:  tw.Project,
		Priority: mapTaskwarriorPriority(tw.Priority),
		Done:     tw.Status == "completed",
	}

	// Skip empty tasks
	if strings.TrimSpace(task.Text) == "" {
		return PreviewTask{}, false
	}
	task.Text = strings.TrimSpace(task.Text)

	// Parse due date
	if tw.Due != "" {
		if dueDate := parseTaskwarriorDate(tw.Due); dueDate != nil {
			task.DueDate = dueDate
		}
	}

	return task, true
}

// mapTaskwarriorPriority converts Taskwarrior priority to our system.
// Taskwarrior: H = high, M = medium, L = low
func mapTaskwarriorPriority(priority string) storage.Priority {
	switch strings.ToUpper(strings.TrimSpace(priority)) {
	case "H":
		return storage.PriorityHigh
	case "M":
		return storage.PriorityMedium
	case "L":
		return storage.PriorityLow
	default:
		return storage.PriorityNone
	}
}

// parseTaskwarriorDate parses Taskwarrior's date format.
// Format: 20140928T211124Z (ISO 8601 basic format)
func parseTaskwarriorDate(dateStr string) *time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return nil
	}

	// Taskwarrior format: 20140928T211124Z
	formats := []string{
		"20060102T150405Z",
		"20060102T150405",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			// Convert to local time
			localTime := t.Local()
			return &localTime
		}
	}

	return nil
}
