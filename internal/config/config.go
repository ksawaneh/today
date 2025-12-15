// Package config handles configuration loading and defaults for the today app.
// Configuration is loaded from XDG-compliant paths (typically ~/.config/today/config.yaml).
package config

import (
	"os"
	"path/filepath"
	"strings"

	"today/internal/fsutil"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	// DataDir overrides the default data directory (~/.today)
	DataDir string `yaml:"data_dir,omitempty"`

	// Theme customizes the visual appearance
	Theme ThemeConfig `yaml:"theme,omitempty"`

	// Keys customizes keyboard shortcuts
	Keys KeysConfig `yaml:"keys,omitempty"`

	// UX customizes user experience settings
	UX UXConfig `yaml:"ux,omitempty"`

	// Sync configures git synchronization
	Sync SyncConfig `yaml:"sync,omitempty"`

	// Notifications configures desktop notifications
	Notifications NotificationConfig `yaml:"notifications,omitempty"`
}

// NotificationConfig defines desktop notification settings.
type NotificationConfig struct {
	// Enabled enables/disables notifications
	Enabled bool `yaml:"enabled,omitempty"`

	// HabitReminder is the time for daily habit reminders (HH:MM format)
	HabitReminder string `yaml:"habit_reminder,omitempty"`

	// TimerMilestones are durations (in minutes) at which to notify
	TimerMilestones []int `yaml:"timer_milestones,omitempty"`

	// Sound enables notification sounds
	Sound bool `yaml:"sound,omitempty"`
}

// SyncConfig defines git synchronization settings.
type SyncConfig struct {
	// Enabled enables/disables git sync
	Enabled bool `yaml:"enabled,omitempty"`

	// AutoCommit automatically commits changes after saves
	AutoCommit bool `yaml:"auto_commit,omitempty"`

	// AutoPush automatically pushes after each commit
	AutoPush bool `yaml:"auto_push,omitempty"`

	// PullOnStartup pulls latest changes when the app starts
	PullOnStartup bool `yaml:"pull_on_startup,omitempty"`

	// CommitMessage is the commit message template ("auto" for auto-generated)
	CommitMessage string `yaml:"commit_message,omitempty"`
}

// ThemeConfig defines color and style settings.
type ThemeConfig struct {
	// Primary color for focused elements (hex, e.g., "#FF5733")
	Primary string `yaml:"primary,omitempty"`

	// Accent color for highlights (hex)
	Accent string `yaml:"accent,omitempty"`

	// Muted color for secondary text (hex)
	Muted string `yaml:"muted,omitempty"`

	// Background color (hex)
	Background string `yaml:"background,omitempty"`

	// Text color (hex)
	Text string `yaml:"text,omitempty"`
}

// KeysConfig defines customizable keyboard shortcuts.
// Each field accepts a comma-separated list of key bindings.
// Examples: "q,ctrl+c", "tab", "j,down"
type KeysConfig struct {
	// Global keys
	Quit     string `yaml:"quit,omitempty"`      // default: "q,ctrl+c"
	Help     string `yaml:"help,omitempty"`      // default: "?"
	NextPane string `yaml:"next_pane,omitempty"` // default: "tab"
	Pane1    string `yaml:"pane_1,omitempty"`    // default: "1"
	Pane2    string `yaml:"pane_2,omitempty"`    // default: "2"
	Pane3    string `yaml:"pane_3,omitempty"`    // default: "3"

	// Navigation keys
	Up     string `yaml:"up,omitempty"`     // default: "k,up"
	Down   string `yaml:"down,omitempty"`   // default: "j,down"
	Top    string `yaml:"top,omitempty"`    // default: "g"
	Bottom string `yaml:"bottom,omitempty"` // default: "G"

	// Task keys
	AddTask    string `yaml:"add_task,omitempty"`    // default: "a"
	ToggleTask string `yaml:"toggle_task,omitempty"` // default: "d,enter,space"
	DeleteTask string `yaml:"delete_task,omitempty"` // default: "x"

	// Habit keys
	AddHabit    string `yaml:"add_habit,omitempty"`    // default: "a"
	ToggleHabit string `yaml:"toggle_habit,omitempty"` // default: "d,enter,space"
	DeleteHabit string `yaml:"delete_habit,omitempty"` // default: "x"

	// Timer keys
	ToggleTimer string `yaml:"toggle_timer,omitempty"` // default: "space,enter"
	SwitchTimer string `yaml:"switch_timer,omitempty"` // default: "s"
	StopTimer   string `yaml:"stop_timer,omitempty"`   // default: "x"

	// Input keys
	Confirm string `yaml:"confirm,omitempty"` // default: "enter"
	Cancel  string `yaml:"cancel,omitempty"`  // default: "esc"

	// Undo/Redo keys
	Undo string `yaml:"undo,omitempty"` // default: "ctrl+z,u"
	Redo string `yaml:"redo,omitempty"` // default: "ctrl+y"
}

