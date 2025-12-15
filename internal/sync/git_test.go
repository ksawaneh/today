// Package sync provides git synchronization for the today app data directory.
// This file contains tests for git sync functionality.
package sync

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// skipIfNoGit skips the test if git is not installed.
func skipIfNoGit(t *testing.T) {
	t.Helper()
	if !IsGitInstalled() {
		t.Skip("git not installed")
	}
}

// createTestDir creates a temporary directory for testing.
func createTestDir(t *testing.T) string {
	t.Helper()

	// Avoid mutating the developer's global git config during tests.
	// These env vars override git config for commits made by this process.
	t.Setenv("GIT_AUTHOR_NAME", "Test User")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "Test User")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")

	return t.TempDir()
}

// TestIsGitInstalled tests git installation detection.
func TestIsGitInstalled(t *testing.T) {
	// This test just verifies the function runs without panic.
	// Result depends on the test environment.
	_ = IsGitInstalled()
}

// TestGitSync_Init tests repository initialization.
func TestGitSync_Init(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true, AutoCommit: true}
	gs := New(dir, cfg)

	// Should not be a repo initially
	if gs.IsRepo() {
		t.Error("Expected IsRepo() to return false before init")
	}

	// Initialize
	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	// Should be a repo now
	if !gs.IsRepo() {
		t.Error("Expected IsRepo() to return true after init")
	}

	// .gitignore should exist
	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error("Expected .gitignore to be created")
	}

	// .gitignore should have correct content
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("Failed to read .gitignore: %v", err)
	}

	expectedPatterns := []string{"backups/", "*.bak", "*.corrupt.*"}
	for _, pattern := range expectedPatterns {
		if !contains(string(content), pattern) {
			t.Errorf("Expected .gitignore to contain %q", pattern)
		}
	}
}

// TestGitSync_IsRepo tests repository detection.
func TestGitSync_IsRepo(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true}
	gs := New(dir, cfg)

	// Not a repo initially
	if gs.IsRepo() {
		t.Error("Expected IsRepo() to return false for non-repo")
	}

	// Create .git directory manually
	gitDir := filepath.Join(dir, ".git")
	if err := os.MkdirAll(gitDir, 0700); err != nil {
		t.Fatalf("Failed to create .git dir: %v", err)
	}

	// Should detect it now
	if !gs.IsRepo() {
		t.Error("Expected IsRepo() to return true after creating .git")
	}
}

// TestGitSync_Commit tests staging and committing files.
func TestGitSync_Commit(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true, AutoCommit: true, CommitMessage: "auto"}
	gs := New(dir, cfg)

	// Initialize repo
	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	// Create a test file
	testFile := "tasks.json"
	testPath := filepath.Join(dir, testFile)
	if err := os.WriteFile(testPath, []byte(`{"tasks":[]}`), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Commit it
	if err := gs.Commit([]string{testFile}); err != nil {
		t.Fatalf("Commit() error: %v", err)
	}

	// Verify commit was created
	output, err := gs.runGit("log", "--oneline")
	if err != nil {
		t.Fatalf("Failed to get git log: %v", err)
	}

	if !contains(output, "Update tasks") {
		t.Errorf("Expected commit message 'Update tasks', got: %s", output)
	}
}

// TestGitSync_CommitMultipleFiles tests committing multiple files.
func TestGitSync_CommitMultipleFiles(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true, AutoCommit: true, CommitMessage: "auto"}
	gs := New(dir, cfg)

	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	// Create multiple test files
	files := []string{"tasks.json", "habits.json"}
	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
			t.Fatalf("Failed to write %s: %v", f, err)
		}
	}

	// Commit them
	if err := gs.Commit(files); err != nil {
		t.Fatalf("Commit() error: %v", err)
	}

	// Verify commit message mentions multiple files
	output, err := gs.runGit("log", "--oneline", "-1")
	if err != nil {
		t.Fatalf("Failed to get git log: %v", err)
	}

	if !contains(output, "Update 2 files") {
		t.Errorf("Expected commit message 'Update 2 files', got: %s", output)
	}
}

// TestGitSync_CommitNoChanges tests committing when there are no changes.
func TestGitSync_CommitNoChanges(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true}
	gs := New(dir, cfg)

	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	// Commit without any files - should not error
	if err := gs.Commit([]string{}); err != nil {
		t.Errorf("Commit() with no files should not error: %v", err)
	}
}

// TestGitSync_Status tests status retrieval.
func TestGitSync_Status(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true}
	gs := New(dir, cfg)

	// Status of non-repo
	status, err := gs.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if status.IsRepo {
		t.Error("Expected IsRepo=false for non-repo")
	}

	// Initialize and check status
	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	status, err = gs.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	if !status.IsRepo {
		t.Error("Expected IsRepo=true after init")
	}

	// Should have a branch (master or main)
	if status.Branch == "" {
		t.Error("Expected Branch to be set")
	}

	// No remote configured
	if status.HasRemote {
		t.Error("Expected HasRemote=false without remote")
	}
}

// TestGitSync_StatusWithChanges tests status with uncommitted changes.
func TestGitSync_StatusWithChanges(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true}
	gs := New(dir, cfg)

	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	// Create an uncommitted file
	testPath := filepath.Join(dir, "test.json")
	if err := os.WriteFile(testPath, []byte(`{}`), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	status, err := gs.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	if !status.HasChanges {
		t.Error("Expected HasChanges=true with uncommitted file")
	}
}

