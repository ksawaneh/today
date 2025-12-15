// Package backup provides backup and restore functionality for the today app.
// It manages timestamped backups of all data files (tasks, habits, timer).
package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"today/internal/fsutil"
)

// Version constants for the backup format.
const (
	ManifestVersion = "1.0"
	ManifestFile    = "manifest.json"
	BackupsDir      = "backups"
)

// Data files that are backed up.
var dataFiles = []string{"tasks.json", "habits.json", "timer.json"}

// Manager handles backup and restore operations.
type Manager struct {
	dataDir    string // Path to data directory (e.g., ~/.today)
	backupDir  string // Path to backups directory (e.g., ~/.today/backups)
	appVersion string // Application version for manifest
}

// Manifest contains metadata about a backup.
type Manifest struct {
	Version    string         `json:"version"`
	CreatedAt  time.Time      `json:"created_at"`
	AppVersion string         `json:"app_version"`
	Files      []string       `json:"files"`
	Stats      map[string]int `json:"stats"`
}

// BackupInfo contains summary information about a backup.
type BackupInfo struct {
	Name      string         // Directory name (2025-12-15_143022)
	Path      string         // Full path to backup directory
	CreatedAt time.Time      // When the backup was created
	Stats     map[string]int // Statistics (tasks, habits, timer_entries)
}

// NewManager creates a new backup manager.
func NewManager(dataDir, appVersion string) *Manager {
	return &Manager{
		dataDir:    dataDir,
		backupDir:  filepath.Join(dataDir, BackupsDir),
		appVersion: appVersion,
	}
}

// Create creates a new backup of all data files.
// Returns the backup name (timestamp format) on success.
func (m *Manager) Create() (string, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(m.backupDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup name from current timestamp (with milliseconds for uniqueness)
	now := time.Now()
	name := fmt.Sprintf("%s_%03d", now.Format("2006-01-02_150405"), now.Nanosecond()/1e6)
	backupPath := filepath.Join(m.backupDir, name)

	// Create the backup directory
	if err := os.MkdirAll(backupPath, 0700); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	// Copy data files
	var copiedFiles []string
	stats := make(map[string]int)

	for _, filename := range dataFiles {
		srcPath := filepath.Join(m.dataDir, filename)
		dstPath := filepath.Join(backupPath, filename)

		// Skip if source file doesn't exist
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}

		// Copy the file (private + atomic)
		if err := copyFileAtomic(srcPath, dstPath); err != nil {
			// Clean up on failure
			_ = os.RemoveAll(backupPath)
			return "", fmt.Errorf("failed to copy %s: %w", filename, err)
		}

		copiedFiles = append(copiedFiles, filename)

		// Gather stats
		count, err := countItems(srcPath, filename)
		if err == nil {
			stats[statsKeyForFile(filename)] = count
		}
	}

	// Create manifest
	manifest := Manifest{
		Version:    ManifestVersion,
		CreatedAt:  now,
		AppVersion: m.appVersion,
		Files:      copiedFiles,
		Stats:      stats,
	}

	manifestPath := filepath.Join(backupPath, ManifestFile)
	if err := writeJSON(manifestPath, manifest); err != nil {
		_ = os.RemoveAll(backupPath)
		return "", fmt.Errorf("failed to write manifest: %w", err)
	}

	return name, nil
}

// List returns all available backups, sorted by creation time (newest first).
func (m *Manager) List() ([]BackupInfo, error) {
	// Check if backup directory exists
	if _, err := os.Stat(m.backupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil
	}

	entries, err := os.ReadDir(m.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		backupPath := filepath.Join(m.backupDir, entry.Name())
		manifestPath := filepath.Join(backupPath, ManifestFile)

		// Try to read manifest
		var manifest Manifest
		if err := readJSON(manifestPath, &manifest); err != nil {
			// Try to parse timestamp from directory name
			createdAt, parseErr := parseBackupName(entry.Name())
			if parseErr != nil {
				continue // Skip invalid backups
			}
			manifest.CreatedAt = createdAt
			manifest.Stats = make(map[string]int)
		}

		backups = append(backups, BackupInfo{
			Name:      entry.Name(),
			Path:      backupPath,
			CreatedAt: manifest.CreatedAt,
			Stats:     manifest.Stats,
		})
	}

	// Sort by creation time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// Restore restores data from a specific backup.
// It creates a safety backup before restoring.
func (m *Manager) Restore(name string) error {
	if err := validateBackupName(name); err != nil {
		return err
	}

	backupPath := filepath.Join(m.backupDir, name)

	// Validate backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", name)
	}

	// Read manifest to know which files to restore
	manifestPath := filepath.Join(backupPath, ManifestFile)
	var manifest Manifest
	if err := readJSON(manifestPath, &manifest); err != nil {
		// Fall back to default file list if manifest is missing
		manifest.Files = dataFiles
	}

	// Create safety backup first
	safetyName, err := m.Create()
	if err != nil {
		return fmt.Errorf("failed to create safety backup: %w", err)
	}

	// Restore files
	for _, filename := range manifest.Files {
		srcPath := filepath.Join(backupPath, filename)
		dstPath := filepath.Join(m.dataDir, filename)

		// Skip if backup file doesn't exist
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}

		// Copy file (overwrite existing, atomically)
		if err := copyFileAtomic(srcPath, dstPath); err != nil {
			return fmt.Errorf("failed to restore %s (safety backup: %s): %w", filename, safetyName, err)
		}
	}

	// Validate restored files
	for _, filename := range manifest.Files {
		dstPath := filepath.Join(m.dataDir, filename)
		if err := validateJSON(dstPath); err != nil {
			return fmt.Errorf("restored file %s is invalid (safety backup: %s): %w", filename, safetyName, err)
		}
	}

	return nil
}

