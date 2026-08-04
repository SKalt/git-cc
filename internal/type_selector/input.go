package type_selector

import (
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/skalt/git-cc/internal/config"
	"github.com/skalt/git-cc/internal/controls"
	"github.com/skalt/git-cc/internal/single_select"
	"github.com/skalt/git-cc/internal/utils"
	"github.com/skalt/git-cc/pkg/parser"
)

type Model struct {
	input  single_select.Model
	logger *slog.Logger
}

// FullHelp implements [help.KeyMap].
func (m Model) FullHelp() [][]key.Binding { return m.input.FullHelp() }

// ShortHelp implements [help.KeyMap].
func (m Model) ShortHelp() []key.Binding { return m.input.ShortHelp() }

var (
	_ controls.InputComponent = Model{}
	_ help.KeyMap             = Model{}
)

func NewModel(cc *parser.CC, cfg *config.Cfg) (m Model) {
	m.logger = cfg.Logger.With("name", "type_selector")
	opts := make([]single_select.ListItem, 0, cfg.CommitTypes.Len())
	for _, o := range config.Items(cfg.CommitTypes) {
		opts = append(opts, single_select.ListItem(o))
	}
	m.input = single_select.NewModel(
		config.Faint("select a commit type: "), cc.Type, opts, cfg.Logger,
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
func (m Model) Ready() bool {
	val := m.Value()
	logger := m.logger.With("value", val)
	for opt := range m.input.Options() {
		if opt == val {
			logger.Debug("ready")
			return true
		}
	}
	logger.Debug("not ready")
	return false
}
