package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type confirmModel struct {
	prompt      string
	timeout     time.Duration
	startTime   time.Time
	quitting    bool
	confirmed   bool
	autoApprove bool
}

type tickMsg time.Time

func (m confirmModel) Init() tea.Cmd {
	if m.timeout > 0 {
		return tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return nil
}

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			m.confirmed = true
			m.quitting = true
			return m, tea.Quit
		case "n", "N", "esc":
			m.confirmed = false
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c":
			m.confirmed = false
			m.quitting = true
			return m, tea.Quit
		}

	case tickMsg:
		if time.Since(m.startTime) >= m.timeout {
			m.confirmed = m.autoApprove
			m.quitting = true
			return m, tea.Quit
		}
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	return m, nil
}

var (
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	timeoutStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)
)

func (m confirmModel) View() string {
	if m.quitting {
		return ""
	}

	s := fmt.Sprintf("%s %s", titleStyle.Render(m.prompt), lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("[y/n]"))
	
	if m.timeout > 0 {
		remaining := int(m.timeout.Seconds() - time.Since(m.startTime).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		s += fmt.Sprintf("\n%s", timeoutStyle.Render(fmt.Sprintf("(Auto-%s in %ds...)", 
			func() string { if m.autoApprove { return "approving" }; return "cancelling" }(), 
			remaining)))
	}
	
	return s
}

// Confirm asks the user for confirmation with optional timeout and auto-approve.
func Confirm(prompt string, timeout time.Duration, autoApprove bool) (bool, error) {
	m := confirmModel{
		prompt:      prompt,
		timeout:     timeout,
		startTime:   time.Now(),
		autoApprove: autoApprove,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	return finalModel.(confirmModel).confirmed, nil
}
