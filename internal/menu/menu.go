package menu

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	cursor   int
	selected string
}

var items = []string{
	"Night Sky",
	"Space Shooter",
	"Matrix Rain",
	"Fireplace",
	"Snake",
	"Music Visualizer",
	"Exit",
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D5FFF"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFFF"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D5FFF")).
			Padding(1, 2)
)

func initialModel() model {
	return model{
		cursor: 0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(items)-1 {
				m.cursor++
			}

		case "enter":
			m.selected = items[m.cursor]
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	title := titleStyle.Render("✦ TERMFUN ✦")
	subtitle := helpStyle.Render("Terminal Playground")

	menu := ""

	for i, item := range items {
		cursor := "  "
		name := normalStyle.Render(item)

		if m.cursor == i {
			cursor = "❯ "
			name = selectedStyle.Render(item)
		}

		menu += fmt.Sprintf("%s%s\n", cursor, name)
	}

	help := helpStyle.Render("↑/↓ navigate • enter select • q quit")

	content := fmt.Sprintf(
		"%s\n%s\n\n%s\n%s",
		title,
		subtitle,
		menu,
		help,
	)

	return borderStyle.Render(content)
}

func Run() (string, error) {
	p := tea.NewProgram(initialModel())

	finalModel, err := p.Run()
	if err != nil {
		return "", err
	}

	m := finalModel.(model)

	return m.selected, nil
}
