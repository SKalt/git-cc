package tui_breaking_change_input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/skalt/git-cc/pkg/config"
)

type Model struct {
	input textinput.Model
}

func (m Model) Value() string {
	return m.input.Value()
}

func (m Model) View() string {
	return textinput.View(m.input) + "\n\n" + config.HelpBar
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = textinput.Update(msg, m.input)
	return m, cmd
}

func NewModel() Model {
	input := textinput.NewModel()
	input.Prompt = termenv.String("Breaking changes: ").Faint().String()
	input.Placeholder = "if any."
	input.Focus()
	return Model{input}
}
