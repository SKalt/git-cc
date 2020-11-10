package main

import "github.com/skalt/git-cc/cmd"

// TODO: var version *string
func main() {
	// here's where I'd do an ldflags injection
	cmd.Cmd.Execute()
}
