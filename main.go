package main

import "github.com/skalt/git-cc/cmd"

// provided by goreleaser; see .goreleaser.yml
var version string = "no version provided"

func main() {
	// here's where I'd do an ldflags injection
	cmd.SetVersion(version)
	cmd.Cmd.Execute()
}
