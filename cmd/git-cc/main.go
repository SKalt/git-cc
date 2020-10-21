package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	input "github.com/skalt/git-cc/pkg/tui_single_select"
)

// TODO: move to cfg
var commitTypes = [...]string{ // ordered in terms of frequency ?
	"feat",
	"fix",
	"docs",
	"chore",
	"style",
	"refactor",
	"perf",
	"test",
	"build",
	"ci",
	"revert",
}
var commitTypePrompts = map[string]string{
	// see https://github.com/angular/angular.js/blob/master/DEVELOPERS.md#type
	// see https://github.com/conventional-changelog/commitlint/blob/master/%40commitlint/config-conventional/index.js#L23
	"feat":     "adds a new feature",
	"fix":      "fixes a bug",
	"docs":     "changes only the documentation",
	"style":    "changes the style but not the meaning of the code (such as formatting)",
	"perf":     "improves performance",
	"test":     "adds or corrects tests",
	"build":    "changes the build system or external dependencies",
	"chore":    "changes outside the code, docs, or tests",
	"ci":       "changes to the Continuous Inegration (CI) system",
	"refactor": "changes the code without changing behavior",
	"revert":   "reverts prior changes",
}

type model struct {
	choice    chan<- string
	textInput input.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func main() {
	choice := make(chan string, 1)
	ui := tea.NewProgram(initialModel(choice))
	if err := ui.Start(); err != nil {
		log.Fatal(err)
	}

	if r := <-choice; r != "" {
		fmt.Printf("\n---\nYou chose %s!\n", r)
	}
}

// Pass a channel to the model to listen to the result value. This is a
// function that returns the initialize function and is typically how you would
// pass arguments to a tea.Init function.
func initialModel(choice chan string) model {
	hints := []string{}
	options := []string{}
	for _, commitType := range commitTypes {
		options = append(options, commitType)
		hint := commitTypePrompts[commitType]
		hints = append(hints, hint)
	}
	textModel := input.NewModel("   ", options, hints, choice)
	return model{
		textInput: textModel,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.textInput, cmd = input.Update(msg, m.textInput)
	return m, cmd
}

func (m model) View() string {
	return m.textInput.View()
}
