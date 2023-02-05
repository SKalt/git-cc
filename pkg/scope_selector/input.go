package scope_selector

import (
	"fmt"
	"log"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/skalt/git-cc/pkg/config"
	"github.com/skalt/git-cc/pkg/helpbar"
	"github.com/skalt/git-cc/pkg/parser"
	"github.com/skalt/git-cc/pkg/single_select"
)

const emptyScopeTemplate = "scopes:\n%s\n"
const newScopeTemplate = "  %s: description of what short-form \"%s\" represents"

type Model struct {
	input             single_select.Model
	helpBar           helpbar.Model
	newScope          string
	copiedToClipboard bool
}

type editorStartMsg struct{}
type editorFinishedMsg struct{ err error }

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
func makeOptions(options *config.OrderedMap) (keys []string, values []string) {
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
			match,
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

func (m Model) View() string {
	s := strings.Builder{}
	if m.newScope != "" {
		s.WriteString("new scope \"")
		s.WriteString(m.newScope)
		s.WriteString("\" ")
		if !m.copiedToClipboard {
			s.WriteString("not ")
		}
		s.WriteString("copied to clipboard\n")
		return s.String()
	}
	s.WriteString(m.input.View())
	s.WriteRune('\n')
	s.WriteString(m.helpBar.View())
	return s.String()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
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
		// editorStartMsg{newScope}
		editorCmd := config.EditCfgFileCmd(
			config.CentralStore,
			config.ExampleCfgFileHeader+config.ExampleCfgFileCommitTypes+"\n"+fmt.Sprintf(
				emptyScopeTemplate,
				fmt.Sprintf(newScopeTemplate, m.newScope, m.newScope),
			),
		)
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
		if err := config.CentralStore.ReadCfgFile(); err != nil {
			newScope := m.input.CurrentInput()
			editorCmd := config.EditCfgFileCmd(
				config.CentralStore,
				config.ExampleCfgFileHeader+config.ExampleCfgFileCommitTypes+"\n"+fmt.Sprintf(
					emptyScopeTemplate,
					fmt.Sprintf(newScopeTemplate, newScope, newScope),
				),
			)
			cmd = tea.ExecProcess(editorCmd, func(err error) tea.Msg {
				return editorFinishedMsg{err}
			})
			return m, cmd
		} // else {} // TODO: warn about parse error
		values, hints := makeOptions(config.CentralStore.Scopes)
		m.input.Options = values
		m.input.Hints = hints
		if m.input.Cursor >= len(m.input.Options) {
			m.input.Cursor = len(m.input.Options) - 1
		}
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) ShouldSkip(currentValue string) bool {
	for _, opt := range m.input.Options {
		if currentValue == opt && opt != "" {
			return true
		}
	}
	return len(m.input.Options) == 0 // should skip if no scope options are configured
}
