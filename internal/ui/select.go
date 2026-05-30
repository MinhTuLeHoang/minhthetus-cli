package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("212")).Bold(true)
)

type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	if index == m.Index() {
		fmt.Fprint(w, selectedItemStyle.Render("> "+str[2:]))
	} else {
		fmt.Fprint(w, itemStyle.Render(str))
	}
}

type selectModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

type simpleChooseModel struct {
	title    string
	options  []string
	cursor   int
	quitting bool
	choice   string
}

func (m simpleChooseModel) Init() tea.Cmd {
	return nil
}

func (m simpleChooseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "enter":
			m.choice = m.options[m.cursor]
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m simpleChooseModel) View() string {
	if m.quitting {
		return ""
	}
	s := ""
	if m.title != "" {
		s += titleStyle.Render(m.title) + "\n\n"
	}
	for i, opt := range m.options {
		if i == m.cursor {
			s += selectedItemStyle.Render("> " + opt) + "\n"
		} else {
			s += itemStyle.Render("  " + opt) + "\n"
		}
	}
	return s
}

// Choose displays a list of options and returns the selected one.
// It automatically switches to a compact inline layout for 7 or fewer options to avoid scrolling issues.
func Choose(title string, options []string) (string, error) {
	if len(options) <= 7 {
		m := simpleChooseModel{
			title:   title,
			options: options,
		}
		p := tea.NewProgram(m)
		finalModel, err := p.Run()
		if err != nil {
			return "", err
		}
		return finalModel.(simpleChooseModel).choice, nil
	}

	items := make([]list.Item, len(options))
	for i, opt := range options {
		items[i] = item(opt)
	}

	l := list.New(items, itemDelegate{}, 20, 10)
	l.Title = titleStyle.Render(title)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	m := selectModel{list: l}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	return finalModel.(selectModel).choice, nil
}
