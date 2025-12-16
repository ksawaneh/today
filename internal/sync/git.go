// Package sync provides git synchronization for the today app data directory.
// It handles automatic commits, pull, and push operations.
package sync

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	gosync "sync"
	"time"

	"today/internal/fsutil"
	"today/internal/storage"
)

// Config holds git sync configuration.
type Config struct {
	Enabled       bool   `yaml:"enabled"`
	AutoCommit    bool   `yaml:"auto_commit"`
	AutoPush      bool   `yaml:"auto_push"`
	PullOnStartup bool   `yaml:"pull_on_startup"`
	CommitMessage string `yaml:"commit_message"` // "auto" or custom template
}

// DefaultConfig returns the default sync configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:       false,
		AutoCommit:    true,
		AutoPush:      false,
		PullOnStartup: false,
		CommitMessage: "auto",
	}
}

// Status represents the current git status.
type Status struct {
	IsRepo       bool
	HasRemote    bool
	RemoteName   string
	RemoteURL    string
	Branch       string
	Ahead        int
	Behind       int
	HasChanges   bool
	LastCommitAt *time.Time
}

// GitSync manages git operations for the data directory.
type GitSync struct {
	dataDir string
	config  *Config

	// Debouncing for auto-commit
	pendingFiles    map[string]bool
	pendingContexts []storage.SaveContext // Semantic context for commit messages
	commitTimer     *time.Timer
	mu              gosync.Mutex

	// Serializes git operations to avoid index/lock conflicts.
	opMu gosync.Mutex

	// Debounce duration (configurable for testing)
	debounceDuration time.Duration
}

// New creates a new GitSync instance.
func New(dataDir string, cfg *Config) *GitSync {
	return &GitSync{
		dataDir:          dataDir,
		config:           cfg,
		pendingFiles:     make(map[string]bool),
		debounceDuration: 2 * time.Second,
	}
}

