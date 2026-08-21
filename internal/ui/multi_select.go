package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type multiChooseModel struct {
	title    string
	options  []string
	selected []bool
	cursor   int
	quitting bool
	canceled bool
}

func (m multiChooseModel) Init() tea.Cmd {
	return nil
}

func (m multiChooseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.options) - 1
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "space":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "enter":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.canceled = true
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m multiChooseModel) View() string {
	if m.quitting {
		return ""
	}
	s := ""
	if m.title != "" {
		s += titleStyle.Render(m.title) + "\n\n"
	}
	for i, opt := range m.options {
		cursorStr := "  "
		if i == m.cursor {
			cursorStr = "> "
		}

		checkedStr := "[ ]"
		if m.selected[i] {
			checkedStr = "[x]"
		}

		line := fmt.Sprintf("%s%s %s", cursorStr, checkedStr, opt)
		if i == m.cursor {
			s += selectedItemStyle.Render(line) + "\n"
		} else {
			s += itemStyle.Render(line) + "\n"
		}
	}
	s += timeoutStyle.Render("\n(Press <space> to toggle, <enter> to confirm, <esc> to cancel)") + "\n"
	return s
}

// MultiChoose displays a list of options allowing multi-selection using space, and returns a boolean slice of selections.
func MultiChoose(title string, options []string, preselected []bool) ([]bool, error) {
	selected := make([]bool, len(options))
	copy(selected, preselected)

	m := multiChooseModel{
		title:    title,
		options:  options,
		selected: selected,
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	fm := finalModel.(multiChooseModel)
	if fm.canceled {
		return nil, fmt.Errorf("selection canceled")
	}

	return fm.selected, nil
}
