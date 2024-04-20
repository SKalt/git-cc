package description_editor

// like what Glow has, but without the markdown-stashing

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/ansi"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/termenv"
	"github.com/skalt/git-cc/internal/config"
	"github.com/skalt/git-cc/internal/helpbar"
)

const prePrompt = "A short description of the changes:"

type Model struct {
	width       int
	input       textinput.Model // TODO: make input a pointer
	lengthLimit int
	helpBar     helpbar.Model
	prefix      string
}

func (m Model) SetPrefix(prefix string) Model {
	m.prefix = prefix
	m.input.Prompt = prefix
	return m
}
func (m Model) SetErr(err error) Model {
	m.input.Err = err
	return m
}
func (m Model) Focus() tea.Cmd {
	m.input.Focus()
	return nil
}
func (m Model) Value() string {
	return m.input.Value()
}

func NewModel(lengthLimit int, value string, enforced bool) Model {
	input := textinput.NewModel()
	input.SetValue(value)
	input.SetCursor(len(value))
	// input.Cursor = len(value)
	input.Prompt = config.Faint(prePrompt)
	if enforced {
		input.CharLimit = lengthLimit
	}
	input.Focus()
	return Model{
		lengthLimit: lengthLimit,
		input:       input,
		helpBar: helpbar.NewModel(
			config.HelpSubmit,
			config.HelpBack,
			config.HelpCancel,
		),
	}
}

// a styled length-counter, e.g. ( 9/80)
func viewCounter(m Model) string {
	current := len(m.prefix) + len(m.input.Value())
	paddedFormat := fmt.Sprintf(
		"(%%%dd/%d)", len(fmt.Sprintf("%d", m.lengthLimit)), m.lengthLimit,
	)
	view := fmt.Sprintf(paddedFormat, current)
	if current < m.lengthLimit {
		return config.Faint(view)
	} else if current == m.lengthLimit {
		return view // render in a warning color termenv.String(view).
	} else { // render in an alert color
		return termenv.String(view).Underline().String()
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			return m, tea.Quit
		default:
			m.input, cmd = m.input.Update(msg)
			m.input.Focus()
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.helpBar, cmd = m.helpBar.Update(msg)
		m.input.Width = msg.Width
		m.width = msg.Width
		return m, cmd
	default:
		m.input, _ = m.input.Update(msg)
		cmd = m.input.Focus()
		return m, cmd
	}
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString(wordwrap.String(config.Faint(prePrompt), m.width))
	s.WriteRune('\n')
	s.WriteRune('\n')
	s.WriteString(m.input.View())
	s.WriteRune('\n')
	s.WriteRune('\n')
	helpBar := m.helpBar.View()
	counter := viewCounter(m)
	s.WriteString(helpBar)

	helpBarLines := strings.Split(helpBar, "\n")
	last := helpBarLines[len(helpBarLines)-1]
	if ansi.PrintableRuneWidth(last)+ansi.PrintableRuneWidth(counter) >= m.width {
		s.WriteRune('\n')
	} else {
		s.WriteRune(' ')
	}
	s.WriteString(counter)
	return s.String()
}

func (m Model) Init() tea.Cmd {
	return nil // textinput.Blink(m.input)?
}