// TestGitSync_Debounce tests that rapid saves result in a single commit.
func TestGitSync_Debounce(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true, AutoCommit: true}
	gs := New(dir, cfg)
	gs.debounceDuration = 100 * time.Millisecond // Short duration for testing

	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	// Create test files
	files := []string{"tasks.json", "habits.json", "timer.json"}
	for _, f := range files {
		path := filepath.Join(dir, f)
		if err := os.WriteFile(path, []byte(`{}`), 0600); err != nil {
			t.Fatalf("Failed to write %s: %v", f, err)
		}
	}

	// Trigger multiple saves rapidly
	for _, f := range files {
		gs.OnFileSaved(f)
	}

	// Wait for debounce to complete
	time.Sleep(200 * time.Millisecond)

	// Check commit count (should be 2: init + one combined commit)
	output, err := gs.runGit("rev-list", "--count", "HEAD")
	if err != nil {
		t.Fatalf("Failed to count commits: %v", err)
	}

	// Should have exactly 2 commits
	if trimOutput(output) != "2" {
		t.Errorf("Expected 2 commits (init + debounced), got: %s", output)
	}
}

// TestGitSync_OnFileSavedDisabled tests that OnFileSaved does nothing when disabled.
func TestGitSync_OnFileSavedDisabled(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: false, AutoCommit: true}
	gs := New(dir, cfg)

	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	// Create a test file
	testPath := filepath.Join(dir, "tasks.json")
	if err := os.WriteFile(testPath, []byte(`{}`), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Trigger save (should be ignored since disabled)
	gs.OnFileSaved("tasks.json")

	// Wait briefly
	time.Sleep(50 * time.Millisecond)

	// Should still have uncommitted changes
	status, err := gs.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	if !status.HasChanges {
		t.Error("Expected changes to remain uncommitted when sync is disabled")
	}
}

// TestGitSync_CustomCommitMessage tests custom commit messages.
func TestGitSync_CustomCommitMessage(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	customMsg := "Custom sync message"
	cfg := &Config{Enabled: true, CommitMessage: customMsg}
	gs := New(dir, cfg)

	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	// Create and commit a file
	testPath := filepath.Join(dir, "tasks.json")
	if err := os.WriteFile(testPath, []byte(`{}`), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	if err := gs.Commit([]string{"tasks.json"}); err != nil {
		t.Fatalf("Commit() error: %v", err)
	}

	// Verify custom message was used
	output, err := gs.runGit("log", "-1", "--format=%s")
	if err != nil {
		t.Fatalf("Failed to get commit message: %v", err)
	}

	if trimOutput(output) != customMsg {
		t.Errorf("Expected commit message %q, got: %s", customMsg, output)
	}
}

// TestGitSync_CommitNotARepo tests committing when not a repo.
func TestGitSync_CommitNotARepo(t *testing.T) {
	dir := createTestDir(t)
	cfg := &Config{Enabled: true}
	gs := New(dir, cfg)

	err := gs.Commit([]string{"tasks.json"})
	if err == nil {
		t.Error("Expected error when committing to non-repo")
	}
}

// TestGitSync_PullNoRemote tests pulling without a remote.
func TestGitSync_PullNoRemote(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true}
	gs := New(dir, cfg)

	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	err := gs.Pull()
	if err == nil {
		t.Error("Expected error when pulling without remote")
	}
}

// TestGitSync_PushNoRemote tests pushing without a remote.
func TestGitSync_PushNoRemote(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true}
	gs := New(dir, cfg)

	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	err := gs.Push()
	if err == nil {
		t.Error("Expected error when pushing without remote")
	}
}

// TestGitSync_Flush tests immediate flush of pending commits.
func TestGitSync_Flush(t *testing.T) {
	skipIfNoGit(t)

	dir := createTestDir(t)
	cfg := &Config{Enabled: true, AutoCommit: true}
	gs := New(dir, cfg)
	gs.debounceDuration = 10 * time.Second // Long duration

	if err := gs.Init(); err != nil {
		t.Fatalf("Init() error: %v", err)
	}

	// Create a test file
	testPath := filepath.Join(dir, "tasks.json")
	if err := os.WriteFile(testPath, []byte(`{}`), 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Trigger save
	gs.OnFileSaved("tasks.json")

	// Flush immediately (don't wait for debounce)
	gs.Flush()

	// Should be committed now
	status, err := gs.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	if status.HasChanges {
		t.Error("Expected no changes after Flush()")
	}
}

// TestGenerateCommitMessage tests commit message generation.
func TestGenerateCommitMessage(t *testing.T) {
	cfg := &Config{CommitMessage: "auto"}
	gs := New("", cfg)

	tests := []struct {
		files    []string
		expected string
	}{
		{[]string{"tasks.json"}, "Update tasks"},
		{[]string{"habits.json"}, "Update habits"},
		{[]string{"timer.json"}, "Update timer"},
		{[]string{"other.json"}, "Update other.json"},
		{[]string{"tasks.json", "habits.json"}, "Update 2 files"},
		{[]string{"a.json", "b.json", "c.json"}, "Update 3 files"},
	}

	for _, tc := range tests {
		result := gs.generateCommitMessage(tc.files)
		if result != tc.expected {
			t.Errorf("generateCommitMessage(%v) = %q, want %q", tc.files, result, tc.expected)
		}
	}
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
