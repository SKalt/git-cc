package tui_single_select

import (
	"fmt"
	"strings"

	_ "sort"
	_ "strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	_ "github.com/muesli/reflow/ansi"
	_ "github.com/muesli/reflow/wordwrap"
	term "github.com/muesli/termenv"
)

type Model struct {
	Options         []string
	Hints           []string
	filtered        [][2]string
	matched         [][2]string
	Cursor          int
	maxOptionLength int

	Width  int // in runes
	Height int // in lines

	Choice    chan<- string
	textInput textinput.Model
}

func NewModel(prompt string, options []string, hints []string, choice chan<- string) Model {
	if len(options) != len(hints) {
		panic(fmt.Errorf("#opts : %d != %d : #hints", len(options), len(hints)))
	}
	textInputModel := textinput.NewModel()
	textInputModel.Placeholder = "type to select"
	textInputModel.Prompt = "   "
	textInputModel.Focus()

	result := Model{
		Options:   options,
		Hints:     hints,
		textInput: textInputModel,
		Choice:    choice,
	}
	result.matched, result.filtered = result.filter("")
	return result
}
func (m Model) Ready() bool {
	return len(m.matched) > 0
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

func Update(msg tea.Msg, model Model) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			close(model.Choice)
			return model, tea.Quit
		case tea.KeyEnter, tea.KeyTab:
			if len(model.matched) > 0 {
				model.Choice <- model.matched[model.Cursor][0]
				return model, tea.Quit
			} else {
				return model, textinput.Blink(model.textInput)
			}
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
		// TODO: reflow text here!
		return model, cmd
	}
	return model, cmd
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString(textinput.View(m.textInput) + "\n")

	for i, match := range m.matched {
		opt := match[0]
		padding := strings.Repeat(" ", m.maxOptLen()-len(opt))
		hint := match[1]
		if m.Cursor == i {
			s.WriteString(" > " + term.String(opt).Bold().Underline().String())
			s.WriteString(term.String(padding).Underline().String())
			s.WriteString(term.String(hint).Underline().String())
		} else {
			s.WriteString("   " + opt + padding)
			s.WriteString(term.String(hint).Faint().String())
		}
		s.WriteString("\n")
	}
	s.WriteString("\n")
	for _, rejected := range m.filtered {
		opt, hint := rejected[0], rejected[1]
		padding := strings.Repeat(" ", m.maxOptLen()-len(opt))
		s.WriteString("   " + term.String(opt+padding+hint).Faint().String())
		s.WriteString("\n")
	}

	s.WriteString("\n")

	s.WriteString(
		term.String("\n(tab/enter to select, up/down to navigate, Ctrl+C to quit)\n").Faint().String(),
	)

	return s.String()
}
