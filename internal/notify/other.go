//go:build !darwin && !linux

// Package notify provides desktop notification support.
// This file provides a no-op implementation for unsupported platforms.
package notify

// stubNotifier is a no-op notifier for unsupported platforms.
type stubNotifier struct{}

// newPlatformNotifier creates a stub notifier.
func newPlatformNotifier() Notifier {
	return &stubNotifier{}
}

// Send is a no-op on unsupported platforms.
func (n *stubNotifier) Send(title, message string) error {
	return nil
}

// SendWithSound is a no-op on unsupported platforms.
func (n *stubNotifier) SendWithSound(title, message string) error {
	return nil
}

// IsSupported returns false for unsupported platforms.
func (n *stubNotifier) IsSupported() bool {
	return false
}
