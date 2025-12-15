package ui

import (
	"os"
	"path/filepath"
	"testing"

	"today/internal/config"
	"today/internal/storage"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// setupTest prepares the test environment for deterministic rendering.
// It disables colors to ensure consistent golden file output across environments.
func setupTest(t *testing.T) {
	t.Helper()
	// Use ASCII profile to disable all color codes in output
	lipgloss.SetColorProfile(termenv.Ascii)
}

// createTestStorage creates a Storage instance with a temporary directory.
func createTestStorage(t *testing.T) *storage.Storage {
	t.Helper()
	store, err := storage.New(t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test storage: %v", err)
	}
	return store
}

// createTestStyles creates a default Styles instance for testing.
func createTestStyles() *Styles {
	return NewStylesFromTheme(&config.ThemeConfig{})
}

// goldenPath returns the path to a golden file in the testdata directory.
func goldenPath(name string) string {
	return filepath.Join("testdata", name+".golden")
}

// updateGolden checks if the -update flag is set for updating golden files.
var updateGolden = os.Getenv("UPDATE_GOLDEN") == "1"

// assertGolden compares output against a golden file.
// If UPDATE_GOLDEN=1 is set, it updates the golden file instead.
func assertGolden(t *testing.T, name string, actual string) {
	t.Helper()

	path := goldenPath(name)

	if updateGolden {
		// Create testdata directory if it doesn't exist
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata directory: %v", err)
		}
		if err := os.WriteFile(path, []byte(actual), 0644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		t.Logf("Updated golden file: %s", path)
		return
	}

	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v\nRun with UPDATE_GOLDEN=1 to create it", path, err)
	}

	if actual != string(expected) {
		t.Errorf("output mismatch for %s\n\nGot:\n%s\n\nWant:\n%s", name, actual, string(expected))
	}
}
