package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/skalt/git-cc/pkg/config"
	"github.com/skalt/git-cc/pkg/tui_description_editor"
	"github.com/skalt/git-cc/pkg/tui_single_select"
)

type componentIndex int

const ( // the order of the components
	commitType componentIndex = iota
	scope
	shortDescription
	breakingChange
	body
	done
)

type InputComponent interface {
	View() string
	// Update(tea.Msg) (tea.Model, tea.Cmd)

	Value() string
	// // tea.Model       // Init() tea.Cmd, Update(tea.Msg) (tea.Model, tea.Cmd), View() string
	// Focus() tea.Cmd // should focus any internals, i.e. text inputs
	// // Cancel()  // should clean up any resources (i.e. open channels)
	// Submit()  // send the input to the output channel
}

type model struct {
	// components [done]InputComponent
	commit  [done]string
	viewing componentIndex

	typeInput           tui_single_select.Model
	scopeInput          tui_single_select.Model
	descriptionInput    tui_description_editor.Model
	breakingChangeInput tui_description_editor.Model
	// description body/footers handled by external editors...for now
	// commitType
	//
	// commitTypeChoice chan<- string
	// commitTypeSelector input.Model
	// editor editor.Model
}

func (m model) Init() tea.Cmd {
	switch m.viewing {
	case commitType:
		return m.typeInput.Focus()
	case scope:
		return m.scopeInput.Focus()
	case shortDescription:
		return m.descriptionInput.Focus()
	case breakingChange:
		return m.breakingChangeInput.Focus()
	default:
		panic("???")
	}
}

// TODO: validate commit-msg

func (m model) currentComponent() InputComponent {
	return [...]InputComponent{
		m.typeInput,
		m.scopeInput,
		m.descriptionInput,
		m.breakingChangeInput,
	}[m.viewing]
}

func (m model) goBack() model {
	if m.viewing > commitType {
		m.viewing--
	}
	m.focusCurrentComponent()
	return m
}

func (m model) goForward() model {
	if m.viewing < body {
		m.viewing++
	}
	m.focusCurrentComponent()
	return m
}

func main() {
	choice := make(chan string, 1)
	m := initialModel(choice)
	ui := tea.NewProgram(m)
	if err := ui.Start(); err != nil {
		log.Fatal(err)
	}
	if r := <-choice; r != "" {
		fmt.Printf("\n---\nYou chose %s!\n", r)
	} else {
		os.Exit(1)
	}
}

// Pass a channel to the model to listen to the result value. This is a
// function that returns the initialize function and is typically how you would
// pass arguments to a tea.Init function.
func initialModel(choice chan string) model {
	cfg := config.Init()
	data := config.Lookup(cfg)
	commitTypeOptions, commitTypeHints := []string{}, []string{}
	for _, commitType := range data.CommitTypes {
		commitTypeOptions = append(commitTypeOptions, commitType.Name)
		commitTypeHints = append(commitTypeHints, commitType.Description)
	}
	scopeOptions, scopeHints := []string{"cli", "parser"}, []string{"the cli UI", "the conventional-commit parser"}
	// TODO: read in scopes from cfg.
	for _, scope := range data.Scopes {
		scopeOptions = append(scopeOptions, scope.Name)
		scopeHints = append(scopeHints, scope.Description)
	}
	typeModel := tui_single_select.NewModel(commitTypeOptions, commitTypeHints)
	scopeModel := tui_single_select.NewModel(scopeOptions, scopeHints)
	descModel := tui_description_editor.NewModel("\n", 50)           // TODO: get soft-length limit
	fake := tui_description_editor.NewModel("breaking change", 1000) // ???
	return model{
		commit:              [done]string{}, // TODO: read initial state from cli
		typeInput:           typeModel,
		scopeInput:          scopeModel,
		descriptionInput:    descModel,
		breakingChangeInput: fake,
		viewing:             commitType,
	}
}

func (m model) updateCurrentInput(msg tea.Msg) model {
	switch m.viewing {
	case commitType:
		m.typeInput, _ = m.typeInput.Update(msg)
	case scope:
		m.scopeInput, _ = m.scopeInput.Update(msg)
	case shortDescription:
		m.descriptionInput, _ = m.descriptionInput.Update(msg)
	case breakingChange:
		m.breakingChangeInput, _ = m.descriptionInput.Update(msg)
	}
	return m
}

// deprecate?
func (m model) focusCurrentComponent() model {
	switch m.viewing {
	case commitType:
		m.typeInput.Focus()
	case scope:
		m.scopeInput.Focus()
	case shortDescription:
		m.descriptionInput.Focus()
	case breakingChange:
		m.descriptionInput.Focus()
	}
	return m
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlD:
			os.Exit(1)
		case tea.KeyShiftTab:
			m = m.goBack()
			return m, cmd
		case tea.KeyEnter:
			switch m.viewing {
			case scope:
				m.descriptionInput.SetPrefix(fmt.Sprintf("%s(%s)", m.commit[commitType], m.commit[scope]))
				m = m.goForward()
			case body:
				fmt.Println("body not yet implemented")
				os.Exit(1)
			case done:
				fmt.Printf("%d > done", m.viewing)
				os.Exit(1)
			default:
				m.commit[m.viewing] = m.currentComponent().Value()
				m = m.goForward()
			}
			return m, cmd
		default:
			// m.components[m.viewing], cmd =
			m = m.updateCurrentInput(msg)
		}
	}
	return m, cmd
}

func (m model) View() string {
	return m.currentComponent().View()
}
