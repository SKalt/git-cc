package type_selector

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/skalt/git-cc/pkg/config"
	"github.com/skalt/git-cc/pkg/parser"
	"github.com/skalt/git-cc/pkg/single_select"
)

var helpBar = config.HelpBar(
	config.HelpSubmit, config.HelpSelect, config.HelpCancel,
)

type Model struct {
	input single_select.Model
}

func NewModel(cc *parser.CC, cfg config.Cfg) Model {
	return Model{
		single_select.NewModel(
			config.Faint("select a commit type: "), cc.Type, cfg.CommitTypes,
			single_select.MatchStart,
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
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) ShouldSkip(currentValue string) bool {
	for _, opt := range m.input.Options {
		if opt == currentValue {
			return true
		}
	}
	return false
}
