package cmd

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// Long: ,
func generateShellCompletion(cmd *cobra.Command, args []string) {
	var shell string
	if len(args) > 0 {
		shell = args[0]
	} else {
		shell = path.Base(os.Getenv("SHELL"))
	}
	switch shell {
	case "bash":
		cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		cmd.Root().GenPowerShellCompletion(os.Stdout)
	default:
		log.Fatal(fmt.Errorf("unknown/unsupported shell `%s`", shell))
	}
}

func generateManPage(cmd *cobra.Command, args []string) {
	root := cmd.Root()
	header := &doc.GenManHeader{
		Title:   "GIT-CC",
		Section: "1",
	}
	var out bytes.Buffer
	process := exec.Command("manpath")
	process.Stdout = &out
	err := process.Run()
	if err != nil {
		log.Fatal(err)
	}
	manpath := strings.Split(out.String(), ":")
	for _, place := range manpath {
		err = doc.GenManTree(root, header, path.Join(place, "man1"))
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Fatal(err)
	}
	// done, otherwise
}
