package description_editor

// like what Glow has, but without the markdown-stashing

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/skalt/git-cc/pkg/config"
)

const prePrompt = "A short description of the changes:\n\n"

type Model struct {
	prefix      string
	prefixLen   int
	input       textinput.Model
	lengthLimit int // TODO: make *int and use nil to eliminate countdown
}

func (m Model) SetPrefix(prefix string) Model {
	m.prefixLen = len(prefix)
	m.input.Prompt = termenv.String(prePrompt).Faint().String() + prefix
	return m
}
func (m Model) SetErr(err error) Model {
	m.input.Err = err
	return m
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

func NewModel(lengthLimit int, value string, enforced bool) Model {
	input := textinput.NewModel()
	input.SetValue(value)
	input.SetCursor(len(value))
	// input.Cursor = len(value)
	input.Prompt = termenv.String(prePrompt).Faint().String()
	if enforced {
		input.CharLimit = lengthLimit
	}
	input.Focus()
	return Model{
		prefixLen:   0,
		lengthLimit: lengthLimit,
		input:       input,
	}
}

func viewCounter(m Model) string {
	current := m.prefixLen + len(m.input.Value())
	paddedFormat := fmt.Sprintf("(%%%dd/%d)", len(fmt.Sprintf("%d", m.lengthLimit)), m.lengthLimit)
	view := fmt.Sprintf(paddedFormat, current)
	if current < m.lengthLimit {
		return termenv.String(view).Faint().String()
	} else if current == m.lengthLimit {
		return view // render in a warning color termenv.String(view).
	} else {
		return termenv.String(view).Underline().String() // render in an alert color
	}
}

func viewHelpBar(m Model) string {
	return fmt.Sprintf("\n%s %s", config.HelpBar, viewCounter(m))
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
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
	return textinput.View(m.input) + "\n" + viewHelpBar(m)
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink(m.input)
}
