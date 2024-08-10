package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/skalt/git-cc/internal/breaking_change_input"
	"github.com/skalt/git-cc/internal/config"
	"github.com/skalt/git-cc/internal/description_editor"
	"github.com/skalt/git-cc/internal/scope_selector"
	"github.com/skalt/git-cc/internal/type_selector"
	"github.com/skalt/git-cc/pkg/parser"
)

type componentIndex int

const ( // the order of the components
	commitTypeIndex componentIndex = iota
	scopeIndex
	shortDescriptionIndex
	breakingChangeIndex
	// body omitted -- performed by GIT_EDITOR
	nIndices // the number of indices
)

var (
	boolFlags = [...]string{
		"all",
		"signoff",
		"no-signoff",
		"no-post-rewrite",
		"no-gpg-sign",
		"no-verify", // https://git-scm.com/docs/git-commit#Documentation/git-commit.txt---no-verify
	}
)

type InputComponent interface {
	View() string
	Value() string
}

type model struct {
	commit  [nIndices]string
	viewing componentIndex

	typeInput           type_selector.Model
	scopeInput          scope_selector.Model
	descriptionInput    description_editor.Model
	breakingChangeInput breaking_change_input.Model
	// the width of the terminal; needed for instantiating components
	// width  int
	choice chan string
	// any body stashed during the initial parse of command-line --message args
	remainingBody string
}

// returns whether the minimum requirements for a conventional commit are met.
func (m model) ready() bool {
	return len(m.commit[commitTypeIndex]) > 0 && len(m.commit[shortDescriptionIndex]) > 0
}

// returns the context portion of the CC header, e.g `type(scope): `.
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
func (m model) descriptionValue() string {
	return m.commit[shortDescriptionIndex]
}
func (m model) breakingChangeValue() string {
	return m.commit[breakingChangeIndex]
}

// Returns a pretty-printed CC string. The model should be `.ready()` before you call `.value()`.
func (m model) value() string {
	result := strings.Builder{}
	result.WriteString(m.contextValue())
	result.WriteString(m.descriptionValue())
	result.WriteString("\n")
	if m.remainingBody != "" {
		result.WriteString(m.remainingBody)
		result.WriteString("\n")
	}
	if breakingChange := m.breakingChangeValue(); breakingChange != "" {
		// TODO: handle multiple breaking change footers(?)
		result.WriteString(fmt.Sprintf("\n\nBREAKING CHANGE: %s\n", breakingChange))
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
func initialModel(choice chan string, cc *parser.CC, cfg *config.Cfg) model {
	typeModel := type_selector.NewModel(cc, cfg)
	scopeModel := scope_selector.NewModel(cc, *cfg)
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
	commit := [nIndices]string{
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
		viewing:             commitTypeIndex,
		remainingBody:       cc.Body,
	}
	if m.shouldSkip(m.viewing) {
		m = m.submit().advance()
		m.descriptionInput = m.descriptionInput.SetPrefix(m.contextValue())
	}
	return m
}

// pass the `msg` to the currently-displayed component/view
func (m model) updateCurrentInput(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.viewing {
	case commitTypeIndex:
		m.typeInput, cmd = m.typeInput.Update(msg)
	case scopeIndex:
		m.scopeInput, cmd = m.scopeInput.Update(msg)
	case shortDescriptionIndex:
		m.descriptionInput, cmd = m.descriptionInput.Update(msg)
	case breakingChangeIndex:
		m.breakingChangeInput, cmd = m.breakingChangeInput.Update(msg)
	}
	return m, cmd
}

func (m model) shouldSkip(component componentIndex) bool {
	switch component {
	case commitTypeIndex:
		return m.typeInput.ShouldSkip(m.commit[commitTypeIndex])
	case scopeIndex:
		return m.scopeInput.ShouldSkip(m.commit[scopeIndex])
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
		case tea.KeyEnter, tea.KeyTab:
			switch m.viewing {
			default:
				m = m.submit().advance()
			case commitTypeIndex:
				if m.currentComponent().Value() == "" {
					return m, cmd
				} else {
					m = m.submit().advance()
				}
			case scopeIndex:
				if m.currentComponent().Value() == "new scope" {
					m.scopeInput, cmd = m.scopeInput.Update(msg)
					return m, cmd
				} else {
					m = m.submit().advance()
				}
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
			}
			return m, cmd
		default:
			m, cmd = m.updateCurrentInput(msg)
		}
	case tea.WindowSizeMsg:
		// ensure instances of tea.WindowSizeMsg reach all child-components
		m.typeInput, _ = m.typeInput.Update(msg)
		m.scopeInput, _ = m.scopeInput.Update(msg)
		m.descriptionInput, _ = m.descriptionInput.Update(msg)
		m.breakingChangeInput, cmd = m.breakingChangeInput.Update(msg)
	default:
		m, cmd = m.updateCurrentInput(msg)
	}
	return m, cmd
}

func (m model) View() string {
	return m.currentComponent().View() + "\n"
}