// UXConfig defines user experience settings.
type UXConfig struct {
	// ConfirmDeletions shows confirmation dialogs before deleting items
	ConfirmDeletions bool `yaml:"confirm_deletions,omitempty"` // default: true

	// ShowOnboarding shows welcome screen on first run
	ShowOnboarding bool `yaml:"show_onboarding,omitempty"` // default: true

	// NarrowLayoutThreshold is the terminal width below which to use stacked layout
	NarrowLayoutThreshold int `yaml:"narrow_layout_threshold,omitempty"` // default: 80
}

// Default returns the default configuration.
func Default() *Config {
	return &Config{
		DataDir: defaultDataDir(),
		Theme: ThemeConfig{
			Primary:    "#7C3AED", // Violet
			Accent:     "#10B981", // Emerald
			Muted:      "#6B7280", // Gray
			Background: "",        // Terminal default
			Text:       "",        // Terminal default
		},
		Keys: KeysConfig{
			// Defaults are empty strings, which means use built-in defaults
		},
		UX: UXConfig{
			ConfirmDeletions:      true,
			ShowOnboarding:        true,
			NarrowLayoutThreshold: 80,
		},
		Sync: SyncConfig{
			Enabled:       false, // Disabled by default
			AutoCommit:    true,  // Auto-commit when enabled
			AutoPush:      false, // Don't auto-push by default
			PullOnStartup: false, // Don't auto-pull by default
			CommitMessage: "auto",
		},
		Notifications: NotificationConfig{
			Enabled:         false, // Disabled by default
			HabitReminder:   "",    // No reminder by default
			TimerMilestones: nil,   // No milestones by default
			Sound:           false, // No sound by default
		},
	}
}

// defaultDataDir returns the default data directory path.
func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".today"
	}
	return filepath.Join(home, ".today")
}

// configDir returns the configuration directory path (XDG compliant).
func configDir() string {
	// Check XDG_CONFIG_HOME first
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "today")
	}

	// Fall back to ~/.config/today
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "today")
}

// configPath returns the path to the config file.
func configPath() string {
	dir := configDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "config.yaml")
}

// Load reads configuration from disk, merging with defaults.
// If no config file exists, returns default configuration.
func Load() (*Config, error) {
	cfg := Default()

	path := configPath()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file, use defaults
			return cfg, nil
		}
		return nil, err
	}

	// Parse YAML and merge with defaults
	var userCfg Config
	if err := yaml.Unmarshal(data, &userCfg); err != nil {
		return nil, err
	}

	var doc yaml.Node
	_ = yaml.Unmarshal(data, &doc) // best-effort; fall back to conservative merge if this fails

	// Merge user config with defaults (presence-aware for booleans/slices)
	cfg.mergeFromYAML(&userCfg, &doc)

	return cfg, nil
}

