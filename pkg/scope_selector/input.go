package scope_selector

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/skalt/git-cc/pkg/config"
	"github.com/skalt/git-cc/pkg/parser"
	"github.com/skalt/git-cc/pkg/single_select"
)

var helpBar = config.HelpBar(
	config.HelpSubmit, config.HelpSelect, config.HelpBack, config.HelpCancel,
)

type Model struct {
	input single_select.Model
}

func NewModel(cc *parser.CC, cfg config.Cfg) Model {
	return Model{
		single_select.NewModel(
			config.Faint("select a scope:"),
			cc.Scope,
			append(
				[]map[string]string{{"": "unscoped; affects the entire project"}},
				cfg.Scopes...,
			),
		), // TODO: Option to add new scope?
	}
}

func (m Model) Value() string {
	return m.input.Value()
}

func (m Model) View() string {
	return m.input.View() + helpBar
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) ShouldSkip(currentValue string) bool {
	if len(m.input.Options) == 0 {
		return true
	}
	for _, opt := range m.input.Options {
		if currentValue == opt && opt != "" {
			return true
		}
	}
	return false
}