// IsGitInstalled checks if git is available on the system.
func IsGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsRepo checks if the data directory is a git repository.
func (g *GitSync) IsRepo() bool {
	gitDir := filepath.Join(g.dataDir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

const (
	defaultGitTimeout  = 10 * time.Second
	pullPushGitTimeout = 60 * time.Second
	commitGitTimeout   = 15 * time.Second
)

// Init initializes a git repository in the data directory.
func (g *GitSync) Init() error {
	g.opMu.Lock()
	defer g.opMu.Unlock()

	if !IsGitInstalled() {
		return fmt.Errorf("git is not installed")
	}

	// Initialize git repo
	if _, err := g.runGitTimeout(commitGitTimeout, "init"); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Create .gitignore
	gitignoreContent := `# today app - git sync ignore file
backups/
*.bak
*.corrupt.*
`
	gitignorePath := filepath.Join(g.dataDir, ".gitignore")
	if err := fsutil.WriteFileAtomic(gitignorePath, []byte(gitignoreContent), 0600); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	// Stage and commit .gitignore
	if _, err := g.runGitTimeout(defaultGitTimeout, "add", ".gitignore"); err != nil {
		return fmt.Errorf("failed to stage .gitignore: %w", err)
	}

	if _, err := g.runGitTimeout(commitGitTimeout, "-c", "commit.gpgsign=false", "commit", "-m", "Initialize today data repository"); err != nil {
		if !isGitNothingToCommit(err) {
			return fmt.Errorf("failed to create initial commit: %w", err)
		}
	}

	return nil
}

// Status returns the current git status.
func (g *GitSync) Status() (*Status, error) {
	g.opMu.Lock()
	defer g.opMu.Unlock()

	status := &Status{
		IsRepo: g.IsRepo(),
	}

	if !status.IsRepo {
		return status, nil
	}

	// Get current branch
	branch, err := g.runGitTimeout(defaultGitTimeout, "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		status.Branch = trimOutput(branch)
	}

	// Check for remote
	remotes, err := g.runGitTimeout(defaultGitTimeout, "remote", "-v")
	if err == nil && trimOutput(remotes) != "" {
		status.HasRemote = true
		// Parse remote name and URL (first line: "origin\tgit@...\t(fetch)")
		lines := bytes.Split([]byte(remotes), []byte("\n"))
		if len(lines) > 0 {
			parts := bytes.Fields(lines[0])
			if len(parts) >= 2 {
				status.RemoteName = string(parts[0])
				status.RemoteURL = string(parts[1])
			}
		}
	}

	// Check for uncommitted changes
	statusOutput, err := g.runGitTimeout(defaultGitTimeout, "status", "--porcelain")
	if err == nil {
		status.HasChanges = trimOutput(statusOutput) != ""
	}

	// Get ahead/behind count if there's a remote
	if status.HasRemote && status.Branch != "" {
		remote := status.RemoteName + "/" + status.Branch
		revList, err := g.runGitTimeout(defaultGitTimeout, "rev-list", "--left-right", "--count", status.Branch+"..."+remote)
		if err == nil {
			var ahead, behind int
			fmt.Sscanf(trimOutput(revList), "%d\t%d", &ahead, &behind)
			status.Ahead = ahead
			status.Behind = behind
		}
	}

	// Get last commit time
	lastCommit, err := g.runGitTimeout(defaultGitTimeout, "log", "-1", "--format=%ci")
	if err == nil && trimOutput(lastCommit) != "" {
		t, err := time.Parse("2006-01-02 15:04:05 -0700", trimOutput(lastCommit))
		if err == nil {
			status.LastCommitAt = &t
		}
	}

	return status, nil
}

// Commit stages and commits the specified files.
func (g *GitSync) Commit(files []string) error {
	g.opMu.Lock()
	defer g.opMu.Unlock()

	if !g.IsRepo() {
		return fmt.Errorf("not a git repository - run 'today sync --init' first")
	}

	if len(files) == 0 {
		return nil
	}

	// Stage files
	args := append([]string{"add"}, files...)
	if _, err := g.runGitTimeout(defaultGitTimeout, args...); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Check if there are staged changes.
	staged, err := g.runGitTimeout(defaultGitTimeout, "diff", "--cached", "--name-only")
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}
	if trimOutput(staged) == "" {
		// No changes staged, nothing to commit
		return nil
	}

	// Generate commit message
	message := g.generateCommitMessage(files)

	// Commit
	if _, err := g.runGitTimeout(commitGitTimeout, "-c", "commit.gpgsign=false", "commit", "-m", message); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Auto-push if enabled
	if g.config.AutoPush {
		if err := g.Push(); err != nil {
			// Log but don't fail - data is safely committed locally
			return fmt.Errorf("committed locally, but push failed: %w", err)
		}
	}

	return nil
}

// CommitAll stages and commits all tracked changes.
func (g *GitSync) CommitAll() error {
	g.opMu.Lock()
	defer g.opMu.Unlock()

	if !g.IsRepo() {
		return fmt.Errorf("not a git repository - run 'today sync --init' first")
	}

	// Stage all tracked files
	if _, err := g.runGitTimeout(defaultGitTimeout, "add", "-A"); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Check if there are staged changes
	staged, err := g.runGitTimeout(defaultGitTimeout, "diff", "--cached", "--name-only")
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}
	if trimOutput(staged) == "" {
		// No changes staged, nothing to commit
		return nil
	}

	// Commit
	message := "Update today data"
	if _, err := g.runGitTimeout(commitGitTimeout, "-c", "commit.gpgsign=false", "commit", "-m", message); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// Pull fetches and merges changes from the remote.
func (g *GitSync) Pull() error {
	g.opMu.Lock()
	defer g.opMu.Unlock()

	if !g.IsRepo() {
		return fmt.Errorf("not a git repository")
	}

	// Check if there's a remote
	remotes, err := g.runGitTimeout(defaultGitTimeout, "remote")
	if err != nil || trimOutput(remotes) == "" {
		return fmt.Errorf("no remote configured")
	}

	// Pull with rebase to keep history clean
	if _, err := g.runGitTimeout(pullPushGitTimeout, "pull", "--rebase"); err != nil {
		return fmt.Errorf("pull failed: %w", err)
	}

	return nil
}

// Push pushes local commits to the remote.
func (g *GitSync) Push() error {
	g.opMu.Lock()
	defer g.opMu.Unlock()

	if !g.IsRepo() {
		return fmt.Errorf("not a git repository")
	}

	// Check if there's a remote
	remotes, err := g.runGitTimeout(defaultGitTimeout, "remote")
	if err != nil || trimOutput(remotes) == "" {
		return fmt.Errorf("no remote configured - add one with 'git remote add origin <url>'")
	}

	// Push
	if _, err := g.runGitTimeout(pullPushGitTimeout, "push"); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	return nil
}

// AddRemote adds a git remote with the given name and URL.
// If the remote already exists, it will be updated.
func (g *GitSync) AddRemote(name, url string) error {
	g.opMu.Lock()
	defer g.opMu.Unlock()

	if !g.IsRepo() {
		return fmt.Errorf("not a git repository - run 'today sync --init' first")
	}

	if name == "" {
		return fmt.Errorf("remote name is required")
	}
	if url == "" {
		return fmt.Errorf("remote URL is required")
	}

	// Check if remote already exists
	remotes, _ := g.runGitTimeout(defaultGitTimeout, "remote")
	hasRemote := false
	for _, line := range strings.Split(trimOutput(remotes), "\n") {
		if strings.TrimSpace(line) == name {
			hasRemote = true
			break
		}
	}

	if hasRemote {
		// Update existing remote
		if _, err := g.runGitTimeout(defaultGitTimeout, "remote", "set-url", name, url); err != nil {
			return fmt.Errorf("failed to update remote: %w", err)
		}
	} else {
		// Add new remote
		if _, err := g.runGitTimeout(defaultGitTimeout, "remote", "add", name, url); err != nil {
			return fmt.Errorf("failed to add remote: %w", err)
		}
	}

	return nil
}

// OnFileSaved is called when a data file is saved.
// It queues the file for commit with debouncing.
// Deprecated: Use OnFileSavedWithContext for semantic commit messages.
func (g *GitSync) OnFileSaved(filename string) {
	if !g.config.Enabled || !g.config.AutoCommit {
		return
	}

	if !g.IsRepo() {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.pendingFiles[filename] = true

	// Reset timer - commit after debounce duration of no changes
	if g.commitTimer != nil {
		g.commitTimer.Stop()
	}
	g.commitTimer = time.AfterFunc(g.debounceDuration, g.flushCommit)
}

// OnFileSavedWithContext is called when a data file is saved with semantic context.
// It queues the file and context for a commit with debouncing.
// The context enables meaningful commit messages like "Complete task: Review PR".
func (g *GitSync) OnFileSavedWithContext(ctx storage.SaveContext) {
	if !g.config.Enabled || !g.config.AutoCommit {
		return
	}

	if !g.IsRepo() {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	g.pendingFiles[ctx.Filename] = true
	g.pendingContexts = append(g.pendingContexts, ctx)

	// Reset timer - commit after debounce duration of no changes
	if g.commitTimer != nil {
		g.commitTimer.Stop()
	}
	g.commitTimer = time.AfterFunc(g.debounceDuration, g.flushCommit)
}

// Flush immediately commits any pending files without waiting for debounce.
func (g *GitSync) Flush() {
	g.mu.Lock()
	if g.commitTimer != nil {
		g.commitTimer.Stop()
		g.commitTimer = nil
	}
	g.mu.Unlock()

	g.flushCommit()
}

// flushCommit commits all pending files with their contexts.
func (g *GitSync) flushCommit() {
	g.mu.Lock()
	files := make([]string, 0, len(g.pendingFiles))
	for f := range g.pendingFiles {
		files = append(files, f)
	}
	contexts := g.pendingContexts
	g.pendingFiles = make(map[string]bool)
	g.pendingContexts = nil
	g.mu.Unlock()

	if len(files) > 0 {
		// Ignore errors from auto-commit (log them in future)
		_ = g.commitWithContexts(files, contexts)
	}
}

// commitWithContexts stages and commits files with semantic context for the message.
func (g *GitSync) commitWithContexts(files []string, contexts []storage.SaveContext) error {
	g.opMu.Lock()
	defer g.opMu.Unlock()

	if !g.IsRepo() {
		return fmt.Errorf("not a git repository - run 'today sync --init' first")
	}

	if len(files) == 0 {
		return nil
	}

	// Stage files
	args := append([]string{"add"}, files...)
	if _, err := g.runGitTimeout(defaultGitTimeout, args...); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Check if there are staged changes.
	staged, err := g.runGitTimeout(defaultGitTimeout, "diff", "--cached", "--name-only")
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}
	if trimOutput(staged) == "" {
		// No changes staged, nothing to commit
		return nil
	}

	// Generate commit message from contexts
	message := g.generateCommitMessageFromContexts(files, contexts)

	// Commit
	if _, err := g.runGitTimeout(commitGitTimeout, "-c", "commit.gpgsign=false", "commit", "-m", message); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	// Auto-push if enabled
	if g.config.AutoPush {
		if err := g.Push(); err != nil {
			// Log but don't fail - data is safely committed locally
			return fmt.Errorf("committed locally, but push failed: %w", err)
		}
	}

	return nil
}

// generateCommitMessage creates an appropriate commit message for the files.
// Deprecated: Use generateCommitMessageFromContexts for semantic messages.
func (g *GitSync) generateCommitMessage(files []string) string {
	if g.config.CommitMessage != "" && g.config.CommitMessage != "auto" {
		return g.config.CommitMessage
	}

	if len(files) == 1 {
		switch files[0] {
		case "tasks.json":
			return "Update tasks"
		case "habits.json":
			return "Update habits"
		case "timer.json":
			return "Update timer"
		}
		return fmt.Sprintf("Update %s", files[0])
	}

	return fmt.Sprintf("Update %d files", len(files))
}

// generateCommitMessageFromContexts creates semantic commit messages from contexts.
// Examples: "Complete task: Review PR", "Add habit: Exercise", "Start timer: work"
func (g *GitSync) generateCommitMessageFromContexts(files []string, contexts []storage.SaveContext) string {
	// If custom message configured, use it
	if g.config.CommitMessage != "" && g.config.CommitMessage != "auto" {
		return g.config.CommitMessage
	}

	// No contexts - fall back to file-based message
	if len(contexts) == 0 {
		return g.generateCommitMessage(files)
	}

	// Single context - generate semantic message
	if len(contexts) == 1 {
		ctx := contexts[0]
		return formatSemanticMessage(ctx)
	}

	// Multiple contexts - check if they're all the same operation type
	firstOp := contexts[0].Operation
	firstType := contexts[0].ItemType
	allSame := true
	for _, ctx := range contexts[1:] {
		if ctx.Operation != firstOp || ctx.ItemType != firstType {
			allSame = false
			break
		}
	}

	if allSame {
		// All same operation: "Complete 3 tasks"
		return fmt.Sprintf("%s %d %ss", capitalizeFirst(firstOp), len(contexts), firstType)
	}

	// Mixed operations - summarize
	return fmt.Sprintf("Update: %d changes", len(contexts))
}

// formatSemanticMessage creates a human-readable message from a SaveContext.
func formatSemanticMessage(ctx storage.SaveContext) string {
	verb := capitalizeFirst(ctx.Operation)

	switch ctx.Operation {
	case "add":
		return fmt.Sprintf("Add %s: %s", ctx.ItemType, ctx.ItemName)
	case "complete":
		return fmt.Sprintf("Complete %s: %s", ctx.ItemType, ctx.ItemName)
	case "delete":
		return fmt.Sprintf("Delete %s: %s", ctx.ItemType, ctx.ItemName)
	case "reopen":
		return fmt.Sprintf("Reopen %s: %s", ctx.ItemType, ctx.ItemName)
	case "toggle":
		return fmt.Sprintf("Toggle %s: %s", ctx.ItemType, ctx.ItemName)
	case "restore":
		return fmt.Sprintf("Restore %s: %s", ctx.ItemType, ctx.ItemName)
	case "start":
		return fmt.Sprintf("Start %s: %s", ctx.ItemType, ctx.ItemName)
	case "stop":
		return fmt.Sprintf("Stop %s: %s", ctx.ItemType, ctx.ItemName)
	default:
		if ctx.ItemName != "" {
			return fmt.Sprintf("%s %s: %s", verb, ctx.ItemType, ctx.ItemName)
		}
		return fmt.Sprintf("%s %s", verb, ctx.ItemType)
	}
}

// capitalizeFirst capitalizes the first letter of a string.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// runGit executes a git command and returns its output.
func (g *GitSync) runGit(args ...string) (string, error) {
	return g.runGitTimeout(defaultGitTimeout, args...)
}

func (g *GitSync) runGitTimeout(timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = g.dataDir
	cmd.Env = envWithOverrides(os.Environ(), map[string]string{
		"GIT_TERMINAL_PROMPT": "0",
		"GIT_ASKPASS":         "",
		"SSH_ASKPASS":         "",
	})
	cmd.Stdin = bytes.NewReader(nil)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("git %s timed out after %s", strings.Join(args, " "), timeout)
		}

		errMsg := stderr.String()
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("%s", trimOutput(errMsg))
	}
	return stdout.String(), nil
}

func envWithOverrides(base []string, overrides map[string]string) []string {
	if len(overrides) == 0 {
		return base
	}
	out := make([]string, 0, len(base)+len(overrides))
	seen := make(map[string]bool, len(overrides))
	for _, kv := range base {
		k, _, ok := strings.Cut(kv, "=")
		if !ok {
			out = append(out, kv)
			continue
		}
		if v, ok := overrides[k]; ok {
			out = append(out, k+"="+v)
			seen[k] = true
			continue
		}
		out = append(out, kv)
	}
	for k, v := range overrides {
		if !seen[k] {
			out = append(out, k+"="+v)
		}
	}
	return out
}

func isGitNothingToCommit(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "nothing to commit") ||
		strings.Contains(msg, "nothing added to commit") ||
		strings.Contains(msg, "no changes added to commit")
}

// trimOutput removes leading/trailing whitespace from command output.
func trimOutput(s string) string {
	return string(bytes.TrimSpace([]byte(s)))
}
