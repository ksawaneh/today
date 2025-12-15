// Package backup provides backup and restore functionality for the today app.
// This file contains tests for the backup module.
package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// createTestData creates sample data files for testing.
func createTestData(t *testing.T, dataDir string) {
	t.Helper()

	// Create tasks.json
	tasks := map[string]interface{}{
		"tasks": []map[string]interface{}{
			{"id": "t_1", "text": "Task 1", "done": false},
			{"id": "t_2", "text": "Task 2", "done": true},
		},
	}
	writeTestJSON(t, filepath.Join(dataDir, "tasks.json"), tasks)

	// Create habits.json
	habits := map[string]interface{}{
		"habits": []map[string]interface{}{
			{"id": "h_1", "name": "Exercise", "icon": "🏃"},
		},
		"logs": []map[string]interface{}{
			{"habit_id": "h_1", "date": "2025-12-15"},
		},
	}
	writeTestJSON(t, filepath.Join(dataDir, "habits.json"), habits)

	// Create timer.json
	timer := map[string]interface{}{
		"current": nil,
		"entries": []map[string]interface{}{
			{
				"project":    "test-project",
				"started_at": "2025-12-15T10:00:00Z",
				"ended_at":   "2025-12-15T11:00:00Z",
			},
		},
	}
	writeTestJSON(t, filepath.Join(dataDir, "timer.json"), timer)
}

// writeTestJSON writes JSON to a file for testing.
func writeTestJSON(t *testing.T, path string, v interface{}) {
	t.Helper()

	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}

// readTestJSON reads JSON from a file for testing.
func readTestJSON(t *testing.T, path string) map[string]interface{} {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	return result
}

// TestManager_Create tests backup creation.
func TestManager_Create(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	createTestData(t, tmpDir)

	manager := NewManager(tmpDir, "1.2.0-test")

	// Create backup
	name, err := manager.Create()
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Verify backup name format (2006-01-02_150405_XXX where XXX is milliseconds)
	if len(name) != 21 { // "2006-01-02_150405_XXX"
		t.Errorf("Expected backup name length 21, got %d: %s", len(name), name)
	}

	// Verify backup directory exists
	backupPath := filepath.Join(tmpDir, BackupsDir, name)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Errorf("Backup directory not created: %s", backupPath)
	}

	// Verify files were copied
	for _, filename := range dataFiles {
		filePath := filepath.Join(backupPath, filename)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("File not backed up: %s", filename)
		}
	}

	// Verify manifest
	manifestPath := filepath.Join(backupPath, ManifestFile)
	manifest := readTestJSON(t, manifestPath)

	if manifest["version"] != ManifestVersion {
		t.Errorf("Expected manifest version %s, got %v", ManifestVersion, manifest["version"])
	}

	if manifest["app_version"] != "1.2.0-test" {
		t.Errorf("Expected app_version 1.2.0-test, got %v", manifest["app_version"])
	}

	// Verify stats
	stats, ok := manifest["stats"].(map[string]interface{})
	if !ok {
		t.Fatal("Stats not found in manifest")
	}

	if int(stats["tasks"].(float64)) != 2 {
		t.Errorf("Expected 2 tasks, got %v", stats["tasks"])
	}

	if int(stats["habits"].(float64)) != 1 {
		t.Errorf("Expected 1 habit, got %v", stats["habits"])
	}

	if int(stats["timer_entries"].(float64)) != 1 {
		t.Errorf("Expected 1 timer entry, got %v", stats["timer_entries"])
	}
}

// TestManager_List tests listing backups.
func TestManager_List(t *testing.T) {
	tmpDir := t.TempDir()
	createTestData(t, tmpDir)

	manager := NewManager(tmpDir, "1.0.0")

	// No backups initially
	backups, err := manager.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups, got %d", len(backups))
	}

	// Create some backups
	name1, _ := manager.Create()
	time.Sleep(10 * time.Millisecond)
	name2, _ := manager.Create()

	// List should return both, newest first
	backups, err = manager.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if len(backups) != 2 {
		t.Fatalf("Expected 2 backups, got %d", len(backups))
	}

	// Newest should be first
	if backups[0].Name != name2 {
		t.Errorf("Expected newest backup %s first, got %s", name2, backups[0].Name)
	}

	if backups[1].Name != name1 {
		t.Errorf("Expected older backup %s second, got %s", name1, backups[1].Name)
	}
}

// TestManager_Restore tests restoring from a backup.
func TestManager_Restore(t *testing.T) {
	tmpDir := t.TempDir()
	createTestData(t, tmpDir)

	manager := NewManager(tmpDir, "1.0.0")

	// Create backup
	name, err := manager.Create()
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Modify original data
	tasks := map[string]interface{}{
		"tasks": []map[string]interface{}{
			{"id": "t_new", "text": "New Task", "done": false},
		},
	}
	writeTestJSON(t, filepath.Join(tmpDir, "tasks.json"), tasks)

	// Verify modification
	modified := readTestJSON(t, filepath.Join(tmpDir, "tasks.json"))
	modifiedTasks := modified["tasks"].([]interface{})
	if len(modifiedTasks) != 1 {
		t.Fatalf("Expected 1 task after modification, got %d", len(modifiedTasks))
	}

	// Restore
	if err := manager.Restore(name); err != nil {
		t.Fatalf("Restore() error: %v", err)
	}

	// Verify restoration
	restored := readTestJSON(t, filepath.Join(tmpDir, "tasks.json"))
	restoredTasks := restored["tasks"].([]interface{})
	if len(restoredTasks) != 2 {
		t.Errorf("Expected 2 tasks after restore, got %d", len(restoredTasks))
	}
}

