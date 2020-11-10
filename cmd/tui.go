package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/skalt/git-cc/pkg/breaking_change_input"
	"github.com/skalt/git-cc/pkg/config"
	"github.com/skalt/git-cc/pkg/description_editor"
	"github.com/skalt/git-cc/pkg/parser"
	"github.com/skalt/git-cc/pkg/single_select"
)

type componentIndex int

const ( // the order of the components
	commitTypeIndex componentIndex = iota
	scopeIndex
	shortDescriptionIndex
	breakingChangeIndex
	// body omitted -- performed by GIT_EDITOR
	doneIndex
)

var (
	boolFlags = [...]string{"all", "signoff", "no-post-rewrite", "no-gpg-sign"}
)

type InputComponent interface {
	View() string
	Value() string

	// Update(tea.Msg) (tea.Model, tea.Cmd)
	// // tea.Model       // Init() tea.Cmd, Update(tea.Msg) (tea.Model, tea.Cmd), View() string
	// Focus() tea.Cmd // should focus any internals, i.e. text inputs
	// // Cancel()  // should clean up any resources (i.e. open channels)
	// Submit()  // send the input to the output channel
}

type model struct {
	// components [done]InputComponent
	commit  [doneIndex]string
	viewing componentIndex

	typeInput           single_select.Model
	scopeInput          single_select.Model
	descriptionInput    description_editor.Model
	breakingChangeInput breaking_change_input.Model

	choice chan string
}

func (m model) ready() bool {
	return len(m.commit[commitTypeIndex]) > 0 && len(m.commit[shortDescriptionIndex]) > 0
}

func (m model) contextValue() string {
	result := strings.Builder{}
	result.WriteString(m.commit[commitTypeIndex])
	scope := m.commit[scopeIndex]
	breakingChange := m.commit[breakingChangeIndex]
	if scope != "" {
		result.WriteString(fmt.Sprintf("(%s)", scope))
	}
	if breakingChange != "" {
		result.WriteRune('!')
	}
	result.WriteString(": ")
	return result.String()
}
func (m model) value() string {
	result := strings.Builder{}
	result.WriteString(m.contextValue())
	result.WriteString(m.commit[shortDescriptionIndex])
	result.WriteString("\n")
	breakingChange := m.commit[breakingChangeIndex]
	if breakingChange != "" {
		result.WriteString(fmt.Sprintf("\n\nBREAKING CHANGE: %s\n", breakingChange))
		// TODO: handle muliple breaking change footers(?)
	}
	return result.String()
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) currentComponent() InputComponent {
	return [...]InputComponent{
		m.typeInput,
		m.scopeInput,
		m.descriptionInput,
		m.breakingChangeInput,
	}[m.viewing]
}

// Pass a channel to the model to listen to the result value. This is a
// function that returns the initialize function and is typically how you would
// pass arguments to a tea.Init function.
func initialModel(choice chan string, cc *parser.CC, cfg config.Cfg) model {
	typeModel := single_select.NewModel(
		termenv.String("select a commit type: ").Faint().String(), // context
		cc.Type, // value
		cfg.CommitTypes,
	)
	scopeModel := single_select.NewModel(
		termenv.String("select a scope:").Faint().String(),
		cc.Scope,
		append(
			[]map[string]string{{"": "unscoped; affects the entire project"}},
			cfg.Scopes...,
		),
	) // TODO: Option to add new scope?
	descModel := description_editor.NewModel(
		cfg.HeaderMaxLength, cc.Description, cfg.EnforceMaxLength,
	)
	bcModel := breaking_change_input.NewModel()
	breakingChanges := ""
	if cc.BreakingChange {
		for _, footer := range cc.Footers {
			result, err := parser.BreakingChange([]rune(footer))
			if err == nil {
				breakingChanges += string(result.Remaining) + "\n"
			}
		}
	}
	commit := [doneIndex]string{
		cc.Type,
		cc.Scope,
		cc.Description,
		breakingChanges,
	}
	m := model{
		choice:              choice,
		commit:              commit,
		typeInput:           typeModel,
		scopeInput:          scopeModel,
		descriptionInput:    descModel,
		breakingChangeInput: bcModel,
		viewing:             commitTypeIndex}
	if m.shouldSkip(m.viewing) {
		m = m.submit().advance()
		m.descriptionInput = m.descriptionInput.SetPrefix(m.contextValue())
	}

	return m
}

func (m model) updateCurrentInput(msg tea.Msg) model {
	switch m.viewing {
	case commitTypeIndex:
		m.typeInput, _ = m.typeInput.Update(msg)
	case scopeIndex:
		m.scopeInput, _ = m.scopeInput.Update(msg)
	case shortDescriptionIndex:
		m.descriptionInput, _ = m.descriptionInput.Update(msg)
	case breakingChangeIndex:
		m.breakingChangeInput, _ = m.breakingChangeInput.Update(msg)
	}
	return m
}

func (m model) shouldSkip(component componentIndex) bool {
	switch component {
	case commitTypeIndex:
		commitType := m.commit[commitTypeIndex]
		for _, opt := range m.typeInput.Options {
			if commitType == opt {
				return true
			}
		}
		return false
	case scopeIndex:
		if len(m.scopeInput.Options) == 0 {
			return true
		}
		scope := m.commit[scopeIndex]
		for _, opt := range m.scopeInput.Options {
			if scope == opt && opt != "" {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (m model) advance() model { // TODO: consider submitting w/in this fn
	for {
		m.viewing++
		if !m.shouldSkip(m.viewing) {
			break
		}
	}
	return m
}

func (m model) submit() model {
	m.commit[m.viewing] = m.currentComponent().Value()
	m.descriptionInput = m.descriptionInput.SetPrefix(m.contextValue())
	return m
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			m.choice <- ""
			return m, tea.Quit
		case tea.KeyShiftTab:
			if m.viewing > commitTypeIndex {
				m.viewing--
			}
			return m, cmd
		case tea.KeyEnter:
			switch m.viewing {
			default:
				m = m.submit().advance()
			case breakingChangeIndex:
				m = m.submit()
				if m.ready() {
					m.choice <- m.value()
					return m, tea.Quit
				} else {
					// TODO: better validation messages
					if m.commit[commitTypeIndex] == "" {
						m.viewing = commitTypeIndex
					} else if m.commit[shortDescriptionIndex] == "" {
						m.viewing = shortDescriptionIndex
					}
					return m, cmd
				}
			case doneIndex:
				fmt.Printf("%d > done", m.viewing)
				os.Exit(1)
			}
			return m, cmd
		default:
			m = m.updateCurrentInput(msg)
		}
	}
	return m, cmd
}

func (m model) View() string {
	return m.currentComponent().View() + "\n"
}
