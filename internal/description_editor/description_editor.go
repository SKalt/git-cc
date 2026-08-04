package description_editor

// like what Glow has, but without the markdown-stashing

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/skalt/git-cc/internal/config"
	"github.com/skalt/git-cc/internal/controls"
	"github.com/skalt/git-cc/internal/utils"
	"github.com/skalt/git-cc/pkg/parser"
)

const header = "A short description of the changes:"

type Model struct {
	input       textinput.Model
	lengthLimit int
}

var _ controls.InputComponent = Model{}

func (m Model) SetPrefix(prefix string) Model {
	m.input.Prompt = config.Faint(prefix)
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

var keymap = struct {
	next, back, cancel key.Binding
}{next: controls.Keymap.Next, back: controls.Keymap.Back, cancel: controls.Keymap.Cancel}

func NewModel(cc *parser.CC, lengthLimit int, enforced bool) Model {
	var value string
	if cc != nil {
		value = cc.Description
	}
	input := textinput.New()
	input.SetValue(value)
	input.SetCursor(len(value))
	if enforced {
		input.CharLimit = lengthLimit
	}
	input.Focus()
	return Model{
		lengthLimit: lengthLimit,
		input:       input,
	}
}

// a styled length-counter, e.g. ( 9/80)
func viewCounter(m Model) string {
	current := len(m.input.Prompt) + len(m.input.Value())
	paddedFormat := fmt.Sprintf(
		"(%%%dd/%d)", len(fmt.Sprintf("%d", m.lengthLimit)), m.lengthLimit,
	)
	view := fmt.Sprintf(paddedFormat, current)
	if current < m.lengthLimit {
		return config.Faint(view)
	} else if current == m.lengthLimit {
		return view // render in a warning color termenv.String(view).
	} else { // render in an alert color
		return lipgloss.NewStyle().Underline(true).Render(view)
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		k := msg.Key()
		switch {
		case key.Matches(k, keymap.cancel):
			return m, tea.Quit
		default:
			m.input, cmd = m.input.Update(msg)
			m.input.Focus()
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.input.SetWidth(msg.Width)
		return m, cmd
	default:
		cmds := make([]tea.Cmd, 2)
		m.input, cmds[0] = m.input.Update(msg)
		cmds[1] = m.input.Focus()
		return m, tea.Batch(cmds...)
	}
}

func (m Model) Render(s *strings.Builder) {
	style := lipgloss.NewStyle().Width(m.input.Width()).Faint(true)
	utils.Must(s.WriteString(style.Render(header, viewCounter(m))))
	utils.Must(s.WriteString("\n\n"))
	val := m.input.View()
	utils.Must(s.WriteString(val))
	utils.Must(s.WriteString("\n\n"))
}

func (m Model) Init() tea.Cmd {
	return nil // textinput.Blink(m.input)?
}

func (m Model) Ready() bool { return m.input.Value() != "" }
