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
	cursor      int // 0 for Yes/Agree, 1 for No/Cancel
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
		case "left", "h", "a":
			m.cursor = 0
			return m, nil
		case "right", "l", "d":
			m.cursor = 1
			return m, nil
		case "tab":
			m.cursor = 1 - m.cursor
			return m, nil
		case "y", "Y":
			m.confirmed = true
			m.quitting = true
			return m, tea.Quit
		case "n", "N", "esc":
			m.confirmed = false
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.confirmed = (m.cursor == 0)
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

	activeButtonStyle = lipgloss.NewStyle().
		Bold(true).
		Padding(0, 3)

	inactiveButtonStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244")).
		Background(lipgloss.Color("236")).
		Padding(0, 3)
)

func (m confirmModel) View() string {
	if m.quitting {
		return ""
	}

	s := fmt.Sprintf("%s\n\n", titleStyle.Render(m.prompt))

	var yesBtn, noBtn string
	if m.cursor == 0 {
		yesBtn = activeButtonStyle.Background(lipgloss.Color("2")).Foreground(lipgloss.Color("0")).Render(" Yes ")
		noBtn = inactiveButtonStyle.Render(" No ")
	} else {
		yesBtn = inactiveButtonStyle.Render(" Yes ")
		noBtn = activeButtonStyle.Background(lipgloss.Color("1")).Foreground(lipgloss.Color("0")).Render(" No ")
	}

	s += fmt.Sprintf("  %s    %s", yesBtn, noBtn)

	if m.timeout > 0 {
		remaining := int(m.timeout.Seconds() - time.Since(m.startTime).Seconds())
		if remaining < 0 {
			remaining = 0
		}
		s += fmt.Sprintf("\n\n%s", timeoutStyle.Render(fmt.Sprintf("(Auto-%s in %ds...)", 
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
		cursor:      0, // default focus on Yes
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return false, err
	}

	return finalModel.(confirmModel).confirmed, nil
}
