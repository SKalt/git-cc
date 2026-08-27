package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runComplete creates a fresh command via Cmd() and runs __complete in-process.
// It overrides the root Run to prevent os.Exit from the main logic.
func runComplete(t testing.TB, compLine string) string {
	t.Helper()
	parts := strings.Fields(compLine)
	c := Cmd("version-test", "abc123", "2024-01-01")
	if len(parts) > 1 {
		parts = parts[1:]
	}
	c.SetArgs(append([]string{"__complete"}, parts...))

	// Swap Run to prevent os.Exit in the root command's Run handler.
	// The __complete subcommand has its own Run and is unaffected.
	origRun := c.Run
	c.Run = func(cmd *cobra.Command, args []string) {}
	defer func() { c.Run = origRun }()

	var buf bytes.Buffer
	c.SetOut(&buf)
	if err := c.Execute(); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}

// parseCompletions extracts completion values from __complete output.
// Each completion line is "value\tdescription". The last line is a directive.
func parseCompletions(output string) []string {
	var completions []string
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Directive lines end with :N (e.g. ":4")
		if strings.HasPrefix(line, ":") && len(line) <= 4 {
			continue
		}
		// Skip Cobra debug lines that leak to stdout in some versions
		if strings.HasPrefix(line, "Completion ended") {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		completions = append(completions, parts[0])
	}
	return completions
}

// containsAll returns any elements of `want` not found in `items`.
func containsAll(items, want []string) []string {
	itemSet := make(map[string]bool, len(items))
	for _, item := range items {
		itemSet[item] = true
	}
	var missing []string
	for _, w := range want {
		if !itemSet[w] {
			missing = append(missing, w)
		}
	}
	return missing
}

func TestComplete_LongFlagPrefix(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "--d matches --dry-run and --date",
			line: "git-cc --d",
			want: []string{"--dry-run", "--date"},
		},
		{
			name: "--v matches --version and --verify",
			line: "git-cc --v",
			want: []string{"--version", "--verify"},
		},
		{
			name: "--no matches all --no-* flags",
			line: "git-cc --no",
			want: []string{"--no-edit", "--no-verify", "--no-gpg-sign", "--no-post-rewrite", "--no-signoff"},
		},
		{
			name: "--s matches --show-config and --signoff",
			line: "git-cc --s",
			want: []string{"--show-config", "--signoff"},
		},
		{
			name: "--g matches --generate-shell-completion and --generate-man-page",
			line: "git-cc --g",
			want: []string{"--generate-shell-completion", "--generate-man-page"},
		},
		{
			name: "--m matches --message",
			line: "git-cc --m",
			want: []string{"--message"},
		},
		{
			name: "--dry-run matches itself",
			line: "git-cc --dry-run",
			want: []string{"--dry-run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := runComplete(t, tt.line)
			completions := parseCompletions(output)
			missing := containsAll(completions, tt.want)
			if len(missing) > 0 {
				t.Errorf("missing completions %v; got %v", missing, completions)
			}
		})
	}
}

func TestComplete_AllLongFlags(t *testing.T) {
	output := runComplete(t, "git-cc --")
	completions := parseCompletions(output)

	allFlags := []string{
		"--all",
		"--allow-empty",
		"--author",
		"--config-format",
		"--date",
		"--dry-run",
		"--generate-man-page",
		"--generate-shell-completion",
		"--help",
		"--init",
		"--message",
		"--no-edit",
		"--no-gpg-sign",
		"--no-post-rewrite",
		"--no-signoff",
		"--no-verify",
		"--print-schema",
		"--redo",
		"--show-config",
		"--signoff",
		"--verify",
		"--version",
	}

	missing := containsAll(completions, allFlags)
	if len(missing) > 0 {
		t.Errorf("missing flags in completion: %v", missing)
	}
}

func TestComplete_ShortFlags(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{"-h completes to -h", "git-cc -h", "-h"},
		{"-a completes to -a", "git-cc -a", "-a"},
		{"-s completes to -s", "git-cc -s", "-s"},
		{"-n completes to -n", "git-cc -n", "-n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := runComplete(t, tt.line)
			completions := parseCompletions(output)
			for _, c := range completions {
				if c == tt.want {
					return
				}
			}
			t.Errorf("expected %q in completions, got %v", tt.want, completions)
		})
	}
}

