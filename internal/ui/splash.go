package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Splash returns the colorful splash header.
func Splash() string {
	colors := []lipgloss.AdaptiveColor{
		lipgloss.AdaptiveColor{Light: "#5F5FFF", Dark: "#5F5FFF"}, // Blue
		lipgloss.AdaptiveColor{Light: "#875FFF", Dark: "#875FFF"}, // Purple
		lipgloss.AdaptiveColor{Light: "#D787FF", Dark: "#D787FF"}, // Pink
		lipgloss.AdaptiveColor{Light: "#FF5F87", Dark: "#FF5F87"}, // Red-ish
		lipgloss.AdaptiveColor{Light: "#FF875F", Dark: "#FF875F"}, // Orange
		lipgloss.AdaptiveColor{Light: "#FFAF5F", Dark: "#FFAF5F"}, // Yellow-ish
		lipgloss.AdaptiveColor{Light: "#D7FF5F", Dark: "#D7FF5F"}, // Lime
		lipgloss.AdaptiveColor{Light: "#87FF5F", Dark: "#87FF5F"}, // Green
		lipgloss.AdaptiveColor{Light: "#5FFF87", Dark: "#5FFF87"}, // Teal
		lipgloss.AdaptiveColor{Light: "#5FFFFF", Dark: "#5FFFFF"}, // Cyan
	}

	text := "MINH THE TUS CLI"
	words := strings.Fields(text)
	var rendered []string

	colorIdx := 0
	for _, word := range words {
		var wordRendered []string
		for _, char := range word {
			style := lipgloss.NewStyle().Foreground(colors[colorIdx%len(colors)]).Bold(true)
			wordRendered = append(wordRendered, style.Render(string(char)))
			colorIdx++
		}
		rendered = append(rendered, strings.Join(wordRendered, " "))
	}

	star := lipgloss.NewStyle().Foreground(Purple).Render("✦")
	return fmt.Sprintf("\n  %s  %s  %s\n", star, strings.Join(rendered, "   "), star)
}
