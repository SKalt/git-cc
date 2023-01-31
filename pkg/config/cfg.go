package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/mapstructure"
	"github.com/muesli/termenv"
	yaml "gopkg.in/yaml.v3"
)

const ExampleCfgFileHeader = `## commit_convention.yml
## omit the commit_types to use the default angular-style commit types`
const ExampleCfgFileCommitTypes = `
# commit_types:
#   - type: description of what the short-form "type" means`
const ExampleCfgFileScopes = `
# scopes:
#   - scope: description of what the short-form "scope" represents`
const ExampleCfgFile = ExampleCfgFileHeader + ExampleCfgFileCommitTypes + ExampleCfgFileScopes

var (
	// see https://github.com/angular/angular.js/blob/master/DEVELOPERS.md#type
	// see https://github.com/conventional-changelog/commitlint/blob/master/%40commitlint/config-conventional/index.js#L23
	AngularPresetCommitTypes = []map[string]string{
		{"feat": "adds a new feature"},
		{"fix": "fixes a bug"},
		{"docs": "changes only the documentation"},
		{"style": "changes the style but not the meaning of the code (such as formatting)"},
		{"perf": "improves performance"},
		{"test": "adds or corrects tests"},
		{"build": "changes the build system or external dependencies"},
		{"chore": "changes outside the code, docs, or tests"},
		{"ci": "changes to the Continuous Integration (CI) system"},
		{"refactor": "changes the code without changing behavior"},
		{"revert": "reverts prior changes"},
	}
	CentralStore *Cfg
)

const (
	HelpSubmit = "submit: tab/enter"
	HelpBack   = "go back: shift+tab"
	HelpCancel = "cancel: ctrl+c"
	HelpSelect = "navigate: up/down"
)

func Faint(s string) string {
	return termenv.String(s).Faint().String()
}

type Cfg struct {
	gitRepoRoot string
	gitDir      string
	configFile  string
	CommitTypes []map[string]string `mapstructure:"commit_types"`
	Scopes      []map[string]string `mapstructure:"scopes"`
	// this caps the max len of the `type(scope): description`, not the body
	// naming inspired by conventional-changelog/commitlint
	HeaderMaxLength  int  `mapstructure:"header_max_length"`
	EnforceMaxLength bool `mapstructure:"enforce_header_max_length"`
	DryRun           bool
}

func (original *Cfg) merge(other *Cfg) {
	if other.configFile != "" {
		original.configFile = other.configFile
	}
	if len(other.CommitTypes) > 0 {
		original.CommitTypes = other.CommitTypes
	}
	if len(other.Scopes) > 0 {
		original.Scopes = other.Scopes
	}
	original.EnforceMaxLength = other.EnforceMaxLength
	if other.HeaderMaxLength > 0 {
		original.HeaderMaxLength = other.HeaderMaxLength
	}
}

// Find &/ read the configuration file into the passed config object
func (cfg *Cfg) ReadCfgFile() (err error) {
	configFile := cfg.configFile
	if configFile == "" {
		configFile, err = findCCConfigFile(cfg.gitRepoRoot)
		if err != nil {
			// fall back to defaults
			return err
		}
	}
	next, err := parseCCConfigurationFile(configFile)
	if err != nil {
		return err
	}
	cfg.merge(next)
	return err
}

func Init(dryRun bool) (*Cfg, error) {
	cfg := Cfg{
		CommitTypes:     AngularPresetCommitTypes,
		Scopes:          []map[string]string{},
		HeaderMaxLength: 72,
		//^ s.t. `git log --oneline` should remain within 80 columns w/ a 7-rune
		// commit hash and one space before the commit message.
		EnforceMaxLength: false,
		DryRun:           dryRun,
	}
	gitDir, err := getGitDir()
	if err != nil {
		if dryRun {
			// CentralStore.gitDir = "./.git"
		} else {
			// fatal since we need to be able to read/write .git/COMMIT_EDITMESSAGE
			return nil, err
		}
	}
	cfg.gitDir = gitDir
	repoRoot, err := getGitRepoRoot()
	if err != nil {
		if dryRun {
			CentralStore.gitRepoRoot = "."
		} else {
			// fatal since we need to look for configuration there
			return nil, err
		}
	}
	cfg.gitRepoRoot = repoRoot
	if err := cfg.ReadCfgFile(); err != nil {
		return nil, err
	}
	CentralStore = &cfg
	return CentralStore, nil
}