// mergeNonEmpty applies non-empty values from other to c.
// It intentionally does not touch booleans or slices (those require presence-aware merging).
func (c *Config) mergeNonEmpty(other *Config) {
	if other.DataDir != "" {
		c.DataDir = other.DataDir
	}

	// Theme merging
	if other.Theme.Primary != "" {
		c.Theme.Primary = other.Theme.Primary
	}
	if other.Theme.Accent != "" {
		c.Theme.Accent = other.Theme.Accent
	}
	if other.Theme.Muted != "" {
		c.Theme.Muted = other.Theme.Muted
	}
	if other.Theme.Background != "" {
		c.Theme.Background = other.Theme.Background
	}
	if other.Theme.Text != "" {
		c.Theme.Text = other.Theme.Text
	}

	// Keys merging
	if other.Keys.Quit != "" {
		c.Keys.Quit = other.Keys.Quit
	}
	if other.Keys.Help != "" {
		c.Keys.Help = other.Keys.Help
	}
	if other.Keys.NextPane != "" {
		c.Keys.NextPane = other.Keys.NextPane
	}
	if other.Keys.Pane1 != "" {
		c.Keys.Pane1 = other.Keys.Pane1
	}
	if other.Keys.Pane2 != "" {
		c.Keys.Pane2 = other.Keys.Pane2
	}
	if other.Keys.Pane3 != "" {
		c.Keys.Pane3 = other.Keys.Pane3
	}
	if other.Keys.Up != "" {
		c.Keys.Up = other.Keys.Up
	}
	if other.Keys.Down != "" {
		c.Keys.Down = other.Keys.Down
	}
	if other.Keys.Top != "" {
		c.Keys.Top = other.Keys.Top
	}
	if other.Keys.Bottom != "" {
		c.Keys.Bottom = other.Keys.Bottom
	}
	if other.Keys.AddTask != "" {
		c.Keys.AddTask = other.Keys.AddTask
	}
	if other.Keys.ToggleTask != "" {
		c.Keys.ToggleTask = other.Keys.ToggleTask
	}
	if other.Keys.DeleteTask != "" {
		c.Keys.DeleteTask = other.Keys.DeleteTask
	}
	if other.Keys.AddHabit != "" {
		c.Keys.AddHabit = other.Keys.AddHabit
	}
	if other.Keys.ToggleHabit != "" {
		c.Keys.ToggleHabit = other.Keys.ToggleHabit
	}
	if other.Keys.DeleteHabit != "" {
		c.Keys.DeleteHabit = other.Keys.DeleteHabit
	}
	if other.Keys.ToggleTimer != "" {
		c.Keys.ToggleTimer = other.Keys.ToggleTimer
	}
	if other.Keys.SwitchTimer != "" {
		c.Keys.SwitchTimer = other.Keys.SwitchTimer
	}
	if other.Keys.StopTimer != "" {
		c.Keys.StopTimer = other.Keys.StopTimer
	}
	if other.Keys.Confirm != "" {
		c.Keys.Confirm = other.Keys.Confirm
	}
	if other.Keys.Cancel != "" {
		c.Keys.Cancel = other.Keys.Cancel
	}
	if other.Keys.Undo != "" {
		c.Keys.Undo = other.Keys.Undo
	}
	if other.Keys.Redo != "" {
		c.Keys.Redo = other.Keys.Redo
	}

	// UX ints/strings (presence-aware in mergeFromYAML)
	if other.UX.NarrowLayoutThreshold > 0 {
		c.UX.NarrowLayoutThreshold = other.UX.NarrowLayoutThreshold
	}

	// Sync strings (presence-aware in mergeFromYAML)
	if other.Sync.CommitMessage != "" {
		c.Sync.CommitMessage = other.Sync.CommitMessage
	}

	// Notifications strings (presence-aware in mergeFromYAML)
	if other.Notifications.HabitReminder != "" {
		c.Notifications.HabitReminder = other.Notifications.HabitReminder
	}
}

