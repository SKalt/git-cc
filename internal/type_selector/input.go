package type_selector

import (
	"io"

	tea "charm.land/bubbletea/v2"
	"github.com/skalt/git-cc/internal/config"
	"github.com/skalt/git-cc/internal/helpbar"
	"github.com/skalt/git-cc/internal/single_select"
	"github.com/skalt/git-cc/internal/utils"
	"github.com/skalt/git-cc/pkg/parser"
)

type Model struct {
	input   single_select.Model
	helpBar helpbar.Model
}

func NewModel(cc *parser.CC, cfg *config.Cfg) Model {
	types, hints := config.ZippedOrderedKeyValuePairs(cfg.CommitTypes)
	return Model{
		single_select.NewModel(
			config.Faint("select a commit type: "),
			cc.Type,
			types, hints,
		),
		helpbar.NewModel(
			config.HelpSubmit, config.HelpSelect, config.HelpCancel,
		),
	}
}

func (m Model) Value() string {
	return m.input.Value()
}

func (m Model) Render(s io.StringWriter) {
	m.input.Render(s)
	_ = utils.Must(s.WriteString("\n"))
	m.helpBar.Render(s)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.helpBar, _ = m.helpBar.Update(msg)
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// whether this component should be skipped (during backtracking for error correction?)
func (m Model) ShouldSkip(currentValue string) bool {
	for _, opt := range m.input.Options() {
		if opt == currentValue {
			return true
		}
	}
	return false
}
