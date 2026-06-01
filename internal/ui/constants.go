package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Standard ANSI Colors to match legacy exactly
	Red     = lipgloss.Color("1")
	Green   = lipgloss.Color("2")
	Yellow  = lipgloss.Color("3")
	Blue    = lipgloss.Color("4")
	Purple  = lipgloss.Color("5")
	Cyan    = lipgloss.Color("6")
	Magenta = lipgloss.Color("5") // Same as Purple in ANSI
	Gray    = lipgloss.Color("8") // ANSI 90m is color 8

	// Styles
	BoldStyle = lipgloss.NewStyle().Bold(true)
	
	// Icons
	CheckIcon     = "✅"
	ErrorIcon     = "❌"
	InfoIcon      = "ℹ️ "
	WarningIcon   = "⚠️"
	TagIcon       = "🏷️"
	BulletIcon    = "•"
	RocketIcon    = "🚀"
	HammerIcon    = "🔨"
	HourglassIcon = "⏳"
	SwitchIcon    = "🔄"
)

// SuccessMessage returns a formatted success message
func SuccessMessage(msg string) string {
	return lipgloss.NewStyle().Foreground(Green).Render(CheckIcon + " " + msg)
}

// ErrorMessage returns a formatted error message
func ErrorMessage(msg string) string {
	return lipgloss.NewStyle().Foreground(Red).Render(ErrorIcon + " " + msg)
}

// InfoMessage returns a formatted info message
func InfoMessage(msg string) string {
	return lipgloss.NewStyle().Foreground(Blue).Render(InfoIcon + " " + msg)
}

// WarningMessage returns a formatted warning message
func WarningMessage(msg string) string {
	return lipgloss.NewStyle().Foreground(Yellow).Render(WarningIcon + " " + msg)
}

// Style helpers
func RedStyle() lipgloss.Style    { return lipgloss.NewStyle().Foreground(Red) }
func GreenStyle() lipgloss.Style  { return lipgloss.NewStyle().Foreground(Green) }
func YellowStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(Yellow) }
func BlueStyle() lipgloss.Style   { return lipgloss.NewStyle().Foreground(Blue) }
func CyanStyle() lipgloss.Style   { return lipgloss.NewStyle().Foreground(Cyan) }
func PurpleStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(Purple) }
func GrayStyle() lipgloss.Style   { return lipgloss.NewStyle().Foreground(Gray) }

// ClearScreen clears the terminal
func ClearScreen() {
	fmt.Print("\033[H\033[2J")
}
