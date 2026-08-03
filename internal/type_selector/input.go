package type_selector

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/skalt/git-cc/internal/config"
	"github.com/skalt/git-cc/internal/controls"
	"github.com/skalt/git-cc/internal/single_select"
	"github.com/skalt/git-cc/internal/utils"
	"github.com/skalt/git-cc/pkg/parser"
)

type Model struct {
	input single_select.Model
}

var _ controls.InputComponent = Model{}

func NewModel(cc *parser.CC, cfg *config.Cfg) (m Model) {
	opts := make([]single_select.ListItem, 0, cfg.CommitTypes.Len())
	for _, o := range config.Items(cfg.CommitTypes) {
		opts = append(opts, single_select.ListItem(o))
	}
	m.input = single_select.NewModel(
		config.Faint("select a commit type: "), cc.Type, opts,
	)
	return m
}

func (m Model) Value() string {
	return m.input.Value()
}

func (m Model) Render(s *strings.Builder) {
	m.input.Render(s)
	utils.Must(s.WriteString("\n"))
}

func (m Model) Update(msg tea.Msg) (w Model, cmd tea.Cmd) {
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// whether this component should be skipped (during backtracking for error correction?)
func (m Model) ShouldSkip() (shouldSkip bool) {
	val := m.Value()
	for opt := range m.input.Options() {
		if shouldSkip = opt == val; shouldSkip {
			break
		}
	}
	return shouldSkip
}

func (m Model) Ready() bool { return m.Value() != "" }