func (c *Config) mergeFromYAML(other *Config, doc *yaml.Node) {
	// Fall back to conservative behavior if we can't inspect presence.
	if doc == nil || len(doc.Content) == 0 {
		// Avoid clobbering defaults with zero-values: only apply non-empty strings and non-zero ints.
		c.mergeNonEmpty(other)
		if len(other.Notifications.TimerMilestones) > 0 {
			c.Notifications.TimerMilestones = other.Notifications.TimerMilestones
		}
		return
	}

	// First apply all non-empty string-ish merges.
	c.mergeNonEmpty(other)

	// Now re-apply booleans and slices only when present in YAML.
	if yamlHasPath(doc, "ux", "confirm_deletions") {
		c.UX.ConfirmDeletions = other.UX.ConfirmDeletions
	}
	if yamlHasPath(doc, "ux", "show_onboarding") {
		c.UX.ShowOnboarding = other.UX.ShowOnboarding
	}
	if yamlHasPath(doc, "ux", "narrow_layout_threshold") && other.UX.NarrowLayoutThreshold > 0 {
		c.UX.NarrowLayoutThreshold = other.UX.NarrowLayoutThreshold
	}

	if yamlHasPath(doc, "sync", "enabled") {
		c.Sync.Enabled = other.Sync.Enabled
	}
	if yamlHasPath(doc, "sync", "auto_commit") {
		c.Sync.AutoCommit = other.Sync.AutoCommit
	}
	if yamlHasPath(doc, "sync", "auto_push") {
		c.Sync.AutoPush = other.Sync.AutoPush
	}
	if yamlHasPath(doc, "sync", "pull_on_startup") {
		c.Sync.PullOnStartup = other.Sync.PullOnStartup
	}
	if yamlHasPath(doc, "sync", "commit_message") {
		c.Sync.CommitMessage = other.Sync.CommitMessage
	}

	if yamlHasPath(doc, "notifications", "enabled") {
		c.Notifications.Enabled = other.Notifications.Enabled
	}
	if yamlHasPath(doc, "notifications", "sound") {
		c.Notifications.Sound = other.Notifications.Sound
	}
	if yamlHasPath(doc, "notifications", "habit_reminder") {
		c.Notifications.HabitReminder = other.Notifications.HabitReminder
	}
	if yamlHasPath(doc, "notifications", "timer_milestones") {
		c.Notifications.TimerMilestones = other.Notifications.TimerMilestones
	}
}

func yamlHasPath(doc *yaml.Node, path ...string) bool {
	if doc == nil || len(path) == 0 {
		return false
	}

	// Document -> root mapping.
	n := doc
	if n.Kind == yaml.DocumentNode && len(n.Content) > 0 {
		n = n.Content[0]
	}
	for _, key := range path {
		if n == nil || n.Kind != yaml.MappingNode {
			return false
		}
		var next *yaml.Node
		for i := 0; i+1 < len(n.Content); i += 2 {
			k := n.Content[i]
			v := n.Content[i+1]
			if k.Kind == yaml.ScalarNode && k.Value == key {
				next = v
				break
			}
		}
		if next == nil {
			return false
		}
		n = next
	}
	return true
}

// Save writes the configuration to disk.
func (c *Config) Save() error {
	path := configPath()
	if path == "" {
		return nil
	}

	// Create config directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return fsutil.WriteFileAtomic(path, data, 0600)
}

// GetDataDir returns the resolved data directory path.
func (c *Config) GetDataDir() string {
	if c.DataDir != "" {
		// Expand ~ if present
		if c.DataDir == "~" {
			home, err := os.UserHomeDir()
			if err == nil {
				return home
			}
			return c.DataDir
		}

		if strings.HasPrefix(c.DataDir, "~/") || strings.HasPrefix(c.DataDir, `~\`) {
			home, err := os.UserHomeDir()
			if err == nil {
				trimmed := strings.TrimPrefix(c.DataDir, "~/")
				trimmed = strings.TrimPrefix(trimmed, `~\`)
				trimmed = strings.TrimPrefix(trimmed, `\`)
				return filepath.Join(home, trimmed)
			}
		}
		return c.DataDir
	}
	return defaultDataDir()
}
