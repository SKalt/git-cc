package scope_selector

import (
	"fmt"
	"io"
	"log"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/skalt/git-cc/internal/config"
	"github.com/skalt/git-cc/internal/helpbar"
	"github.com/skalt/git-cc/internal/single_select"
	"github.com/skalt/git-cc/internal/utils"
	"github.com/skalt/git-cc/pkg/parser"
)

const newScopeTemplate = "description of what short-form `%s` represents"

type Model struct {
	input             single_select.Model
	helpBar           helpbar.Model
	newScope          string
	copiedToClipboard bool
}

type editorStartMsg struct{}
type editorFinishedMsg struct{ err error }

// makeMatch returns a match function that captures the current options.
func makeMatch(options []string) func(string, string) bool {
	return func(query, option string) bool {
		if option == "new scope" {
			for _, opt := range options {
				if query == opt {
					return false
				}
			}
			return true
		}
		return single_select.MatchStart(query, option)
	}
}

// given options from config, add the leading "unscoped" and trailing "new scope" options
func makeOptions(options *config.OrderedMap[string, string]) (keys []string, values []string) {
	keys, values = config.ZippedOrderedKeyValuePairs(options)
	keys = append(append([]string{""}, keys...), "new scope")
	values = append(append([]string{"unscoped; affects the entire project"}, values...), "edit a new scope into your configuration file")
	return keys, values
}

func NewModel(cc *parser.CC, cfg config.Cfg) Model {
	options, hints := makeOptions(cfg.Scopes)
	newScope := ""
	copiedToClipboard := false
	return Model{
		single_select.NewModel(
			config.Faint("select a scope:"),
			cc.Scope,
			options, hints,
			makeMatch(options),
		),
		helpbar.NewModel(
			config.HelpSubmit,
			config.HelpSelect,
			config.HelpBack,
			config.HelpCancel,
		),
		newScope,
		copiedToClipboard,
	}
}

func (m Model) Value() string {
	return m.input.Value()
}

func (m Model) Render(s io.StringWriter) {
	if m.newScope != "" {
		_ = utils.Must(s.WriteString("new scope \""))
		_ = utils.Must(s.WriteString(m.newScope))
		_ = utils.Must(s.WriteString("\" "))
		if !m.copiedToClipboard {
			_ = utils.Must(s.WriteString("not "))
		}
		_ = utils.Must(s.WriteString("copied to clipboard\n"))
	}
	m.input.Render(s)
	_ = utils.Must(s.WriteString("\n"))
	m.helpBar.Render(s)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyEnter, tea.KeyTab:
			if m.Value() == "new scope" {
				m.newScope = m.input.CurrentInput()
				cmd = func() tea.Msg {
					return editorStartMsg{}
				}
				return m, cmd
			} else {
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}
	case editorStartMsg:
		{
			err := clipboard.WriteAll(m.newScope)
			m.copiedToClipboard = (err == nil)
		}
		editorCmd := config.EditCfgFileCmd(config.CentralStore)
		cmd = tea.ExecProcess(editorCmd, func(err error) tea.Msg {
			return editorFinishedMsg{err}
		})
		return m, cmd
	case tea.WindowSizeMsg:
		m.helpBar, _ = m.helpBar.Update(msg)
	case editorFinishedMsg:
		m.newScope = ""
		m.copiedToClipboard = false
		if msg.err != nil {
			// TODO: *gracefully* handle editor exiting with an error
			log.Fatal(msg.err)
		}
		if err := config.CentralStore.ReadCfgFile(true); err != nil {
			log.Fatalf(">>%+v", err) // FIXME: handle this error
			newScope := m.input.CurrentInput()
			suggested := config.CentralStore.Clone()
			if suggested.Scopes == nil {
				suggested.Scopes = &config.OrderedMap[string, string]{}
			}
			suggested.Scopes.Set(newScope, fmt.Sprintf(newScopeTemplate, newScope))
			editorCmd := config.EditCfgFileCmd(&suggested)
			cmd = tea.ExecProcess(editorCmd, func(err error) tea.Msg {
				return editorFinishedMsg{err}
			})
			return m, cmd
		} // TODO: warn about parse error
		values, hints := makeOptions(config.CentralStore.Scopes)
		cmd = m.input.UpdateItems(values, hints, makeMatch(values))
		return m, cmd
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) ShouldSkip(currentValue string) bool {
	for _, opt := range m.input.Options() {
		if currentValue == opt && opt != "" {
			return true
		}
	}
	return len(m.input.Options()) == 0 // should skip if no scope options are configured
}
