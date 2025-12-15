// Package ui provides terminal user interface components for the today app.
// This file defines key bindings using the Bubble Tea key package for
// type-safe key matching, help text generation, and future customization.
package ui

import (
	"strings"

	"today/internal/config"

	"github.com/charmbracelet/bubbles/key"
)

// =============================================================================
// Helpers
// =============================================================================

// parseKeys splits a comma-separated string into individual keys.
// If the input is empty, returns the default keys.
func parseKeys(customKeys string, defaultKeys ...string) []string {
	if customKeys == "" {
		return defaultKeys
	}
	keys := strings.Split(customKeys, ",")
	result := make([]string, 0, len(keys))
	for _, k := range keys {
		trimmed := strings.TrimSpace(k)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// =============================================================================
// Global Keys (available in all contexts)
// =============================================================================

// GlobalKeyMap defines keys available throughout the application.
type GlobalKeyMap struct {
	Quit     key.Binding
	Help     key.Binding
	NextPane key.Binding
	Pane1    key.Binding
	Pane2    key.Binding
	Pane3    key.Binding
	Undo     key.Binding
	Redo     key.Binding
}

// DefaultGlobalKeyMap returns the default global key bindings.
func DefaultGlobalKeyMap() GlobalKeyMap {
	return NewGlobalKeyMap(&config.KeysConfig{})
}

// NewGlobalKeyMap creates global key bindings from config.
func NewGlobalKeyMap(cfg *config.KeysConfig) GlobalKeyMap {
	if cfg == nil {
		cfg = &config.KeysConfig{}
	}
	return GlobalKeyMap{
		Quit: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Quit, "q", "ctrl+c")...),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Help, "?")...),
			key.WithHelp("?", "help"),
		),
		NextPane: key.NewBinding(
			key.WithKeys(parseKeys(cfg.NextPane, "tab")...),
			key.WithHelp("tab", "next pane"),
		),
		Pane1: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Pane1, "1")...),
			key.WithHelp("1", "tasks"),
		),
		Pane2: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Pane2, "2")...),
			key.WithHelp("2", "timer"),
		),
		Pane3: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Pane3, "3")...),
			key.WithHelp("3", "habits"),
		),
		Undo: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Undo, "ctrl+z", "u")...),
			key.WithHelp("ctrl+z", "undo"),
		),
		Redo: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Redo, "ctrl+y")...),
			key.WithHelp("ctrl+y", "redo"),
		),
	}
}

// =============================================================================
// Navigation Keys (shared by list-based panes)
// =============================================================================

// NavigationKeyMap defines keys for list navigation.
type NavigationKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Top    key.Binding
	Bottom key.Binding
}

// DefaultNavigationKeyMap returns the default navigation key bindings.
func DefaultNavigationKeyMap() NavigationKeyMap {
	return NewNavigationKeyMap(&config.KeysConfig{})
}

// NewNavigationKeyMap creates navigation key bindings from config.
func NewNavigationKeyMap(cfg *config.KeysConfig) NavigationKeyMap {
	if cfg == nil {
		cfg = &config.KeysConfig{}
	}
	return NavigationKeyMap{
		Up: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Up, "k", "up")...),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Down, "j", "down")...),
			key.WithHelp("j/↓", "down"),
		),
		Top: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Top, "g")...),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Bottom, "G")...),
			key.WithHelp("G", "bottom"),
		),
	}
}

// =============================================================================
// Input Keys (shared by text input fields)
// =============================================================================

// InputKeyMap defines keys for text input mode.
type InputKeyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

// DefaultInputKeyMap returns the default input key bindings.
func DefaultInputKeyMap() InputKeyMap {
	return NewInputKeyMap(&config.KeysConfig{})
}

// NewInputKeyMap creates input key bindings from config.
func NewInputKeyMap(cfg *config.KeysConfig) InputKeyMap {
	if cfg == nil {
		cfg = &config.KeysConfig{}
	}
	return InputKeyMap{
		Confirm: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Confirm, "enter")...),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys(parseKeys(cfg.Cancel, "esc")...),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// =============================================================================
// Task Pane Keys
// =============================================================================

// TaskKeyMap defines keys for the task pane.
type TaskKeyMap struct {
	Add    key.Binding
	Toggle key.Binding
	Delete key.Binding
	NavigationKeyMap
}

// DefaultTaskKeyMap returns the default task pane key bindings.
func DefaultTaskKeyMap() TaskKeyMap {
	return NewTaskKeyMap(&config.KeysConfig{})
}