func TestComplete_Subcommand(t *testing.T) {
	output := runComplete(t, "git-cc ini")
	completions := parseCompletions(output)
	for _, c := range completions {
		if c == "init" {
			return
		}
	}
	t.Errorf("expected 'init' in completions, got %v", completions)
}

func TestComplete_InitSubcommand(t *testing.T) {
	output := runComplete(t, "git-cc init --")
	completions := parseCompletions(output)

	// init only defines --help on itself; parent flags like --dry-run
	// and --config-format are on the root command, not the init subcommand
	for _, want := range []string{"--help"} {
		missing := containsAll(completions, []string{want})
		if len(missing) > 0 {
			t.Errorf("init subcommand missing flags %v; got %v", missing, completions)
		}
	}
	for _, c := range completions {
		if c == "--dry-run" || c == "--config-format" {
			t.Errorf("init subcommand should not suggest parent flag %q; got %v", c, completions)
		}
	}
}

func TestComplete_InitSubcommandAfterFlag(t *testing.T) {
	output := runComplete(t, "git-cc init --help")
	completions := parseCompletions(output)

	// After consuming --help, only flag completions should remain (no file paths)
	for _, c := range completions {
		if !strings.HasPrefix(c, "-") && c != "init" {
			t.Errorf("unexpected completion %q after init --help; got %v", c, completions)
		}
	}
}

func TestComplete_DirectiveFormat(t *testing.T) {
	output := runComplete(t, "git-cc --")
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		t.Fatal("no output from completion")
	}

	lastLine := lines[len(lines)-1]
	// Last line should be a directive like ":4" or ":8"
	if !strings.HasPrefix(lastLine, ":") {
		t.Errorf("expected directive in last line, got: %q", lastLine)
	}
}

func TestComplete_BooleanFlagsDontSuggestValues(t *testing.T) {
	boolFlags := []string{
		"--dry-run",
		"--redo",
		"--allow-empty",
		"--all",
		"--signoff",
		"--no-gpg-sign",
		"--no-post-rewrite",
		"--no-edit",
		"--no-verify",
		"--verify",
		"--no-signoff",
		"--version",
		"--show-config",
		"--generate-man-page",
		"--generate-shell-completion",
		"--print-schema",
		"--init",
	}

	for _, flag := range boolFlags {
		t.Run(flag+" after flag suggests other flags, not values", func(t *testing.T) {
			output := runComplete(t, "git-cc "+flag+" ")
			completions := parseCompletions(output)
			for _, c := range completions {
				if c == "true" || c == "false" {
					t.Errorf("boolean flag %q should not suggest 'true' or 'false'; got %v", flag, completions)
				}
			}
		})
	}
}

func TestComplete_NoFileCompletionForFlags(t *testing.T) {
	output := runComplete(t, "git-cc --d")
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ":") || strings.HasPrefix(line, "Completion") {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		value := parts[0]
		if !strings.HasPrefix(value, "-") {
			t.Errorf("flag completion should not include non-flag values; got %q", value)
		}
	}
}

func TestComplete_FlagAfterCompletedFlag(t *testing.T) {
	output := runComplete(t, "git-cc --dry-run --")
	completions := parseCompletions(output)

	missing := containsAll(completions, []string{"--help", "--version"})
	if len(missing) > 0 {
		t.Errorf("expected remaining flags after --dry-run, missing %v; got %v", missing, completions)
	}
}

func TestComplete_ExactFlagMatch(t *testing.T) {
	output := runComplete(t, "git-cc --generate-shell-completion")
	completions := parseCompletions(output)
	if len(completions) != 1 || completions[0] != "--generate-shell-completion" {
		t.Errorf("expected exactly [--generate-shell-completion], got %v", completions)
	}
}

func TestComplete_PartialMatchMultiple(t *testing.T) {
	output := runComplete(t, "git-cc --ver")
	completions := parseCompletions(output)
	missing := containsAll(completions, []string{"--verify", "--version"})
	if len(missing) > 0 {
		t.Errorf("--ver should match both --verify and --version, missing %v; got %v", missing, completions)
	}
}
