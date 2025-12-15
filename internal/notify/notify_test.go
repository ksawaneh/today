// Package notify provides desktop notification support.
// This file contains tests for the notification functionality.
package notify

import (
	"os"
	"runtime"
	"testing"
)

// TestNew tests that New() returns a valid notifier.
func TestNew(t *testing.T) {
	n := New()
	if n == nil {
		t.Error("New() returned nil")
	}
}

// TestIsSupported tests platform detection.
func TestIsSupported(t *testing.T) {
	n := New()

	// On macOS and Linux, notifications should typically be supported
	// (osascript and notify-send are usually available)
	switch runtime.GOOS {
	case "darwin":
		// osascript should be available on macOS
		if !n.IsSupported() {
			t.Log("Warning: osascript not available on macOS")
		}
	case "linux":
		// notify-send may or may not be available
		t.Logf("Linux notification support: %v", n.IsSupported())
	default:
		// Other platforms should return false
		if n.IsSupported() {
			t.Errorf("IsSupported() should be false on %s", runtime.GOOS)
		}
	}
}

// TestSend tests sending a notification.
// This is a manual test - it will actually show a notification.
func TestSend(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping notification test in short mode")
	}
	if os.Getenv("RUN_NOTIFY_TESTS") != "1" {
		t.Skip("Skipping manual notification test (set RUN_NOTIFY_TESTS=1 to enable)")
	}

	n := New()
	if !n.IsSupported() {
		t.Skip("Notifications not supported on this platform")
	}

	// This will actually display a notification
	err := n.Send("today test", "This is a test notification")
	if err != nil {
		t.Errorf("Send() error: %v", err)
	}
}

// TestDefaultConfig tests the default configuration.
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("Expected Enabled to be false by default")
	}

	if cfg.HabitReminder != "" {
		t.Error("Expected HabitReminder to be empty by default")
	}

	if len(cfg.TimerMilestones) > 0 {
		t.Error("Expected TimerMilestones to be empty by default")
	}

	if cfg.Sound {
		t.Error("Expected Sound to be false by default")
	}
}

// TestEscapeAppleScript tests AppleScript string escaping (darwin only).
func TestEscapeAppleScript(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("AppleScript escaping only relevant on macOS")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"Hello", "Hello"},
		{`Hello "World"`, `Hello \"World\"`},
		{`Path\to\file`, `Path\\to\\file`},
		{`Mix "quote" and \slash`, `Mix \"quote\" and \\slash`},
	}

	for _, tc := range tests {
		result := escapeAppleScript(tc.input)
		if result != tc.expected {
			t.Errorf("escapeAppleScript(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}
