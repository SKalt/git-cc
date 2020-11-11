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
	if !noEdit || len(message) == 0 {
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

func mainMode(cmd *cobra.Command, args []string) {
	cfg := config.Lookup(config.Init())
	commitParams := getGitCommitCmd(cmd)
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if !dryRun {
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
}

var Cmd = &cobra.Command{
	Use:   "git-cc",
	Short: "write conventional commits",
	Run: func(cmd *cobra.Command, args []string) {
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

	// TODO: use git commit's flags; see https://git-scm.com/docs/git-commit
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
