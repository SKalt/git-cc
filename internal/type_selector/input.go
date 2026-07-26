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
	opts := make([]single_select.ListItem, 0, cfg.CommitTypes.Len())
	for _, o := range config.Items(cfg.CommitTypes) {
		opts = append(opts, single_select.ListItem(o))
	}
	return Model{
		single_select.NewModel(
			config.Faint("select a commit type: "), cc.Type, opts,
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
func (m Model) ShouldSkip(currentValue string) (shouldSkip bool) {
	for opt := range m.input.Options() {
		if shouldSkip = opt == currentValue; shouldSkip {
			break
		}
	}
	return shouldSkip
}
