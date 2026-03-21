package helpbar

import (
	"io"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/skalt/git-cc/internal/controls"
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

func (m Model) Render(s io.StringWriter) {
	var h help.Model = help.New()
	v := h.ShortHelpView(
		[]key.Binding{
			controls.Keymap.Back,
			controls.Keymap.Next,
			controls.Keymap.Cancel,
		},
	)
	s.WriteString(v)
	// if len(m.items) == 0 {
	// 	return
	// }
	// item, items := m.items[0], m.items[1:]

	// _ = utils.Must(s.WriteString(config.Faint(item)))
	// currentLen := ansi.PrintableRuneWidth(item)

	// sep, sepLen := config.Faint("; "), 2 // 2 == len(sep)
	// for _, item := range items {
	// 	if currentLen+sepLen+ansi.PrintableRuneWidth(item) <= m.width {
	// 		_ = utils.Must(s.WriteString(sep))
	// 		_ = utils.Must(s.WriteString(config.Faint(item)))
	// 		currentLen += sepLen + len(item)
	// 	} else {
	// 		_ = utils.Must(s.WriteString("\n"))
	// 		currentLen = utils.Must(s.WriteString(config.Faint(item)))
	// 	}
	// }
}

func (m Model) View() string {
	b := strings.Builder{}
	m.Render(&b)
	return b.String()
}
