//go:build darwin

// Package notify provides desktop notification support.
// This file implements macOS notifications using osascript.
package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

// darwinNotifier implements notifications for macOS using osascript.
type darwinNotifier struct{}

// newPlatformNotifier creates the macOS notifier.
func newPlatformNotifier() Notifier {
	return &darwinNotifier{}
}

// Send sends a notification without sound.
func (n *darwinNotifier) Send(title, message string) error {
	return n.sendNotification(title, message, false)
}

// SendWithSound sends a notification with sound.
func (n *darwinNotifier) SendWithSound(title, message string) error {
	return n.sendNotification(title, message, true)
}

// IsSupported returns true if osascript is available.
func (n *darwinNotifier) IsSupported() bool {
	_, err := exec.LookPath("osascript")
	return err == nil
}

// sendNotification sends a macOS notification using osascript.
func (n *darwinNotifier) sendNotification(title, message string, sound bool) error {
	// Escape quotes in title and message
	title = escapeAppleScript(title)
	message = escapeAppleScript(message)

	var script string
	if sound {
		script = fmt.Sprintf(`display notification %q with title %q sound name "default"`, message, title)
	} else {
		script = fmt.Sprintf(`display notification %q with title %q`, message, title)
	}

	cmd := exec.Command("osascript", "-e", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("osascript failed: %w", err)
	}

	return nil
}

// escapeAppleScript escapes special characters for AppleScript strings.
func escapeAppleScript(s string) string {
	// Replace backslashes and quotes
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}