// NewTaskKeyMap creates task key bindings from config.
func NewTaskKeyMap(cfg *config.KeysConfig) TaskKeyMap {
	if cfg == nil {
		cfg = &config.KeysConfig{}
	}
	return TaskKeyMap{
		Add: key.NewBinding(
			key.WithKeys(parseKeys(cfg.AddTask, "a")...),
			key.WithHelp("a", "add task"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(parseKeys(cfg.ToggleTask, "d", "enter", " ")...),
			key.WithHelp("d/space", "toggle done"),
		),
		Delete: key.NewBinding(
			key.WithKeys(parseKeys(cfg.DeleteTask, "x")...),
			key.WithHelp("x", "delete"),
		),
		NavigationKeyMap: NewNavigationKeyMap(cfg),
	}
}

// ShortHelp returns the short help for the task pane (implements help.KeyMap).
func (k TaskKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.Toggle, k.Delete, k.Down}
}

// FullHelp returns the full help for the task pane (implements help.KeyMap).
func (k TaskKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Add, k.Toggle, k.Delete},
		{k.Up, k.Down, k.Top, k.Bottom},
	}
}

// =============================================================================
// Timer Pane Keys
// =============================================================================

// TimerKeyMap defines keys for the timer pane.
type TimerKeyMap struct {
	Toggle key.Binding
	Switch key.Binding
	Stop   key.Binding
}

// DefaultTimerKeyMap returns the default timer pane key bindings.
func DefaultTimerKeyMap() TimerKeyMap {
	return NewTimerKeyMap(&config.KeysConfig{})
}

// NewTimerKeyMap creates timer key bindings from config.
func NewTimerKeyMap(cfg *config.KeysConfig) TimerKeyMap {
	if cfg == nil {
		cfg = &config.KeysConfig{}
	}
	return TimerKeyMap{
		Toggle: key.NewBinding(
			key.WithKeys(parseKeys(cfg.ToggleTimer, " ", "enter")...),
			key.WithHelp("space", "start/stop"),
		),
		Switch: key.NewBinding(
			key.WithKeys(parseKeys(cfg.SwitchTimer, "s")...),
			key.WithHelp("s", "switch project"),
		),
		Stop: key.NewBinding(
			key.WithKeys(parseKeys(cfg.StopTimer, "x")...),
			key.WithHelp("x", "stop"),
		),
	}
}

// ShortHelp returns the short help for the timer pane (implements help.KeyMap).
func (k TimerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Toggle, k.Switch, k.Stop}
}

// FullHelp returns the full help for the timer pane (implements help.KeyMap).
func (k TimerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Toggle, k.Switch, k.Stop},
	}
}

// =============================================================================
// Habits Pane Keys
// =============================================================================

// HabitKeyMap defines keys for the habits pane.
type HabitKeyMap struct {
	Add    key.Binding
	Toggle key.Binding
	Delete key.Binding
	NavigationKeyMap
}

// DefaultHabitKeyMap returns the default habit pane key bindings.
func DefaultHabitKeyMap() HabitKeyMap {
	return NewHabitKeyMap(&config.KeysConfig{})
}

// NewHabitKeyMap creates habit key bindings from config.
func NewHabitKeyMap(cfg *config.KeysConfig) HabitKeyMap {
	if cfg == nil {
		cfg = &config.KeysConfig{}
	}
	return HabitKeyMap{
		Add: key.NewBinding(
			key.WithKeys(parseKeys(cfg.AddHabit, "a")...),
			key.WithHelp("a", "add habit"),
		),
		Toggle: key.NewBinding(
			key.WithKeys(parseKeys(cfg.ToggleHabit, " ", "enter", "d")...),
			key.WithHelp("space", "toggle"),
		),
		Delete: key.NewBinding(
			key.WithKeys(parseKeys(cfg.DeleteHabit, "x")...),
			key.WithHelp("x", "delete"),
		),
		NavigationKeyMap: NewNavigationKeyMap(cfg),
	}
}

// ShortHelp returns the short help for the habit pane (implements help.KeyMap).
func (k HabitKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Add, k.Toggle, k.Delete, k.Down}
}

// FullHelp returns the full help for the habit pane (implements help.KeyMap).
func (k HabitKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Add, k.Toggle, k.Delete},
		{k.Up, k.Down, k.Top, k.Bottom},
	}
}

// =============================================================================
// Help Overlay Keys
// =============================================================================

// HelpKeyMap defines keys for the help overlay.
type HelpKeyMap struct {
	Close key.Binding
}

// DefaultHelpKeyMap returns the default help overlay key bindings.
func DefaultHelpKeyMap() HelpKeyMap {
	return HelpKeyMap{
		Close: key.NewBinding(
			key.WithKeys("?", "esc", "q", "enter", " "),
			key.WithHelp("any key", "close"),
		),
	}
}