func readConfigFile(configFile string) (m map[string]interface{}, err error) {
	f, _ := os.Stat(configFile)
	if f.IsDir() {
		return nil, fmt.Errorf("found directory `%s`", configFile)
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	ext := filepath.Ext(f.Name())
	switch ext {
	case ".yaml", ".yml":
		if err = yaml.Unmarshal(data, &m); err != nil {
			return nil, err
		} else {
			return
		}
	case ".toml":
		if err = toml.Unmarshal(data, &m); err != nil {
			return nil, err
		} else {
			return
		}
	case ".json":
		if f.Name() == "package.json" {
			// allow it as a special case. Otherwise, prefer writing configuration
			// in a format that allows comments
			if err = json.Unmarshal(data, &m); err != nil {
				return nil, err
			} else {
				return
			}
		} else {
			return nil, fmt.Errorf("only package.json supported")
		}
	}
	// all file extensions should already be known when searching for config
	// files
	panic("unreachable: " + ext + " <- " + configFile)
}

func parseCCConfigurationFile(configFile string) (*Cfg, error) {
	raw, err := readConfigFile(configFile)
	if err != nil {
		return nil, err
	}
	var cfg Cfg
	if err := mapstructure.Decode(raw, &cfg); err != nil {
		return nil, err
	}
	cfg.configFile = configFile // always an absolute path
	return &cfg, nil
}

func findCCConfigFile(gitRepoRoot string) (string, error) {
	// pkgMeta := map[string]map[string]interface{}{} // cache the unmarshalled package.json/pyproject.toml for reuse
	candidateFiles := [...]string{
		"commit_convention.toml",
		"commit_convention.yaml",
		"commit_convention.yml",
		// TODO: support commitlint config
		// ".commitlintrc",
		// ".commitlintrc.json",
		// ".commitlintrc.yaml",
		// ".commitlintrc.yml",
		// TODO: handle .commitlintrc.{j,t,cj,ct}s, commitlint.config.{j,t,cj,ct}s
		// "package.json",
		// "pyproject.toml",
	}
	dirsToSearch := make([]string, 3)

	cwd, err := filepath.Abs(".")
	if err == nil {
		dirsToSearch = append(dirsToSearch, cwd)
	}
	if gitRepoRoot != "" {
		dirsToSearch = append(dirsToSearch, gitRepoRoot)
	}
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = os.Getenv("HOME")
	}
	if configHome != "" {
		dirsToSearch = append(dirsToSearch, configHome)
	}
	tried := make([]string, len(candidateFiles)*len(dirsToSearch))
	for _, dir := range dirsToSearch {
		for _, candidate := range candidateFiles {
			configFile := path.Join(dir, candidate)
			_, err := os.Stat(configFile)
			if err == nil {
				return configFile, nil
			} else {
				tried = append(tried, configFile)
			}
		}
	}
	return "", fmt.Errorf("no configuration found in %q", tried)
}

// find the root of the tree that git is working on
func getGitRepoRoot() (string, error) {
	if env := os.Getenv("GIT_WORK_TREE"); env != "" {
		// there might be a `$GIT_COMMON_DIR?`
		return env, nil
	}
	out, err := stdoutFrom("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	} else {
		return out, nil
	}
}

// find the git directory (usually ./.git)
func getGitDir() (string, error) {
	if env := os.Getenv("GIT_COMMON_DIR"); env != "" {
		return env, nil
	}
	if env := os.Getenv("GIT_DIR"); env != "" {
		return env, nil
	}
	out, err := stdoutFrom("git", "rev-parse", "--absolute-git-dir")
	if err != nil {
		return "", err
	}
	return out, nil
}

func stdoutFrom(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

func getGitVar(var_name string) (string, error) {
	out, err := stdoutFrom("git", "var", var_name)
	if err != nil {
		return "", err
	} else {
		return strings.TrimRight(out, " \t\r\n"), err
	}
}

func GetEditor() string {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor
	}
	editor = "vi"
	_, err := exec.LookPath(editor)
	if err != nil {
		msg := "unable to open the fallback editor"
		hint := "hint: set the EDITOR env variable or install vi"
		log.Fatalf(fmt.Sprintf("%s: %q\n%s\n", msg, editor, hint))
	}
	return editor
}

// search GIT_EDITOR, then fall back to $EDITOR
func GetGitEditor() string {
	editor, err := getGitVar("GIT_EDITOR") // TODO: shell-split the string?
	if err != nil {
		return GetEditor()
	}
	return editor
}

func GetCommitMessageFile() string {
	out := CentralStore.gitDir
	return strings.Join(
		[]string{strings.TrimRight(out, " \t\r\n"), "COMMIT_EDITMSG"},
		string(os.PathSeparator),
	)
}

// interactively edit the config file, if any was used.
func EditCfgFileCmd(cfg *Cfg, defaultFileContent string) *exec.Cmd {
	editCmd := []string{}
	// sometimes `$EDITOR` can be a script with spaces, like `code --wait`
	// TODO: handle quotes in `$EDITOR`?
	for _, part := range strings.Split(GetEditor(), " ") {
		if part != "" {
			editCmd = append(editCmd, part)
		}
	}
	cfgFile := cfg.configFile
	if cfgFile == "" {
		cfgFile = "commit_convention.yaml"
		f, err := os.Create(path.Join(cfg.gitRepoRoot, cfgFile))
		if err != nil {
			log.Fatalf("unable to create file %s: %+v", cfgFile, err)
		}
		_, err = f.WriteString(defaultFileContent)
		if err != nil {
			log.Fatalf("unable to write to file: %v", err)
		}
	}
	editCmd = append(editCmd, cfgFile)
	cmd := exec.Command(editCmd[0], editCmd[1:]...)
	return cmd
}
