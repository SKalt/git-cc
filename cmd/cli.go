package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"

	"github.com/skalt/git-cc/internal/config"
	"github.com/skalt/git-cc/internal/utils"
	"github.com/skalt/git-cc/pkg/parser"
)

var version, commit, date string

func printVersion(version, commit, date string) {
	s := strings.Builder{}
	s.WriteString("git-cc ")
	s.WriteString(version)
	s.WriteString(" commit ")
	s.WriteString(commit)
	s.WriteString(" built ")
	s.WriteString(date)
	s.WriteRune('\n')
	fmt.Print(s.String())
}

// construct a shell `git commit` command with flags delegated from the git-cc
// cli
func getGitCommitCmd(cmd *cobra.Command) []string {
	commitCmd := []string{}
	noEdit, _ := cmd.Flags().GetBool("no-edit")
	message, _ := cmd.Flags().GetStringArray("message")
	flags := cmd.Flags()
	for _, name := range boolFlags {
		if flags.Lookup(name).Changed {
			flag, err := flags.GetBool(name)
			if err == nil && flag {
				commitCmd = append(commitCmd, "--"+name)
			}
		}
	}
	if noEdit || len(message) > 0 {
		commitCmd = append(commitCmd, "--no-edit")
	} else {
		commitCmd = append(commitCmd, "--edit")
	}
	return commitCmd
}

// run a potentially interactive `git commit`
func doCommit(message string, dryRun bool, commitParams []string) {
	f := config.GetCommitMessageFile()
	file, err := os.Create(f)
	if err != nil {
		log.Fatalf("unable to create %s: %+v", f, err)
	}
	_, err = file.Write([]byte(message))
	if err != nil {
		log.Fatalf("unable to write to %s: %+v", f, err)
	}
	if dryRun {
		fmt.Println(message)
	}
	cmd := append([]string{"git", "commit", "--message", message}, commitParams...)
	process := exec.Command(cmd[0], cmd[1:]...)
	process.Stdin = os.Stdin
	process.Stdout = os.Stdout
	process.Stderr = os.Stderr
	if dryRun {
		fmt.Printf("would run: `%s`\n", strings.Join(cmd, " "))
		os.Exit(0)
	} else {
		err = process.Run()
		if err != nil {
			log.Fatalf("failed running `%+v`: %+v", cmd, err)
		} else {
			os.Exit(0)
		}
	}
}

// 0000 0001 : invalid type
// 0000 0010 : missing type
// 0000 0100 : invalid scope
// 0000 1000 : missing description
type ValidationErrors = uint8

const (
	InvalidType        uint8 = 1 << 0
	MissingType        uint8 = 1 << 1
	InvalidScope       uint8 = 1 << 2
	MissingDescription uint8 = 1 << 3
)

// run the conventional-commit helper logic. This may/not break into the TUI.
func mainMode(cmd *cobra.Command, args []string, cfg *config.Cfg) {

	commitParams := getGitCommitCmd(cmd)
	committingAllChanges, _ := cmd.Flags().GetBool("all")
	allowEmpty, _ := cmd.Flags().GetBool("allow-empty")
	if !cfg.DryRun && !committingAllChanges {
		buf := &bytes.Buffer{}
		process := exec.Command("git", "diff", "--name-only", "--cached")
		process.Stdout = buf
		err := process.Run()
		if err != nil {
			log.Fatalf("fatal: not a git repository (or any of the parent directories): .git; %+v", err)
		}
		if buf.String() == "" && !allowEmpty {
			log.Fatal("No files staged")
		}
	}

	var cc *parser.CC

	message, _ := cmd.Flags().GetStringArray("message")

	if len(message) > 0 {
		//> If multiple `-m` options are given, their values are concatenated as separate paragraphs.
		//> see https://git-scm.com/docs/git-commit#Documentation/git-commit.txt---messageltmsggt
		cc, _ = parser.ParseAsMuchOfCCAsPossible(strings.Join(message, "\n\n"))
	} else {
		cc, _ = parser.ParseAsMuchOfCCAsPossible((strings.Join(args, " ")))
	}
	var validationErrors ValidationErrors = 0
	if cc.Type == "" {
		validationErrors |= InvalidType
	} else {
		if _, valid := cfg.CommitTypes.Get(cc.Type); !valid {
			validationErrors |= InvalidType
		}
	}
	if cc.Scope != "" {
		if _, valid := cfg.Scopes.Get(cc.Scope); !valid {
			validationErrors |= InvalidScope
		}
	}
	if cc.Description == "" {
		validationErrors |= MissingDescription
	}

	if validationErrors != 0 {
		m := initialModel(cc, cfg)
		ui := tea.NewProgram(m)
		out, err := ui.Run()
		if err != nil {
			log.Fatal(err)
		}
		result := out.(model)
		if !result.ready() {
			os.Exit(1) // no submission
		} else {
			commitMessage := result.value()
			f := config.GetCommitMessageFile()
			file, err := os.Create(f)
			if err != nil {
				log.Fatalf("unable to create fil %s: %+v", f, err)
			}
			_, err = file.Write([]byte(commitMessage))
			if err != nil {
				log.Fatalf("unable to write to file %s: %+v", f, err)
			}
			doCommit(commitMessage, cfg.DryRun, commitParams)
		}
	} else {
		doCommit(cc.ToString(), cfg.DryRun, commitParams)
	}
}

