package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/skalt/git-cc/pkg/config"
	"github.com/skalt/git-cc/pkg/parser"
	"github.com/skalt/git-cc/pkg/tui_breaking_change_input"
	"github.com/skalt/git-cc/pkg/tui_description_editor"
	"github.com/skalt/git-cc/pkg/tui_single_select"
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
	commit  [doneIndex]string
	viewing componentIndex

	typeInput           tui_single_select.Model
	scopeInput          tui_single_select.Model
	descriptionInput    tui_description_editor.Model
	breakingChangeInput tui_breaking_change_input.Model

	choice chan string
}

func (m model) ready() bool {
	return len(m.commit[commitTypeIndex]) > 0 && len(m.commit[shortDescriptionIndex]) > 0
}

func (m model) contextValue() string {
	result := strings.Builder{}
	result.WriteString(m.commit[commitTypeIndex])
	scope := m.commit[scopeIndex]
	if scope != "" {
		result.WriteString(fmt.Sprintf("(%s)", scope))
	}
	return result.String()
}
func (m model) value() string {
	result := strings.Builder{}
	result.WriteString(m.contextValue())
	breakingChange := m.commit[breakingChangeIndex]
	if breakingChange != "" {
		result.WriteRune('!')
	}
	result.WriteString(fmt.Sprintf(": %s\n", m.commit[shortDescriptionIndex]))
	if breakingChange != "" {
		result.WriteString(fmt.Sprintf("\nBREAKING CHANGE: %s\n", breakingChange))
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
	typeModel := tui_single_select.NewModel(
		termenv.String("select a commit type: ").Faint().String(),
		cc.Type,
		cfg.CommitTypes)
	scopeModel := tui_single_select.NewModel(
		termenv.String("select a scope:").Faint().String(),
		cc.Scope,
		cfg.Scopes) // TODO: skip scopes none present?
	descModel := tui_description_editor.NewModel(
		cfg.HeaderMaxLength, cc.Description, cfg.EnforceMaxLength,
	)
	bcModel := tui_breaking_change_input.NewModel()
	breakingChanges := ""
	if cc.BreakingChange {
		for _, footer := range cc.Footers {
			result, err := parser.BreakingChange([]rune(footer))
			if err == nil {
				breakingChanges += string(result.Remaining) + "\n"
			}
		}
	}
	return model{
		choice: choice,
		commit: [doneIndex]string{
			cc.Type,
			cc.Scope,
			cc.Description,
			breakingChanges,
		},
		typeInput:           typeModel,
		scopeInput:          scopeModel,
		descriptionInput:    descModel,
		breakingChangeInput: bcModel,
		viewing:             commitTypeIndex}
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
func (m model) done() (model, tea.Cmd) {
	if m.ready() {
		m.choice <- m.value()
	} else {
		m.choice <- ""
	}
	return m, tea.Quit
}
func (m model) shouldSkip(component componentIndex) bool {
	switch component {
	case commitTypeIndex:
		commitType := m.typeInput.Value()
		for _, opt := range m.typeInput.Options {
			if commitType == opt {
				return true
			}
		}
		return false
	case scopeIndex:
		scope := m.scopeInput.Value()
		for _, opt := range m.scopeInput.Options {
			if scope == opt {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (m model) advance() model {
	for {
		m.viewing++
		if !m.shouldSkip(m.viewing) {
			break
		}
	}
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
				m.commit[m.viewing] = m.currentComponent().Value()
				m = m.advance()
			case scopeIndex:
				m.descriptionInput = m.descriptionInput.SetPrefix(
					m.contextValue() + ": ",
				)
				m = m.advance()
				return m, cmd
			case breakingChangeIndex:
				m.commit[breakingChangeIndex] = m.breakingChangeInput.Value()
				if m.ready() {
					return m.done()
				} else {
					err := fmt.Errorf("required")
					if m.commit[commitTypeIndex] == "" {
						m.viewing = commitTypeIndex
						m.typeInput = m.typeInput.SetErr(err)
					} else if m.commit[shortDescriptionIndex] == "" {
						m.viewing = shortDescriptionIndex
						m.descriptionInput = m.descriptionInput.SetErr(err)
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
	return m.currentComponent().View()
}

var rootCmd = &cobra.Command{
	Use:   "git-cc",
	Short: "yeah",
	// Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Lookup(config.Init())
		delegatedArgs := []string{}
		ccArgs := []string{}
		cc := &parser.CC{}
		for i, arg := range args {
			if len(arg) <= 0 || []rune(arg)[0] != '-' {
				ccArgs = args[i:]
				break // ?
			} else {
				delegatedArgs = append(delegatedArgs, arg)
			}
		}
		if len(ccArgs) > 0 {
			cc, _ = parser.ParseAsMuchOfCCAsPossible(strings.Join(ccArgs, " "))
		}
		valid := cc.MinimallyValid() &&
			cc.ValidCommitType(cfg.CommitTypes) &&
			cc.ValidScope(cfg.Scopes)
		if !valid {
			choice := make(chan string, 1)
			m := initialModel(choice, cc, cfg)
			ui := tea.NewProgram(m)
			if err := ui.Start(); err != nil {
				log.Fatal(err)
			}
			if result := <-choice; result == "" {
				close(choice)
				os.Exit(1) // no submission
			} else {
				f := config.GetCommitMessageFile()
				file, err := os.Create(f)
				if err != nil {
					log.Fatal(err)
				}
				_, err = file.Write([]byte(result))
				if err != nil {
					log.Fatal(err)
				}
				cmd := strings.Split(config.GetGitEditor(), " ")
				cmd = append(cmd, config.GetCommitMessageFile())
				process := exec.Command(cmd[0], cmd[1:]...)
				process.Stdin = os.Stdin
				process.Stdout = os.Stdout
				err = process.Run()
				if err != nil {
					log.Fatal(err)
				} else {
					os.Exit(0)
				}
			}
		}
	},
}

func init() {
	rootCmd.Flags().Bool("help", false, "display a man page if possible")
	rootCmd.Flags().Bool("h", false, "print usage to stderr")
	// TODO: use git commit's flag-args
	// --no-post-rewrite //?
	// --author=<author> // not sure if th
	// --date=<date>
	// --amend ... might be better manually?
	// --no-edit
	rootCmd.Flags().Bool("dry-run", false, "see the git-commit docs for --dry-run")
	rootCmd.Flags().BoolP("all", "a", false, "see the git-commit docs for --all|-a")
	rootCmd.Flags().BoolP("signoff", "s", false, "see the git-commit docs for --signoff|-s")
	rootCmd.Flags().Bool("no-gpg-sign", false, "see the git-commit docs for --no-gpg-sign")
	rootCmd.Flags().Bool("m", false, "ignored if args are passed")

}

func main() {
	rootCmd.Execute()
}
