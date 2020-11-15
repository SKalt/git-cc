package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/skalt/git-cc/pkg/config"
	"github.com/skalt/git-cc/pkg/parser"
)

var version string

func SetVersion(v string) {
	version = v
}

func versionMode() {
	fmt.Printf("git-cc %s\n", version)
}

// construct a shell `git commit` command with flags delegated from the git-cc
// cli
func getGitCommitCmd(cmd *cobra.Command) []string {
	commitCmd := []string{}
	noEdit, _ := cmd.Flags().GetBool("no-edit")
	message, _ := cmd.Flags().GetString("message")
	for _, name := range boolFlags {
		flag, _ := cmd.Flags().GetBool(name)
		if flag {
			commitCmd = append(commitCmd, "--"+name)
		}
	}
	if !noEdit || len(message) == 0 {
		commitCmd = append(commitCmd, "--edit")
	}
	return commitCmd
}

// run a potentially interactive `git commit`
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

// run the conventional-commit helper logic. This may/not break into the TUI.
func mainMode(cmd *cobra.Command, args []string) {
	cfg := config.Lookup(config.Init())
	commitParams := getGitCommitCmd(cmd)
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	committingAllChanges, _ := cmd.Flags().GetBool("all")
	if !dryRun && !committingAllChanges {
		buf := &bytes.Buffer{}
		process := exec.Command("git", "diff", "--name-only", "--cached")
		process.Stdout = buf
		err := process.Run()
		if err != nil {
			log.Fatal(err)
		}
		if buf.String() == "" {
			log.Fatal("No files staged")
		}
	}

	cc := &parser.CC{}
	message, _ := cmd.Flags().GetString("message")
	if len(message) > 0 {
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
}

var Cmd = &cobra.Command{
	Use:   "git-cc",
	Short: "write conventional commits",
	// not using cobra subcommands since they prevent passing arbitrary arguments
	Run: func(cmd *cobra.Command, args []string) {
		version, _ := cmd.Flags().GetBool("version")
		if version {
			versionMode()
			os.Exit(0)
		}
		genCompletion, _ := cmd.Flags().GetBool("generate-shell-completion")
		if genCompletion {
			generateShellCompletion(cmd, args)
			os.Exit(0)
		}
		genManPage, _ := cmd.Flags().GetBool("generate-man-page")
		if genManPage {
			generateManPage(cmd, args)
			os.Exit(0)
		}
		mainMode(cmd, args)
	},
}

func init() {
	Cmd.Flags().BoolP("help", "h", false, "print the usage of git-cc")
	Cmd.Flags().Bool("dry-run", false, "Only print the resulting conventional commit message; don't commit.")
	Cmd.Flags().StringP("message", "m", "", "pass a complete conventional commit. If valid, it'll be committed without editing.")
	Cmd.Flags().Bool("version", false, "print the version")
	// TODO: accept more of git commit's flags; see https://git-scm.com/docs/git-commit
	// likely: --cleanup=<mode>
	// more difficult, and possibly better done manually: --amend, -C <commit>
	// --reuse-message=<commit>, -c <commit>, --reedit-message=<commit>,
	// --fixup=<commit>, --squash=<commit>
	Cmd.Flags().String("author", "", "delegated to git-commit")
	Cmd.Flags().String("date", "", "delegated to git-commit")
	Cmd.Flags().BoolP("all", "a", false, "see the git-commit docs for --all|-a")
	Cmd.Flags().BoolP("signoff", "s", false, "see the git-commit docs for --signoff|-s")
	Cmd.Flags().Bool("no-gpg-sign", false, "see the git-commit docs for --no-gpg-sign")
	Cmd.Flags().Bool("no-post-rewrite", false, "Bypass the post-rewrite hook")
	Cmd.Flags().Bool("no-edit", false, "Use the selected commit message without launching an editor.")

	Cmd.Flags().Bool("generate-man-page", false, "Generate a man page in your manpath")
	Cmd.Flags().Bool(
		"generate-shell-completion",
		false,
		"print a bash/zsh/fish/powershell completion script to stdout",
	)
}
