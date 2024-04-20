package helpbar

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/ansi"
	"github.com/skalt/git-cc/pkg/config"
)

type Model struct {
	// each item should already be joined with an ":", e.g. "foo: bar"
	items []string
	width int
}

func NewModel(items ...string) Model {
	return Model{items, 0}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, cmd
	}
	return m, cmd
}

func (m Model) View() string {
	if len(m.items) == 0 {
		return ""
	}
	item, items := m.items[0], m.items[1:]

	s := strings.Builder{}
	s.WriteString(config.Faint(item))
	currentLen := ansi.PrintableRuneWidth(item)

	sep, sepLen := config.Faint("; "), 2 // 2 == len(sep)
	for _, item := range items {
		if currentLen+sepLen+ansi.PrintableRuneWidth(item) <= m.width {
			s.WriteString(sep)
			s.WriteString(config.Faint(item))
			currentLen += sepLen + len(item)
		} else {
			s.WriteRune('\n')
			currentLen, _ = s.WriteString(config.Faint(item))
		}
	}
	return s.String()
}
