package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
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
	if version == "" {
		version = "<unknown>"
	}
	s := strings.Builder{}
	utils.Must(s.WriteString("git-cc "))
	utils.Must(s.WriteString(version))
	utils.Must(s.WriteString(" commit "))
	utils.Must(s.WriteString(commit))
	utils.Must(s.WriteString(" built "))
	utils.Must(s.WriteString(date))
	utils.Must(s.WriteRune('\n'))
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
func doCommit(message string, dryRun bool, commitParams []string) (err error) {
	f := config.GetCommitMessageFile()
	if !dryRun {
		file, err := os.Create(f)
		if err != nil {
			return fmt.Errorf("unable to create %s: %w", f, err)
		}
		_, err = file.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("unable to write to %s: %w", f, err)
		}
	}
	argv := append([]string{"git", "commit", "--message", message}, commitParams...)
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if dryRun {
		shellQuoted := strings.ReplaceAll(message, `"`, `\"`)
		shellQuoted = strings.ReplaceAll(shellQuoted, "`", "\\`")
		fmt.Printf(
			"# would run:\ngit commit --message \\\n  '%s' \\\n  %s",
			shellQuoted,
			strings.Join(commitParams, " "),
		)
		return nil
	} else {
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed running `%+v`: %w", argv, err)
		} else {
			return
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
func mainMode(cmd *cobra.Command, args []string, cfg *config.Cfg) (err error) {
	commitParams := getGitCommitCmd(cmd)
	committingAllChanges := utils.Must(cmd.Flags().GetBool("all"))
	allowEmpty := utils.Must(cmd.Flags().GetBool("allow-empty"))
	if !cfg.DryRun && !committingAllChanges {
		buf := &bytes.Buffer{}
		process := exec.Command("git", "diff", "--name-only", "--cached")
		process.Stdout = buf
		err = process.Run()
		if err != nil {
			return fmt.Errorf("fatal: not a git repository (or any of the parent directories): .git; %w", err)
		}
		if buf.String() == "" && !allowEmpty {
			return errors.New("no files staged")
		}
	}

	message := utils.Must(cmd.Flags().GetStringArray("message"))
	var toParse string
	if len(message) > 0 {
		toParse = strings.Join(message, "\n\n")
		//> If multiple `-m` options are given, their values are concatenated as separate paragraphs.
		//> see https://git-scm.com/docs/git-commit#Documentation/git-commit.txt---messageltmsggt
	} else {
		toParse = strings.Join(args, " ")
	}
	cc, _ := parser.ParseAsMuchOfCCAsPossible(toParse)
	var validationErrors ValidationErrors = 0
	if cc.Type == "" {
		validationErrors |= InvalidType
	} else {
		if _, valid := cfg.CommitTypes.Get(cc.Type); !valid {
			validationErrors |= InvalidType
		}
	}
	if cc.Scope != nil {
		if _, valid := cfg.Scopes.Get(*cc.Scope); !valid {
			validationErrors |= InvalidScope
		}
	}
	if cc.Description == "" {
		validationErrors |= MissingDescription
	}

	if validationErrors != 0 {
		m := initialModel(&cc, cfg)
		ui := tea.NewProgram(m)
		out, err := ui.Run()
		if err != nil {
			return err
		}
		result := out.(model)
		if !result.ready() {
			return errors.New("cancelled") // no submission
		} else {
			commitMessage := result.value()
			f := config.GetCommitMessageFile()
			file, err := os.Create(f)
			if err != nil {
				log.Fatalf("unable to create fil %s: %+v", f, err)
			}
			_, err = file.Write([]byte(commitMessage))
			if err != nil {
				return fmt.Errorf("unable to write to file %s: %w", f, err)
			}
			return doCommit(commitMessage, cfg.DryRun, commitParams)
		}
	}
	return doCommit(cc.ToString(), cfg.DryRun, commitParams)
}

func redoMessage(cmd *cobra.Command) (err error) {
	flags := cmd.Flags()
	msg := utils.Must(flags.GetStringArray("message"))
	if len(msg) > 0 { // FIXME: do this with flag groups
		log.Fatal("-m|--message is incompatible with --redo")
	}
	commitMessagePath := config.GetCommitMessageFile()
	preExisting, err := os.ReadFile(commitMessagePath)
	if err != nil {
		return fmt.Errorf("unable to read file %q: %w", commitMessagePath, err)
	}
	preExisting = []byte(strings.TrimSpace(string(preExisting)))
	empty := true
	message := make([]byte, 0, len(preExisting))
	for line := range strings.SplitSeq(string(preExisting), "\n") {
		trimmed := strings.TrimLeft(line, " \t\r\n")
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		empty = false
		message = append(message, []byte(line)...)
		message = append(message, byte('\n'))
	}
	if empty {
		return fmt.Errorf("empty commit message: %q", commitMessagePath)
	}

	utils.Check(flags.Set("message", string(message)))
	return
}

// Note: I'm not using cobra subcommands since they prevent passing arbitrary arguments,
// and I'd like to be able to start an invocation like `git-cc this is the commit message`
// without having to think about whether `this` is a subcommand.

func run(cmd *cobra.Command, args []string) (err error) {
	flags := cmd.Flags()
	if shouldPrintVersion, _ := flags.GetBool("version"); shouldPrintVersion {
		printVersion(version, commit, date)
		return
	} else if genCompletion, _ := flags.GetBool("generate-shell-completion"); genCompletion {
		generateShellCompletion(cmd, args)
		return
	} else if genManPage, _ := flags.GetBool("generate-man-page"); genManPage {
		generateManPage(cmd, args)
		return
	} else if shouldPrintSchema, _ := flags.GetBool("print-schema"); shouldPrintSchema {
		fmt.Println(config.Schema)
		return
	} else {
		initLogger()
		defer func() {
			if debugLogFile != nil {
				utils.Check(debugLogFile.Close())
			}
		}()

		dryRun := utils.Must(cmd.Flags().GetBool("dry-run"))
		var cfg *config.Cfg
		cfg, err = config.Init(dryRun, logger)
		if err != nil {
			return err
		}
		if showConfig, _ := flags.GetBool("show-config"); showConfig {
			repoRoot, _ := config.GetGitRepoRoot()
			_, err = config.FindCCConfigFile(repoRoot, logger)
			if err != nil {
				return err
			}
			file := cfg.ConfigFile
			if file == "" {
				file = "<default>"
			}
			fmt.Printf("config file path: %s\n", file)
			return
		}
		if init := utils.Must(flags.GetBool("init")); init {
			format := utils.Must(cmd.Flags().GetString("config-format"))
			switch format {
			case "yaml", "yml", "toml":
				break
			default:
				log.Fatalf("unsupported default config-file format: %s", format)
			}
			if err != nil {
				return err
			}
			if err = config.InitDefaultCfgFile(cfg, format); err != nil {
				return err
			}
			return
		}
		if redo := utils.Must(flags.GetBool("redo")); redo {
			if err = redoMessage(cmd); err != nil {
				return
			}
		}
		return mainMode(cmd, args, cfg)
	}
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Println("init", args)
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	format, _ := cmd.Flags().GetString("config-format")
	cfg, err := config.Init(dryRun, nil)
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

func completion(
	cmd *cobra.Command, args []string, toComplete string,
) (completions []cobra.Completion,
	sc cobra.ShellCompDirective,
) {
	sc = cobra.ShellCompDirectiveNoFileComp
	cfg, err := config.Init(true, slog.New(slog.DiscardHandler))
	if err != nil {
		completions = append(completions, err.Error())
		return
	}
	switch len(args) {
	case 0:
		for _, ct := range config.Items(cfg.CommitTypes) {
			if strings.HasPrefix(ct[0], toComplete) {
				completions = append(completions, cobra.CompletionWithDesc(ct[0], ct[1]))
			}
		}
	case 1:
		switch {
		case strings.HasSuffix(toComplete, ":"):
			return
		case strings.HasSuffix(toComplete, "!"):
			completions = append(completions, ":")
			return
		case strings.HasSuffix(toComplete, ")"):
			completions = append(completions, cobra.CompletionWithDesc("!", "breaking change"), ":")
			return
		case strings.Contains(toComplete, "("):
			items := config.Items(cfg.Scopes)
			completions = make([]cobra.Completion, 0, len(items)+1)
			completions = append(
				completions,
				cobra.CompletionWithDesc("!", "breaking change"),
			)
			for _, i := range items {
				completions = append(completions, cobra.CompletionWithDesc(i[0], i[1]))
			}
			return
		default:
		}
	}
	return
}

func Cmd(version_, commit_, date_ string) (cmd *cobra.Command) {
	version = version_
	commit = commit_
	date = date_

	cmd = &cobra.Command{
		Use:   "git-cc",
		Short: "write conventional commits",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(cmd, args); err != nil {
				utils.Must(fmt.Fprint(os.Stderr, err.Error()))
				if errors.Is(err, fs.ErrNotExist) {
					os.Exit(127)
				} else if _, ok := errors.AsType[*exec.Error](err); ok {
					os.Exit(127)
				} else {
					os.Exit(1)
				}
			}
		},
	}
	cmd.CompletionOptions.DisableDefaultCmd = true
	cmd.SetHelpCommand(&cobra.Command{Hidden: true})
	cmd.ValidArgsFunction = completion
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
		flags.Bool("print-schema", false, "print the schema of the config file to stdout")
		flags.Bool("init", false, "initialize a config file if none is present")
		flags.String("config-format", "yaml", "The format of the config file to generate. One of: toml, yml, yaml")

		cmd.MarkFlagsMutuallyExclusive("signoff", "no-signoff")
		cmd.MarkFlagsMutuallyExclusive("verify", "no-verify")
	}
	cmd.AddCommand(initCmd())
	return cmd
}
