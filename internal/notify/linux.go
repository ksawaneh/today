//go:build linux

// Package notify provides desktop notification support.
// This file implements Linux notifications using notify-send.
package notify

import (
	"fmt"
	"os/exec"
)

// linuxNotifier implements notifications for Linux using notify-send.
type linuxNotifier struct{}

// newPlatformNotifier creates the Linux notifier.
func newPlatformNotifier() Notifier {
	return &linuxNotifier{}
}

// Send sends a notification without sound.
func (n *linuxNotifier) Send(title, message string) error {
	return n.sendNotification(title, message, false)
}

// SendWithSound sends a notification with sound.
// Note: Sound support depends on the notification daemon configuration.
func (n *linuxNotifier) SendWithSound(title, message string) error {
	return n.sendNotification(title, message, true)
}

// IsSupported returns true if notify-send is available.
func (n *linuxNotifier) IsSupported() bool {
	_, err := exec.LookPath("notify-send")
	return err == nil
}

// sendNotification sends a Linux notification using notify-send.
func (n *linuxNotifier) sendNotification(title, message string, sound bool) error {
	args := []string{
		"--app-name=today",
		title,
		message,
	}

	// Add urgency hint for sound (depends on notification daemon)
	if sound {
		args = append([]string{"--urgency=normal"}, args...)
	}

	cmd := exec.Command("notify-send", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("notify-send failed: %w", err)
	}

	return nil
}
