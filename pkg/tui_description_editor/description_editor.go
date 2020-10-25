package tui_description_editor

// like what Glow has, but without the markdown-stashing

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/glamour"
	"github.com/muesli/termenv"
)

type Model struct {
	prefix    string
	prefixLen int
	input     textinput.Model
	softMax   int // TODO: make *int and use nil to eliminate countdown
}

func (m Model) SetPrefix(prefix string) {
	m.input.Prompt = termenv.String(prefix).Faint().String()
}
func (m Model) Focus() tea.Cmd {
	m.input.Focus()
	return textinput.Blink(m.input)
}
func (m Model) Blur() {
	m.input.Blur()
}
func (m Model) Value() string {
	return m.input.Value()
}

func NewModel(prefix string, softLengthLimit int) Model {
	input := textinput.NewModel()
	input.Prompt = termenv.String(prefix).Faint().String()
	input.Focus()
	return Model{
		prefixLen: len(prefix),
		softMax:   softLengthLimit,
		input:     input,
	}
}

func viewCounter(m Model) string {
	current := m.prefixLen + len(m.input.Value())
	paddedFormat := fmt.Sprintf("(%%%dd/%d)", len(fmt.Sprintf("%d", m.softMax)), m.softMax)
	view := fmt.Sprintf(paddedFormat, current)
	if current < m.softMax {
		return termenv.String(view).Faint().String()
	} else if current == m.softMax {
		return view // render in a warning color termenv.String(view).
	} else {
		return termenv.String(view).Underline().String() // render in an alert color
	}
}

func viewHelpBar(m Model) string {
	help := termenv.String("\nsubmit: enter; go back: shift+tab; cancel: ctrl+c").
		Faint().String() + ""
	return help + viewCounter(m)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		// case tea.KeyEnter: // handled by Value()
		// 	m.Choice <- m.input.Value()
		// 	return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyCtrlD:
			// close(m.Choice)
			return m, tea.Quit
		default:
			m.input, cmd = textinput.Update(msg, m.input)
			m.input.Focus()
			return m, cmd
		}
	default:
		m.input, cmd = textinput.Update(msg, m.input)
		m.input.Focus()
		return m, cmd
	}
}

func (m Model) View() string {
	return textinput.View(m.input) + viewHelpBar(m)
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink(m.input)
}