// TestManager_RestoreLatest tests restoring the most recent backup.
func TestManager_RestoreLatest(t *testing.T) {
	tmpDir := t.TempDir()
	createTestData(t, tmpDir)

	manager := NewManager(tmpDir, "1.0.0")

	// Create first backup
	_, err := manager.Create()
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Modify data
	tasks := map[string]interface{}{
		"tasks": []map[string]interface{}{
			{"id": "t_modified", "text": "Modified Task", "done": false},
		},
	}
	writeTestJSON(t, filepath.Join(tmpDir, "tasks.json"), tasks)

	// Create second backup (with modified data)
	time.Sleep(10 * time.Millisecond)
	_, err = manager.Create()
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Modify again
	tasks = map[string]interface{}{
		"tasks": []map[string]interface{}{
			{"id": "t_final", "text": "Final Task", "done": false},
		},
	}
	writeTestJSON(t, filepath.Join(tmpDir, "tasks.json"), tasks)

	// Restore latest (should restore the second backup with "Modified Task")
	if err := manager.RestoreLatest(); err != nil {
		t.Fatalf("RestoreLatest() error: %v", err)
	}

	// Verify restoration
	restored := readTestJSON(t, filepath.Join(tmpDir, "tasks.json"))
	restoredTasks := restored["tasks"].([]interface{})
	if len(restoredTasks) != 1 {
		t.Fatalf("Expected 1 task after restore, got %d", len(restoredTasks))
	}

	firstTask := restoredTasks[0].(map[string]interface{})
	if firstTask["id"] != "t_modified" {
		t.Errorf("Expected restored task id 't_modified', got %v", firstTask["id"])
	}
}

// TestManager_RestoreNonexistent tests restoring a nonexistent backup.
func TestManager_RestoreNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	createTestData(t, tmpDir)

	manager := NewManager(tmpDir, "1.0.0")

	err := manager.Restore("nonexistent-backup")
	if err == nil {
		t.Error("Expected error when restoring nonexistent backup")
	}
}

// TestManager_Delete tests deleting a backup.
func TestManager_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	createTestData(t, tmpDir)

	manager := NewManager(tmpDir, "1.0.0")

	// Create backup
	name, err := manager.Create()
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Delete backup
	if err := manager.Delete(name); err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	// Verify deletion
	backups, _ := manager.List()
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups after delete, got %d", len(backups))
	}
}

// TestManager_Prune tests pruning old backups.
func TestManager_Prune(t *testing.T) {
	tmpDir := t.TempDir()
	createTestData(t, tmpDir)

	manager := NewManager(tmpDir, "1.0.0")

	// Create 5 backups
	for i := 0; i < 5; i++ {
		_, err := manager.Create()
		if err != nil {
			t.Fatalf("Create() error: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Prune, keeping only 2
	deleted, err := manager.Prune(2)
	if err != nil {
		t.Fatalf("Prune() error: %v", err)
	}

	if deleted != 3 {
		t.Errorf("Expected 3 deleted, got %d", deleted)
	}

	// Verify only 2 remain
	backups, _ := manager.List()
	if len(backups) != 2 {
		t.Errorf("Expected 2 backups after prune, got %d", len(backups))
	}
}

// TestManager_CreateWithEmptyData tests creating a backup with no data files.
func TestManager_CreateWithEmptyData(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create any data files
	manager := NewManager(tmpDir, "1.0.0")

	// Should still create a backup (with empty file list)
	name, err := manager.Create()
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Verify backup was created
	info, err := manager.GetBackup(name)
	if err != nil {
		t.Fatalf("GetBackup() error: %v", err)
	}

	if info.Name != name {
		t.Errorf("Expected backup name %s, got %s", name, info.Name)
	}
}

// TestManager_GetBackup tests getting info about a specific backup.
func TestManager_GetBackup(t *testing.T) {
	tmpDir := t.TempDir()
	createTestData(t, tmpDir)

	manager := NewManager(tmpDir, "1.0.0")

	// Create backup
	name, err := manager.Create()
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Get backup info
	info, err := manager.GetBackup(name)
	if err != nil {
		t.Fatalf("GetBackup() error: %v", err)
	}

	if info.Name != name {
		t.Errorf("Expected name %s, got %s", name, info.Name)
	}

	if info.Stats["tasks"] != 2 {
		t.Errorf("Expected 2 tasks, got %d", info.Stats["tasks"])
	}

	// Get nonexistent backup
	_, err = manager.GetBackup("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent backup")
	}
}

// TestManager_RestoreCreateseSafetyBackup tests that restore creates a safety backup.
func TestManager_RestoreCreatesSafetyBackup(t *testing.T) {
	tmpDir := t.TempDir()
	createTestData(t, tmpDir)

	manager := NewManager(tmpDir, "1.0.0")

	// Create initial backup
	name, err := manager.Create()
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Restore should create a safety backup
	if err := manager.Restore(name); err != nil {
		t.Fatalf("Restore() error: %v", err)
	}

	// Should now have at least 2 backups (original + safety)
	backups, _ := manager.List()
	if len(backups) < 2 {
		t.Errorf("Expected at least 2 backups (including safety backup), got %d", len(backups))
	}
}
