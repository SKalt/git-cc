package tui_single_select

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/padding"
	"github.com/muesli/reflow/wordwrap"
	term "github.com/muesli/termenv"
)

type Model struct {
	Options         []string
	Hints           []string
	filtered        [][2]string
	matched         [][2]string
	Cursor          int
	maxOptionLength int
	context         string
	Width           int // in runes
	Height          int // in lines
	selected        string
	textInput       textinput.Model
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink(m.textInput)
}
func NewModel(context string, options []map[string]string) Model {
	values, hints := []string{}, []string{}
	for _, option := range options {
		for value, hint := range option {
			values, hints = append(values, value), append(hints, hint)
		}
	}
	textInputModel := textinput.NewModel()
	textInputModel.Placeholder = "type to select"
	textInputModel.Prompt = "   "
	textInputModel.Focus()
	// textInputModel.Focus() // must be done by the supervising component

	result := Model{
		context:   context,
		Options:   values,
		Hints:     hints,
		textInput: textInputModel,
	}
	result.matched, result.filtered = result.filter("")
	return result
}
func (m *Model) Focus() tea.Cmd {
	m.textInput.Focus()
	return textinput.Blink(m.textInput)
}

func (m Model) SetErr(err error) Model {
	m.textInput.Err = err
	return m
}

func (m Model) Focused() bool {
	return m.textInput.Focused()
}

func (m Model) Blur() {
	m.textInput.Blur()
}

func (m Model) maxOptLen() int {
	if m.maxOptionLength > 0 {
		return m.maxOptionLength
	}
	max := 0
	for _, opt := range m.Options {
		if len(opt) > max {
			max = len(opt)
		}
	}
	m.maxOptionLength = max
	return max
}

func (m Model) filter(startingWith string) ([][2]string, [][2]string) {
	matched, filtered := [][2]string{}, [][2]string{}
	for i, opt := range m.Options {
		hint := m.Hints[i]
		if len(opt) < len(startingWith) || opt[0:len(startingWith)] != startingWith {
			filtered = append(filtered, [2]string{opt, hint})
		} else {
			matched = append(matched, [2]string{opt, hint})
		}
	}
	return matched, filtered
}

func (m Model) Value() string {
	if len(m.matched) > 0 {
		return m.matched[m.Cursor][0]
	} else {
		return ""
	}
}

func Update(msg tea.Msg, model Model) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return model, tea.Quit
		case tea.KeyUp:
			if model.Cursor > 0 {
				model.Cursor -= 1
			} else {
				model.Cursor = len(model.matched) - 1
			}
			return model, textinput.Blink(model.textInput)
		case tea.KeyDown:
			if model.Cursor < len(model.matched) {
				model.Cursor += 1
			} else {
				model.Cursor = 0
			}
			return model, textinput.Blink(model.textInput)
		default:
			model.textInput, cmd = textinput.Update(msg, model.textInput)
			model.matched, model.filtered = model.filter(model.textInput.Value())
			model.Cursor = 0
			return model, cmd
		}

	case tea.WindowSizeMsg:
		model.Height = msg.Height
		model.Width = msg.Width
		return model, cmd
	}
	return model, cmd
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return Update(msg, m)
}

func wrapLine(left uint, hint string, right int, style func(string) term.Style) string {
	lines := strings.SplitN(wordwrap.String(hint, right), "\n", 2)
	result := style(lines[0]).String()
	if len(lines) > 1 {
		result += "\n" + indent.String(style(lines[1]).String(), left)
	}
	return result
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString(m.context + "\n")
	s.WriteString(textinput.View(m.textInput) + "\n")
	leftGutter := 3 // "   "
	maxOptLen := m.maxOptLen()
	leftColumn := (leftGutter + maxOptLen) + 1 // for the space
	rightColumn := m.Width - leftColumn
	for i, match := range m.matched {
		opt, hint := padding.String(match[0], uint(maxOptLen)), " "+match[1]
		if m.Cursor == i {
			style := func(str string) term.Style {
				return term.String(str).Underline()
			}
			s.WriteString(" > " + style(opt).Bold().String())
			s.WriteString(wrapLine(uint(leftColumn), hint, rightColumn, style))
		} else {
			style := func(str string) term.Style {
				return term.String(str).Faint()
			}
			s.WriteString("   " + opt)
			s.WriteString(wrapLine(uint(leftColumn), hint, rightColumn, style))
		}
		s.WriteString("\n")
	}
	// s.WriteString("\n")
	for _, rejected := range m.filtered {
		style := func(str string) term.Style {
			return term.String(str).Faint()
		}
		opt := style(padding.String(rejected[0], uint(maxOptLen))).String() + " "
		hint := rejected[1]
		s.WriteString("   " + opt)
		s.WriteString(wrapLine(uint(leftColumn), hint, rightColumn, style))
		s.WriteString("\n")
	}

	s.WriteString(
		term.String("\n(tab/enter to select, up/down to navigate, Ctrl+C to quit)\n").Faint().String(),
	)

	return s.String()
}
