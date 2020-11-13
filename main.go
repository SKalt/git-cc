package main

import "github.com/skalt/git-cc/cmd"

var version string = "no version provided"

// TODO: var version *string
func main() {
	// here's where I'd do an ldflags injection
	cmd.SetVersion(version)
	cmd.Cmd.Execute()
}