// RestoreLatest restores from the most recent backup.
func (m *Manager) RestoreLatest() error {
	backups, err := m.List()
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		return fmt.Errorf("no backups available")
	}

	return m.Restore(backups[0].Name)
}

// Delete removes a specific backup.
func (m *Manager) Delete(name string) error {
	if err := validateBackupName(name); err != nil {
		return err
	}

	backupPath := filepath.Join(m.backupDir, name)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", name)
	}

	return os.RemoveAll(backupPath)
}

// Prune removes old backups, keeping only the N most recent.
func (m *Manager) Prune(keepCount int) (int, error) {
	if keepCount < 0 {
		return 0, fmt.Errorf("keepCount must be non-negative")
	}

	backups, err := m.List()
	if err != nil {
		return 0, err
	}

	if len(backups) <= keepCount {
		return 0, nil
	}

	deleted := 0
	for _, backup := range backups[keepCount:] {
		if err := m.Delete(backup.Name); err != nil {
			return deleted, err
		}
		deleted++
	}

	return deleted, nil
}

// GetBackup returns information about a specific backup.
func (m *Manager) GetBackup(name string) (*BackupInfo, error) {
	if err := validateBackupName(name); err != nil {
		return nil, err
	}

	backupPath := filepath.Join(m.backupDir, name)

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("backup not found: %s", name)
	}

	manifestPath := filepath.Join(backupPath, ManifestFile)
	var manifest Manifest
	if err := readJSON(manifestPath, &manifest); err != nil {
		// Try to parse timestamp from directory name
		createdAt, parseErr := parseBackupName(name)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid backup: %s", name)
		}
		manifest.CreatedAt = createdAt
		manifest.Stats = make(map[string]int)
	}

	return &BackupInfo{
		Name:      name,
		Path:      backupPath,
		CreatedAt: manifest.CreatedAt,
		Stats:     manifest.Stats,
	}, nil
}

// Helper functions

func validateBackupName(name string) error {
	if name == "" {
		return fmt.Errorf("backup name is required")
	}
	if name != filepath.Base(name) {
		return fmt.Errorf("invalid backup name: %q", name)
	}
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("invalid backup name: %q", name)
	}
	if _, err := parseBackupName(name); err != nil {
		return fmt.Errorf("invalid backup name: %q", name)
	}
	return nil
}

func copyFileAtomic(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(dst, data, 0600)
}

// writeJSON writes a value as JSON to a file.
func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return fsutil.WriteFileAtomic(path, data, 0600)
}

// readJSON reads JSON from a file into a value.
func readJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// validateJSON checks that a file contains valid JSON.
func validateJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Missing file is OK
		}
		return err
	}

	var v interface{}
	return json.Unmarshal(data, &v)
}

// countItems counts items in a data file.
func countItems(path, filename string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return 0, err
	}

	switch filename {
	case "tasks.json":
		if tasks, ok := result["tasks"].([]interface{}); ok {
			return len(tasks), nil
		}
	case "habits.json":
		if habits, ok := result["habits"].([]interface{}); ok {
			return len(habits), nil
		}
	case "timer.json":
		if entries, ok := result["entries"].([]interface{}); ok {
			return len(entries), nil
		}
	}

	return 0, nil
}

// statsKeyForFile returns the stats key for a given filename.
func statsKeyForFile(filename string) string {
	switch filename {
	case "tasks.json":
		return "tasks"
	case "habits.json":
		return "habits"
	case "timer.json":
		return "timer_entries"
	default:
		return filename
	}
}

// parseBackupName parses a backup directory name into a timestamp.
// Supports both old format (2006-01-02_150405) and new format (2006-01-02_150405_XXX).
func parseBackupName(name string) (time.Time, error) {
	// Try new format with milliseconds first
	if len(name) == 21 {
		// Format: 2006-01-02_150405_XXX
		baseTime, err := time.Parse("2006-01-02_150405", name[:17])
		if err != nil {
			return time.Time{}, err
		}
		if name[17] != '_' {
			return time.Time{}, fmt.Errorf("invalid backup format")
		}
		ms, err := strconv.Atoi(name[18:])
		if err != nil || ms < 0 || ms > 999 {
			return time.Time{}, fmt.Errorf("invalid milliseconds")
		}
		return baseTime.Add(time.Duration(ms) * time.Millisecond), nil
	}

	// Try old format without milliseconds
	return time.Parse("2006-01-02_150405", name)
}
