package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/skalt/git-cc/pkg/config"
	input "github.com/skalt/git-cc/pkg/tui_single_select"
)

type model struct {
	commitTypeChoice   chan<- string
	commitTypeSelector input.Model
	//
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
	cfg := config.Init()
	data := config.Lookup(cfg)
	commitTypes := data.CommitTypes
	options := make([]string, len(commitTypes))
	hints := make([]string, len(commitTypes))
	for i, description := range commitTypes {
		options[i] = description.Name
		hints[i] = description.Description
	}
	textModel := input.NewModel("   ", options, hints, choice)
	return model{
		commitTypeSelector: textModel,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.commitTypeSelector, cmd = input.Update(msg, m.commitTypeSelector)
	return m, cmd
}

func (m model) View() string {
	return m.commitTypeSelector.View()
}
