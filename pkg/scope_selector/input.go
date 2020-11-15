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

// the method for determining if the current input matches an option.
func match(m *single_select.Model, query string, option string) bool {
	if option == "new scope" {
		for _, opt := range m.Options {
			if query == opt {
				return false
			}
		}
		return true
	} else {
		return single_select.MatchStart(m, query, option)
	}
}

// given options from config, add the leading "unscoped" and trailing "new scope" options
func makeOptions(options []map[string]string) []map[string]string {
	return append(append(
		[]map[string]string{{"": "unscoped; affects the entire project"}},
		options...,
	), map[string]string{"new scope": "edit a new scope into your configuration file"})
}

// should return two slices of string of equal size.
func makeOptHintPair(options []map[string]string) ([]string, []string) {
	values, hints := []string{}, []string{}
	for _, option := range options {
		for value, hint := range option {
			values = append(values, value)
			hints = append(hints, hint)
		}
	}
	return values, hints
}

func NewModel(cc *parser.CC, cfg config.Cfg) Model {
	return Model{
		single_select.NewModel(
			config.Faint("select a scope:"),
			cc.Scope,
			makeOptions(cfg.Scopes),
			match,
		),
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyTab:
			if m.Value() == "new scope" {
				cfg := config.EditCfgFile(config.CentralStore)
				values, hints := makeOptHintPair(makeOptions(cfg.Scopes))
				m.input.Options = values
				m.input.Hints = hints
				if m.input.Cursor >= len(m.input.Options) {
					m.input.Cursor = len(m.input.Options) - 1
				}
				return m, cmd
			} else {
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}
	}
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