func redoMessage(cmd *cobra.Command) {
	flags := cmd.Flags()
	msg := utils.Must(flags.GetStringArray("message"))
	if len(msg) > 0 {
		log.Fatal("-m|--message is incompatible with --redo")
	}
	commitMessagePath := config.GetCommitMessageFile()
	data, err := os.ReadFile(commitMessagePath)
	// TODO: ignore commented lines in git-commit file
	if err != nil {
		fmt.Printf("file not found: %s", commitMessagePath)
		os.Exit(127)
	}
	empty := true
	for line := range strings.SplitSeq(
		strings.TrimSpace(string(data)),
		"\n",
	) {
		if !strings.HasPrefix(strings.TrimLeft(line, " \t\r\n"), "#") {
			empty = false
			break // ok
		}
	}
	if empty {
		log.Fatalf("Empty commit message: %s", commitMessagePath)
	}

	utils.Check(flags.Set("message", string(data)))
}

// Note: I'm not using cobra subcommands since they prevent passing arbitrary arguments,
// and I'd like to be able to start an invocation like `git-cc this is the commit message`
// without having to think about whether `this` is a subcommand.

func run(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	if shouldPrintVersion, _ := flags.GetBool("version"); shouldPrintVersion {
		printVersion(version, commit, date)
		os.Exit(0)
	} else if genCompletion, _ := flags.GetBool("generate-shell-completion"); genCompletion {
		generateShellCompletion(cmd, args)
		os.Exit(0)
	} else if genManPage, _ := flags.GetBool("generate-man-page"); genManPage {
		generateManPage(cmd, args)
		os.Exit(0)
	} else {
		dryRun := utils.Must(cmd.Flags().GetBool("dry-run"))
		cfg, err := config.Init(dryRun)
		if err != nil {
			log.Fatalf("%s", err)
		}
		if showConfig, _ := flags.GetBool("show-config"); showConfig {
			repoRoot, _ := config.GetGitRepoRoot()
			_, tried, _ := config.FindCCConfigFile(repoRoot)
			for _, f := range tried {
				fmt.Printf("# %s\n", f)
			}
			file := cfg.ConfigFile
			if file == "" {
				file = "<default>"
			}
			fmt.Printf("config file path: %s\n", file)
			os.Exit(0)
		}
		if init := utils.Must(flags.GetBool("init")); init {
			format, _ := cmd.Flags().GetString("config-format")
			switch format {
			case "yaml", "yml", "toml":
				break
			default:
				log.Fatalf("unsupported default config-file format: %s", format)
			}
			if err != nil {
				log.Fatalf("%s", err)
			}
			if err := config.InitDefaultCfgFile(cfg, format); err != nil {
				log.Fatalf("%s", err)
			}
			os.Exit(0)
		}
		if redo := utils.Must(flags.GetBool("redo")); redo {
			redoMessage(cmd)
		}
		mainMode(cmd, args, cfg)
	}
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Println("init", args)
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	format, _ := cmd.Flags().GetString("config-format")
	cfg, err := config.Init(dryRun)
	switch format {
	case "yaml", "yml", "toml":
		break
	default:
		log.Fatalf("unsupported default config-file format: %s", format)
	}
	if err != nil {
		log.Fatalf("%s", err)
	}
	if err := config.InitDefaultCfgFile(cfg, format); err != nil {
		log.Fatalf("%s", err)
	}
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "initialize a config file if none is present",
		Run:   runInit,
	}
}

func Cmd(version_, commit_, date_ string) (cmd *cobra.Command) {
	version = version_
	commit = commit_
	date = date_

	cmd = &cobra.Command{
		Use:   "git-cc",
		Short: "write conventional commits",
		Run:   run,
	}
	{ // flags for git-cc
		flags := cmd.Flags()
		flags.BoolP("help", "h", false, "print the usage of git-cc")
		flags.Bool("dry-run", false, "Only print the resulting conventional commit message; don't commit.")
		flags.Bool("redo", false, "Reuse your last commit message")
		flags.StringArrayP("message", "m", []string{}, "pass a complete conventional commit. If valid, it'll be committed without editing.")
		flags.Bool("version", false, "print the version")
		flags.Bool("show-config", false, "print the path to the config file and the relevant config ")
		flags.Bool("allow-empty", false, "delegated to git-commit")
		// TODO: accept more of git commit's flags; see https://git-scm.com/docs/git-commit
		// likely: --cleanup=<mode>
		// more difficult, and possibly better done manually: --amend, -C <commit>
		// --reuse-message=<commit>, -c <commit>, --reedit-message=<commit>,
		// --fixup=<commit>, --squash=<commit>
		flags.String("author", "", "delegated to git-commit")
		flags.String("date", "", "delegated to git-commit")
		flags.BoolP("all", "a", false, "see the git-commit docs for --all|-a")
		flags.BoolP("signoff", "s", false, "see the git-commit docs for --signoff|-s")
		flags.Bool("no-gpg-sign", false, "see the git-commit docs for --no-gpg-sign")
		// FIXME: gpg-sign
		flags.Bool("no-post-rewrite", false, "Bypass the post-rewrite hook")
		flags.Bool("no-edit", false, "Use the selected commit message without launching an editor.")
		flags.BoolP("no-verify", "n", false, "Bypass git hooks")
		flags.Bool("verify", true, "Ensure git hooks run")
		// https://git-scm.com/docs/git-commit#Documentation/git-commit.txt---no-verify
		flags.Bool("no-signoff", true, "Don't add a a `Signed-off-by` trailer to the commit message")
		flags.Bool("generate-man-page", false, "Generate a man page in your manpath")
		flags.Bool(
			"generate-shell-completion",
			false,
			"print a bash/zsh/fish/powershell completion script to stdout",
		)
		flags.Bool("init", false, "initialize a config file if none is present")
		flags.String("config-format", "yaml", "The format of the config file to generate. One of: toml, yml, yaml")

		cmd.MarkFlagsMutuallyExclusive("signoff", "no-signoff")
		cmd.MarkFlagsMutuallyExclusive("verify", "no-verify")
	}
	cmd.AddCommand(initCmd())
	return cmd
}
