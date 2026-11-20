package ui

import (
	"errors"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ErrInputCanceled is returned by Input when the user cancels with Esc or Ctrl+C.
var ErrInputCanceled = errors.New("input canceled")

type inputModel struct {
	textInput textinput.Model
	value     string
	quitting  bool
	canceled  bool
}

func (m inputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.value = m.textInput.Value()
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.textInput.Width = msg.Width - 4
		if m.textInput.Width > 120 {
			m.textInput.Width = 120
		}
		if m.textInput.Width < 20 {
			m.textInput.Width = 20
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m inputModel) View() string {
	if m.quitting {
		return ""
	}
	return fmt.Sprintf("\n%s\n\n%s\n", m.textInput.Placeholder, m.textInput.View())
}

// Input displays a text input and returns the value, or ErrInputCanceled on Esc/Ctrl+C.
func Input(placeholder, defaultValue string) (string, error) {
	if !isInteractiveTTY() {
		return "", ErrNonInteractiveTTY
	}
	ti := textinput.New()
	ti.Placeholder = placeholder
	if defaultValue != "" {
		ti.SetValue(defaultValue)
	}
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 60

	m := inputModel{textInput: ti}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	fm := finalModel.(inputModel)
	if fm.canceled {
		return "", ErrInputCanceled
	}

	return fm.value, nil
}
