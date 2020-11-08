package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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
func getGitCommitCmd(cmd *cobra.Command) []string {
	commitCmd := []string{}
	// TODO: check message not passed
	noEdit, _ := cmd.Flags().GetBool("no-edit")
	message, _ := cmd.Flags().GetString("message")
	for _, name := range boolFlags {
		flag, _ := cmd.Flags().GetBool(name)
		if flag {
			commitCmd = append(commitCmd, "--"+name)
		}
	}
	if !noEdit || len(message) > 0 {
		commitCmd = append(commitCmd, "--edit")
	}
	return commitCmd
}

func doCommit(message string, dryRun bool, commitParams []string) {
	f := config.GetCommitMessageFile()
	file, err := os.Create(f)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write([]byte(message))
	if err != nil {
		log.Fatal(err)
	}
	if dryRun {
		fmt.Println(message)
		return
	}
	process := exec.Command("git", append([]string{
		"commit", "--message", message},
		commitParams...)...)
	process.Stdin = os.Stdin
	process.Stdout = os.Stdout
	if !dryRun {
		err = process.Run()
		if err != nil {
			log.Fatal(err)
		} else {
			os.Exit(0)
		}
	}
}

var Cmd = &cobra.Command{
	Use: "git-cc",
	// Long: "",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Lookup(config.Init())
		commitParams := getGitCommitCmd(cmd)
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		cc := &parser.CC{}
		message, _ := cmd.Flags().GetString("message")
		// mPassed := false
		if len(message) > 0 {
			// mPassed = true
			message += " "
		}
		message = message + strings.Join(args, " ")
		cc, _ = parser.ParseAsMuchOfCCAsPossible(message)
		valid := cc.MinimallyValid() &&
			cc.ValidCommitType(cfg.CommitTypes) &&
			(cc.ValidScope(cfg.Scopes) || cc.Scope == "")
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
				doCommit(result, dryRun, commitParams)
			}
		} else {
			doCommit(cc.ToString(), dryRun, commitParams)
		}
	},
}

func init() {
	Cmd.Flags().BoolP("help", "h", false, "print the usage of git-cc")
	Cmd.Flags().Bool("dry-run", false, "Only print the resulting conventional commit message; don't commit.")
	Cmd.Flags().StringP("message", "m", "", "pass a complete conventional commit. If valid, it'll be committed without editing.")
	// TODO: use git commit's flags; see https://git-scm.com/docs/git-commit
	// --author=<author>
	// --date=<date>

	// --amend ... might be better manually?
	// --no-edit
	// -C <commit>
	// --reuse-message=<commit>
	// Take an existing commit object, and reuse the log message and the authorship information (including the timestamp) when creating the commit.
	// -c <commit>
	// --reedit-message=<commit>
	// Like -C, but with -c the editor is invoked, so that the user can further edit the commit message.
	// --fixup=<commit>
	// Construct a commit message for use with rebase --autosquash. The commit message will be the subject line from the specified commit with a prefix of "fixup! ". See git-rebase[1] for details.
	// --squash=<commit>
	// Construct a commit message for use with rebase --autosquash. The commit message subject line is taken from the specified commit with a prefix of "squash! ". Can be used with additional commit message options (-m/-c/-C/-F). See git-rebase[1] for details.
	// -short
	// When doing a dry-run, give the output in the short-format. See git-status[1] for details. Implies --dry-run.
	// --cleanup=<mode>
	// This option determines how the supplied commit message should be cleaned up before committing. The <mode> can be strip, whitespace, verbatim, scissors or default.
	Cmd.Flags().BoolP("all", "a", false, "see the git-commit docs for --all|-a")
	Cmd.Flags().BoolP("signoff", "s", false, "see the git-commit docs for --signoff|-s")
	Cmd.Flags().Bool("no-gpg-sign", false, "see the git-commit docs for --no-gpg-sign")
	Cmd.Flags().Bool("no-post-rewrite", false, "Bypass the post-rewrite hook")
	Cmd.Flags().Bool("no-edit", false, "Use the selected commit message without launching an editor.")
}

func getBoolFlag(flags *pflag.FlagSet, name string) []string {
	present, _ := flags.GetBool(name)
	if present {
		return []string{"--" + name}
	} else {
		return []string{}
	}
}
