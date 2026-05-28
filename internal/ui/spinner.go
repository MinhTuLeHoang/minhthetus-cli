package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type spinnerModel struct {
	spinner  spinner.Model
	title    string
	quitting bool
	err      error
}

func (m spinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	case error:
		m.err = msg
		return m, tea.Quit
	case quitMsg:
		m.quitting = true
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

type quitMsg struct{}

func (m spinnerModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}
	if m.quitting {
		return ""
	}
	return fmt.Sprintf("\n %s %s\n", m.spinner.View(), m.title)
}

// RunWithSpinner executes a function while showing a spinner.
func RunWithSpinner(title string, fn func() error) error {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	
	m := spinnerModel{spinner: s, title: title}
	p := tea.NewProgram(m)

	go func() {
		err := fn()
		if err != nil {
			p.Send(err)
		} else {
			p.Send(quitMsg{})
		}
	}()

	_, err := p.Run()
	return err
}
