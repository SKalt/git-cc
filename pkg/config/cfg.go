package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/muesli/termenv"
	"github.com/spf13/viper"
)

var (
	// see https://github.com/angular/angular.js/blob/master/DEVELOPERS.md#type
	// see https://github.com/conventional-changelog/commitlint/blob/master/%40commitlint/config-conventional/index.js#L23
	AngularPresetCommmitTypes = []map[string]string{
		{"feat": "adds a new feature"},
		{"fix": "fixes a bug"},
		{"docs": "changes only the documentation"},
		{"style": "changes the style but not the meaning of the code (such as formatting)"},
		{"perf": "improves performance"},
		{"test": "adds or corrects tests"},
		{"build": "changes the build system or external dependencies"},
		{"chore": "changes outside the code, docs, or tests"},
		{"ci": "changes to the Continuous Inegration (CI) system"},
		{"refactor": "changes the code without changing behavior"},
		{"revert": "reverts prior changes"},
	}
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

func HelpBar(s ...string) string {
	return Faint(fmt.Sprintf("\n%s", strings.Join(s, "; ")))
}

type Cfg struct {
	CommitTypes     []map[string]string `mapstructure:"commit_types"`
	Scopes          []map[string]string `mapstructure:"scopes"`
	HeaderMaxLength int                 `mapstructure:"header_max_length"`
	//^ named similar to conventional-changelog/commitlint
	EnforceMaxLength bool `mapstructure:"enforce_header_max_length"`
}

// viper: need to deserialize YAML commit-type options
// viper: need to deserialize YAML scope options
func Init() *viper.Viper {
	cfg := viper.New()
	cfg.SetConfigName("commit_convention")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath(".")
	cfg.AddConfigPath("$HOME")

	cfg.SetDefault("commit_types", AngularPresetCommmitTypes)
	cfg.SetDefault("scopes", map[string]string{})
	cfg.SetDefault("header_max_length", 72)
	cfg.SetDefault("enforce_header_max_length", false)
	// s.t. git log --oneline should remain within 80 columns w/ a 7-rune
	// commit hash and one space before the commit message.
	// this caps the max len of the `type(scope): description`, not the body
	// TODO: use env vars?

	return cfg
}

func Lookup(cfg *viper.Viper) Cfg {
	err := cfg.ReadInConfig()
	if err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			// can fail safely, we have defaults
			break
		default:
			log.Fatal(err)
		}
	}
	var data Cfg
	err = cfg.Unmarshal(&data)
	if err != nil {
		log.Fatal(err)
	}
	return data
}

func getGitVar(var_name string) (string, error) {
	cmd := exec.Command("git", "var", var_name)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	} else {
		return strings.TrimRight(out.String(), " \t\r\n"), err
	}
}

func GetEditor() string {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor
	}
	return "vi"
}

// search GIT_EDITOR, then fall back to $EDITOR
func GetGitEditor() string {
	editor, err := getGitVar("GIT_EDITOR") // TODO: shell-split the string
	if err != nil {
		return GetEditor()
	}
	return editor
}

func GetCommitMessageFile() string {
	cmd := exec.Command("git", "rev-parse", "--absolute-git-dir")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimRight(out.String(), " \t\r\n") + string(os.PathSeparator) + "COMMIT_EDITMSG"
}
