package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.DataDir == "" {
		t.Error("DataDir should not be empty")
	}

	if cfg.Theme.Primary == "" {
		t.Error("Theme.Primary should have a default value")
	}

	if cfg.Theme.Accent == "" {
		t.Error("Theme.Accent should have a default value")
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	// Set a temp XDG_CONFIG_HOME to avoid loading real config
	tempDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should return defaults
	if cfg.Theme.Primary != "#7C3AED" {
		t.Errorf("Theme.Primary = %q, want #7C3AED", cfg.Theme.Primary)
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Create config file
	configDir := filepath.Join(tempDir, "today")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `
data_dir: /custom/data
theme:
  primary: "#FF0000"
  accent: "#00FF00"
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.DataDir != "/custom/data" {
		t.Errorf("DataDir = %q, want /custom/data", cfg.DataDir)
	}

	if cfg.Theme.Primary != "#FF0000" {
		t.Errorf("Theme.Primary = %q, want #FF0000", cfg.Theme.Primary)
	}

	if cfg.Theme.Accent != "#00FF00" {
		t.Errorf("Theme.Accent = %q, want #00FF00", cfg.Theme.Accent)
	}

	// Muted should still be default
	if cfg.Theme.Muted != "#6B7280" {
		t.Errorf("Theme.Muted = %q, want #6B7280", cfg.Theme.Muted)
	}
}

func TestMerge(t *testing.T) {
	base := Default()
	override := &Config{
		DataDir: "/override/path",
		Theme: ThemeConfig{
			Primary: "#CUSTOM",
		},
	}

	base.mergeNonEmpty(override)

	if base.DataDir != "/override/path" {
		t.Errorf("DataDir = %q, want /override/path", base.DataDir)
	}

	if base.Theme.Primary != "#CUSTOM" {
		t.Errorf("Theme.Primary = %q, want #CUSTOM", base.Theme.Primary)
	}

	// Accent should remain default
	if base.Theme.Accent != "#10B981" {
		t.Errorf("Theme.Accent = %q, want #10B981", base.Theme.Accent)
	}
}

func TestLoad_MissingBoolKeysDoesNotClobberDefaults(t *testing.T) {
	tempDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	// Create config file that only touches theme, not UX/sync booleans.
	configDir := filepath.Join(tempDir, "today")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `
theme:
  primary: "#FF0000"
sync:
  enabled: true
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Explicit key should override default.
	if !cfg.Sync.Enabled {
		t.Errorf("Sync.Enabled = %v, want true", cfg.Sync.Enabled)
	}

	// Omitted keys must not clobber defaults.
	if !cfg.UX.ConfirmDeletions {
		t.Errorf("UX.ConfirmDeletions = %v, want true", cfg.UX.ConfirmDeletions)
	}
	if !cfg.UX.ShowOnboarding {
		t.Errorf("UX.ShowOnboarding = %v, want true", cfg.UX.ShowOnboarding)
	}
	if !cfg.Sync.AutoCommit {
		t.Errorf("Sync.AutoCommit = %v, want true", cfg.Sync.AutoCommit)
	}
}

func TestLoad_ExplicitFalseOverridesDefault(t *testing.T) {
	tempDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	configDir := filepath.Join(tempDir, "today")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configContent := `
ux:
  confirm_deletions: false
sync:
  enabled: true
  auto_commit: false
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.UX.ConfirmDeletions {
		t.Errorf("UX.ConfirmDeletions = %v, want false", cfg.UX.ConfirmDeletions)
	}
	if cfg.Sync.AutoCommit {
		t.Errorf("Sync.AutoCommit = %v, want false", cfg.Sync.AutoCommit)
	}
	if !cfg.Sync.Enabled {
		t.Errorf("Sync.Enabled = %v, want true", cfg.Sync.Enabled)
	}
}

func TestGetDataDir(t *testing.T) {
	tests := []struct {
		name    string
		dataDir string
		want    string
	}{
		{
			name:    "empty uses default",
			dataDir: "",
			want:    "",
		},
		{
			name:    "absolute path",
			dataDir: "/custom/path",
			want:    "/custom/path",
		},
	}

	if home, err := os.UserHomeDir(); err == nil && home != "" {
		tests = append(tests,
			struct {
				name    string
				dataDir string
				want    string
			}{
				name:    "tilde expands home",
				dataDir: "~",
				want:    home,
			},
			struct {
				name    string
				dataDir string
				want    string
			}{
				name:    "tilde path expands home",
				dataDir: "~/mydata",
				want:    filepath.Join(home, "mydata"),
			},
		)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{DataDir: tt.dataDir}
			got := cfg.GetDataDir()

			if tt.dataDir == "" {
				// Should end with .today
				if filepath.Base(got) != ".today" {
					t.Errorf("GetDataDir() = %q, want to end with .today", got)
				}
			} else {
				if tt.want != "" && got != tt.want {
					t.Errorf("GetDataDir() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestSave(t *testing.T) {
	tempDir := t.TempDir()
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

	cfg := Default()
	cfg.DataDir = "/saved/path"
	cfg.Theme.Primary = "#SAVED"

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tempDir, "today", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("config file not created: %v", err)
	}

	// Load and verify
	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.DataDir != "/saved/path" {
		t.Errorf("loaded DataDir = %q, want /saved/path", loaded.DataDir)
	}

	if loaded.Theme.Primary != "#SAVED" {
		t.Errorf("loaded Theme.Primary = %q, want #SAVED", loaded.Theme.Primary)
	}
}
