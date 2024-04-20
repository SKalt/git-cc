package single_select

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/padding"
	"github.com/muesli/reflow/wordwrap"
	term "github.com/muesli/termenv"
)

type Model struct {
	Options  []string
	Hints    []string
	match    func(*Model, string, string) bool
	filtered [][2]string
	matched  [][2]string
	// the offset (0-indexed) of the selection within the options
	Cursor          int
	maxOptionLength int
	context         string
	// in runes
	Width int
	// in lines
	Height    int
	textInput textinput.Model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func NewModel(context string, value string, options []string, hints []string, match func(*Model, string, string) bool) Model {
	if len(options) != len(hints) {
		panic(fmt.Errorf("len(hints) %d != %d len(options)", len(hints), len(options)))
	}
	input := textinput.NewModel()
	input.Placeholder = "type to select"
	input.Prompt = "   "
	input.SetValue(value)
	input.SetCursor(len(value))
	input.Focus()

	result := Model{
		context:   context,
		Options:   options,
		Hints:     hints,
		textInput: input,
		match:     match,
	}
	result.matched, result.filtered = result.filter(value)
	return result
}

func (m *Model) Focus() tea.Cmd {
	return m.textInput.Focus()
}

func (m *Model) Match(query string, option string) bool {
	return m.match(m, query, option)
}

// register an error with the textInput sub-component
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
	return m.maxOptionLength
}

func MatchStart(m *Model, query string, option string) bool {
	return len(query) <= len(option) && option[0:len(query)] == query
}

func (m Model) filter(startingWith string) ([][2]string, [][2]string) {
	matched, filtered := [][2]string{}, [][2]string{}
	for i, opt := range m.Options {
		hint := m.Hints[i]
		if m.Match(startingWith, opt) {
			matched = append(matched, [2]string{opt, hint})
		} else {
			filtered = append(filtered, [2]string{opt, hint})
		}
	}
	return matched, filtered
}

// access the matched, selected value. If no value is matched, this returns "".
func (m Model) Value() string {
	if len(m.matched) > 0 {
		return m.matched[m.Cursor][0]
	} else {
		return ""
	}
}

func (m Model) CurrentInput() string {
	return m.textInput.Value()
}

func Update(msg tea.Msg, model Model) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return model, tea.Quit
		case tea.KeyUp, tea.KeyCtrlP:
			if model.Cursor > 0 {
				model.Cursor -= 1
			} else {
				model.Cursor = len(model.matched) - 1
			}
			return model, cmd
		case tea.KeyDown, tea.KeyCtrlN:
			if model.Cursor < len(model.matched)-1 {
				model.Cursor += 1
			} else {
				model.Cursor = 0
			}
			return model, cmd
		default:
			model.textInput, cmd = model.textInput.Update(msg)
			model.matched, model.filtered = model.filter(model.textInput.Value())
			model.Cursor = 0
			return model, cmd
		}
	case tea.MouseEvent:
		switch msg.Type {
		case tea.MouseWheelUp:
			if model.Cursor > 0 {
				model.Cursor -= 1
			} else {
				model.Cursor = len(model.matched) - 1
			}
			return model, cmd
		case tea.MouseWheelDown:
			if model.Cursor < len(model.matched)-1 {
				model.Cursor += 1
			} else {
				model.Cursor = 0
			}
			return model, cmd
		default:
			model.textInput, cmd = model.textInput.Update(msg)
			model.matched, model.filtered = model.filter(model.textInput.Value())
			model.Cursor = 0
			return model, cmd
		}
	case tea.WindowSizeMsg:
		model.Height = msg.Height
		model.Width = msg.Width
		model.textInput, cmd = model.textInput.Update(msg)
		return model, cmd
	default:
		model.textInput, cmd = model.textInput.Update(msg)
		model.matched, model.filtered = model.filter(model.textInput.Value())
		return model, cmd
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return Update(msg, m)
}

// wrap `text` between column `left` and `right`, applying `style`
func wrapLine(
	left uint,
	text string,
	right int,
	style func(string) term.Style,
) string {
	lines := strings.SplitN(wordwrap.String(text, right), "\n", 2)
	result := style(lines[0]).String()
	if len(lines) > 1 {
		result += "\n" + indent.String(style(lines[1]).String(), left)
	}
	return result
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString(m.context + "\n")
	s.WriteString(m.textInput.View() + "\n")
	leftGutter := 3 // "   "
	maxOptLen := m.maxOptLen()
	leftColumn := (leftGutter + maxOptLen) + 1 // for the space
	rightColumn := m.Width - leftColumn
	pad := func(opt string, max int) string {
		s := padding.String(opt, uint(maxOptLen))
		if s == "" {
			s = strings.Repeat(" ", max)
		}
		return s
	}
	for i, match := range m.matched {
		opt, hint := pad(match[0], maxOptLen), " "+match[1]
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
	style := func(str string) term.Style {
		return term.String(str).Faint()
	}
	for _, rejected := range m.filtered {
		opt, hint := style(pad(rejected[0], maxOptLen)).String(), rejected[1]
		s.WriteString("   " + opt + " ")
		s.WriteString(wrapLine(uint(leftColumn), hint, rightColumn, style))
		s.WriteString("\n")
	}

	return s.String()
}
