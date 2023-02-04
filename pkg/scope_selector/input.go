package scope_selector

import (
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/skalt/git-cc/pkg/config"
	"github.com/skalt/git-cc/pkg/helpbar"
	"github.com/skalt/git-cc/pkg/parser"
	"github.com/skalt/git-cc/pkg/single_select"
)

const emptyScopeTemplate = "scopes:\n%s\n"
const newScopeTemplate = "  %s: description of what short-form \"%s\" represents"

type Model struct {
	input   single_select.Model
	helpBar helpbar.Model
}

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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyTab:
			if m.Value() == "new scope" {
				newScope := m.input.CurrentInput()
				// editorStartMsg{newScope}
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
			} else {
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}
	case tea.WindowSizeMsg:
		m.helpBar, _ = m.helpBar.Update(msg)
	case editorFinishedMsg:
		if msg.err != nil {
			// TODO: handle editor exiting with an error
			log.Fatal(msg.err)
		}
		if err := config.CentralStore.ReadCfgFile(); err != nil {
			// TODO: warn about parse error
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
		}
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
