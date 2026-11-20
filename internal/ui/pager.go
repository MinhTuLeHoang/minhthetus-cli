package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type pagerModel struct {
	title    string
	content  string
	ready    bool
	viewport viewport.Model
}

func (m pagerModel) Init() tea.Cmd {
	return nil
}

func (m pagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "enter", "ctrl+c":
			return m, tea.Quit
		case "home", "g":
			m.viewport.GotoTop()
			return m, nil
		case "end", "G":
			m.viewport.GotoBottom()
			return m, nil
		}

	case tea.WindowSizeMsg:
		verticalMarginHeight := 4
		vpHeight := msg.Height - verticalMarginHeight
		if vpHeight < 3 {
			vpHeight = 3
		}
		vpWidth := msg.Width
		if vpWidth <= 0 {
			vpWidth = 80
		}

		if !m.ready {
			m.viewport = viewport.New(vpWidth, vpHeight)
			m.viewport.YPosition = 2
			m.viewport.MouseWheelEnabled = true
			m.viewport.SetContent(m.content)
			m.viewport.GotoTop()
			m.ready = true
		} else {
			m.viewport.Width = vpWidth
			m.viewport.Height = vpHeight
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m pagerModel) View() string {
	if !m.ready {
		return "\n  Loading..."
	}

	header := BoldStyle.Render("🔍 " + m.title)
	dividerWidth := m.viewport.Width
	if dividerWidth <= 0 {
		dividerWidth = 60
	}
	divider := GrayStyle().Render(strings.Repeat("─", dividerWidth))

	percent := m.viewport.ScrollPercent() * 100
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	footer := GrayStyle().Render(fmt.Sprintf(
		" [%3.0f%%]  ↑/↓/mouse: scroll • PgUp/PgDn • q/Enter: exit",
		percent,
	))

	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s", header, divider, m.viewport.View(), divider, footer)
}

// ScrollView displays scrollable content in a full-screen viewport starting at the very top.
func ScrollView(title, content string) error {
	m := pagerModel{
		title:   title,
		content: content,
	}
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err := p.Run()
	return err
}
