package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func captureOutput(fn func()) string {
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestGenerateShellCompletion(t *testing.T) {
	cmd := Cmd("test", "abc123", "2024-01-01")

	shells := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name:     "bash",
			args:     []string{"bash"},
			contains: []string{"bash completion", "__start_git-cc", "complete -o default -F"},
		},
		{
			name:     "zsh",
			args:     []string{"zsh"},
			contains: []string{"#compdef git-cc", "_git-cc", "compdef _git-cc git-cc"},
		},
		{
			name:     "fish",
			args:     []string{"fish"},
			contains: []string{"# fish completion", "complete -c git-cc"},
		},
		{
			name:     "powershell",
			args:     []string{"powershell"},
			contains: []string{"Register-ArgumentCompleter", "git-cc"},
		},
	}

	for _, shell := range shells {
		t.Run(shell.name, func(t *testing.T) {
			output := captureOutput(func() {
				generateShellCompletion(cmd, shell.args)
			})
			if output == "" {
				t.Fatalf("generateShellCompletion(%s) produced no output", shell.name)
			}
			for _, want := range shell.contains {
				if !strings.Contains(output, want) {
					t.Errorf(
						"generateShellCompletion(%s) output missing %q",
						shell.name, want,
					)
				}
			}
		})
	}
}

func TestGenerateShellCompletion_DefaultsToEnvShell(t *testing.T) {
	cmd := Cmd("test", "abc123", "2024-01-01")

	origShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", origShell)

	t.Run("bash from env", func(t *testing.T) {
		os.Setenv("SHELL", "/bin/bash")
		output := captureOutput(func() {
			generateShellCompletion(cmd, nil)
		})
		if !strings.Contains(output, "bash completion") {
			t.Error("expected bash completion output from SHELL=/bin/bash")
		}
	})

	t.Run("zsh from env", func(t *testing.T) {
		os.Setenv("SHELL", "/bin/zsh")
		output := captureOutput(func() {
			generateShellCompletion(cmd, nil)
		})
		if !strings.Contains(output, "#compdef git-cc") {
			t.Error("expected zsh completion output from SHELL=/bin/zsh")
		}
	})
}

func TestCmd_HasShellCompletionFlag(t *testing.T) {
	cmd := Cmd("test", "abc123", "2024-01-01")
	flag := cmd.Flags().Lookup("generate-shell-completion")
	if flag == nil {
		t.Fatal("expected --generate-shell-completion flag to be defined")
	}
	if flag.DefValue != "false" {
		t.Errorf("expected default value 'false', got %q", flag.DefValue)
	}
}

func TestCmd_HasAllCompletionFlags(t *testing.T) {
	cmd := Cmd("test", "abc123", "2024-01-01")
	expectedFlags := []string{
		"generate-shell-completion",
		"generate-man-page",
		"print-schema",
		"version",
		"help",
		"dry-run",
		"redo",
		"message",
		"allow-empty",
		"all",
		"signoff",
		"no-verify",
		"verify",
		"no-gpg-sign",
		"no-post-rewrite",
		"no-edit",
		"author",
		"date",
		"show-config",
		"init",
		"config-format",
		"no-signoff",
	}

	for _, name := range expectedFlags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("expected flag %q to be defined", name)
		}
	}
}

func TestCmdInit(t *testing.T) {
	cmd := initCmd()
	if cmd.Use != "init" {
		t.Errorf("expected Use 'init', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be non-empty")
	}
}

func TestCmdRootStructure(t *testing.T) {
	cmd := Cmd("test", "abc123", "2024-01-01")
	if cmd.Use != "git-cc" {
		t.Errorf("expected Use 'git-cc', got %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected Short description to be non-empty")
	}
	if len(cmd.Commands()) == 0 {
		t.Error("expected at least one subcommand (init)")
	}

	found := false
	for _, sub := range cmd.Commands() {
		if sub.Use == "init" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'init' subcommand to be registered")
	}
}

func TestMutuallyExclusiveFlags(t *testing.T) {
	cmd := Cmd("test", "abc123", "2024-01-01")
	if cmd.Flags().Lookup("signoff") == nil {
		t.Error("expected 'signoff' flag")
	}
	if cmd.Flags().Lookup("no-signoff") == nil {
		t.Error("expected 'no-signoff' flag")
	}
	if cmd.Flags().Lookup("verify") == nil {
		t.Error("expected 'verify' flag")
	}
	if cmd.Flags().Lookup("no-verify") == nil {
		t.Error("expected 'no-verify' flag")
	}
}

func TestCompletionScriptContainsRootCommandName(t *testing.T) {
	cmd := Cmd("test", "abc123", "2024-01-01")
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			output := captureOutput(func() {
				generateShellCompletion(cmd, []string{shell})
			})
			if !strings.Contains(output, "git-cc") {
				t.Errorf(
					"completion script for %s should reference 'git-cc'",
					shell,
				)
			}
		})
	}
}
