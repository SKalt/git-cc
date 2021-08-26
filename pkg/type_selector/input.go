package type_selector

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/skalt/git-cc/pkg/config"
	"github.com/skalt/git-cc/pkg/helpbar"
	"github.com/skalt/git-cc/pkg/parser"
	"github.com/skalt/git-cc/pkg/single_select"
)

type Model struct {
	input   single_select.Model
	helpBar helpbar.Model
}

func NewModel(cc *parser.CC, cfg config.Cfg) Model {
	return Model{
		single_select.NewModel(
			config.Faint("select a commit type: "), cc.Type, cfg.CommitTypes,
			single_select.MatchStart,
		),
		helpbar.NewModel(
			config.HelpSubmit, config.HelpSelect, config.HelpCancel,
		),
	}
}

func (m Model) Value() string {
	return m.input.Value()
}

func (m Model) View() string {
	s := strings.Builder{}
	s.WriteString(m.input.View())
	s.WriteRune('\n')
	s.WriteString(m.helpBar.View())
	return s.String()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.helpBar, _ = m.helpBar.Update(msg)
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// whether this component should be skipped (during backtracking for error correction?)
func (m Model) ShouldSkip(currentValue string) bool {
	for _, opt := range m.input.Options {
		if opt == currentValue {
			return true
		}
	}
	return false
}
