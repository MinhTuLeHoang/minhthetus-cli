package ui

import (
	"strings"
	"time"
)

// GumChoose opens a choose menu using native TUI
func GumChoose(options ...string) string {
	choice, err := Choose("Select an option:", options)
	if err != nil {
		return ""
	}
	return choice
}

// GumConfirm opens a confirm dialog using native TUI
func GumConfirm(prompt string) bool {
	confirmed, err := Confirm(prompt, 0, false)
	if err != nil {
		return false
	}
	return confirmed
}

// GumConfirmTimeout opens a confirm dialog with a timeout using native TUI
func GumConfirmTimeout(prompt, timeout string) bool {
	dur := parseDuration(timeout)
	confirmed, err := Confirm(prompt, dur, true)
	if err != nil {
		return false
	}
	return confirmed
}

// GumInput opens an input field using native TUI
func GumInput(placeholder, value string) string {
	res, err := Input(placeholder, value)
	if err != nil {
		return ""
	}
	return res
}

// GumFilter opens a filter menu using native TUI
func GumFilter(options []string, placeholder string) string {
	choice, err := Choose(placeholder, options)
	if err != nil {
		return ""
	}
	return choice
}

// Helper to parse duration strings like "2s" or simple numbers "2" to time.Duration
func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	// If it's a raw number, assume seconds
	if !strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "m") && !strings.HasSuffix(s, "h") && !strings.HasSuffix(s, "ms") {
		s += "s"
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}
