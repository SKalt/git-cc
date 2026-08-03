package main

import (
	"fmt"
	"os"

	"github.com/skalt/git-cc/cmd"
	"github.com/skalt/git-cc/internal/utils"
)

// provided by goreleaser; see .goreleaser.yml & https://goreleaser.com/cookbooks/using-main.version/
var (
	version, commit, date string
)

func main() {
	if err := cmd.Cmd(version, commit, date).Execute(); err != nil {
		utils.Must(fmt.Fprintf(os.Stderr, "Fatal: %s\n", err.Error()))
		os.Exit(1)
	}
}
