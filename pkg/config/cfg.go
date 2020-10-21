package config

import (
	"log"

	"github.com/spf13/viper"
)

// var cfg = viper.New()

type Described struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
}

type Cfg struct {
	CommitTypes []Described `mapstructure:"commit_types"`
	Scopes      []Described `mapstructure:"scopes"`
}

// viper: need to deserialize YAML commit-type options
// viper: need to deserialize YAML scope options
func Init() *viper.Viper {
	cfg := viper.New()
	cfg.SetConfigName("commit_convention")
	cfg.SetConfigType("yaml")
	cfg.AddConfigPath(".")
	cfg.AddConfigPath("$HOME")
	// see https://github.com/angular/angular.js/blob/master/DEVELOPERS.md#type
	// see https://github.com/conventional-changelog/commitlint/blob/master/%40commitlint/config-conventional/index.js#L23
	cfg.SetDefault("commit_types", []Described{
		{"feat", "adds a new feature"},
		{"fix", "fixes a bug"},
		{"docs", "changes only the documentation"},
		{"style", "changes the style but not the meaning of the code (such as formatting)"},
		{"perf", "improves performance"},
		{"test", "adds or corrects tests"},
		{"build", "changes the build system or external dependencies"},
		{"chore", "changes outside the code, docs, or tests"},
		{"ci", "changes to the Continuous Inegration (CI) system"},
		{"refactor", "changes the code without changing behavior"},
		{"revert", "reverts prior changes"},
	})
	cfg.SetDefault("scopes", []Described{})
	// TODO: use env vars?

	// TODO: use git commit's flag-args
	// -a, --all
	// --no-edit
	// --amend
	// --no-post-rewrite
	// --dry-run
	// --no-gpg-sign
	// -s, --signoff
	// --author=<author>
	// --date=<date>

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

// func GetCommitTypes() ([]string, []string) {
// 	// TODO:!
// }
