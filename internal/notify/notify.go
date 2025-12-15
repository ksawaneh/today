// Package notify provides cross-platform desktop notification support.
// It uses native notification mechanisms on macOS (osascript) and Linux (notify-send).
package notify

// Notifier defines the interface for sending desktop notifications.
type Notifier interface {
	// Send sends a notification with the given title and message.
	Send(title, message string) error

	// SendWithSound sends a notification with sound.
	SendWithSound(title, message string) error

	// IsSupported returns true if notifications are supported on this platform.
	IsSupported() bool
}

type noopNotifier struct{}

func (n *noopNotifier) Send(title, message string) error {
	return nil
}

func (n *noopNotifier) SendWithSound(title, message string) error {
	return nil
}

func (n *noopNotifier) IsSupported() bool {
	return false
}

// New creates a platform-specific notifier.
// Returns a no-op notifier if the platform doesn't support notifications.
func New() Notifier {
	n := newPlatformNotifier()
	if n == nil || !n.IsSupported() {
		return &noopNotifier{}
	}
	return n
}

// Config holds notification configuration.
type Config struct {
	// Enabled enables/disables notifications
	Enabled bool `yaml:"enabled"`

	// HabitReminder is the time for daily habit reminders (HH:MM format)
	HabitReminder string `yaml:"habit_reminder"`

	// TimerMilestones are durations (in minutes) at which to notify
	TimerMilestones []int `yaml:"timer_milestones"`

	// Sound enables notification sounds
	Sound bool `yaml:"sound"`
}

// DefaultConfig returns the default notification configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:         false,
		HabitReminder:   "",    // No reminder by default
		TimerMilestones: nil,   // No milestones by default
		Sound:           false, // No sound by default
	}
}
