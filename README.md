# git-cc

a git extension to help write [conventional commits][cc-standard]

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/skalt/git-cc)
![Go report card](https://goreportcard.com/badge/github.com/skalt/git-cc)

## Usage

![demo](./assets/demo.gif)

```sh
# you can invoke either of the below to
# interactively write a conventional commit
git cc
git cc feat          # start after the commit-type
git cc 'feat(scope)' # start after the scope

# or validate your conventional commit
git cc feat: added conventional commits # ok! creates a commit
git cc 'feat(cli): added a conventional commit' # ok! creates a commit
git cc feat add a typo  # starts interaction at the scope
git cc -m "invalid(stuff): should return 1"
```

### Configuration

See [`./commit_convention.yml`](./commit_convention.yml) for an example configuration file.

## Why write conventional commits through an interactive CLI?

Figuring out what to write for an informative commit can be difficult.
The conventional commits standard helps figure out what to write.
As a bonus, conventional commits facilitate generating changelogs and semantic version increments.
An interactive command-line application helps with following the standard.

## Licensing

`git-cc` is source-provided software, not free software:

> To use software listed on this site to make money or for work, you need to buy a license. If you’re part of a team, everyone on your team who uses the software needs to buy one. You can [try before you buy](#free-trials), to make sure the software works and integrates well with your prototypes.

For more details, see [./LICENSE.md](./LICENSE.md).

You can buy a paid license at [https://indiecc.com/~skalt/git-cc](https://indiecc.com/~skalt/git-cc).

## Installation

Go to [the project releases page][releases page] to download the appropriate package. Alternately, use the installer script at [./scripts/install.sh](./scripts/install.sh).

Please verify the the shasum of the downloaded executable for you before you run it.
Alternately, verify that the installer script would correctly check the shasum of the downloaded package.

### On linux or mac/darwin

```sh
repo=skalt/git-cc
branch=master
curl -sL https://raw.githubusercontent.com/$repo/$branch/scripts/install.sh | sh
```

### Using platform-specific package managers

You can also use the install script to use platform-specific package managers. See the script's usage instructions for more details.

```sh
repo=skalt/git-cc
branch=master
curl -sL https://raw.githubusercontent.com/$repo/$branch/scripts/install.sh > /tmp/install.sh;
chmod +x /tmp/install.sh
/tmp/install.sh --help
```

### From source with go

To compile from source, run `make install` inside the source directory.
You'll need to have a `go >= 1.19` toolchain and to have your `$GOPATH/bin` on your `$PATH`.

## Prior art:

- [`committizen`][commitizen]
- [`commitsar`][commitsar]

<!-- links -->

[cc-standard]: https://www.conventionalcommits.org/en/v1.0.0/
[commitizen]: https://github.com/commitizen/cz-cli
[commitlint]: https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/config-conventional
[commitsar]: https://github.com/commitsar-app/commitsar
[releases page]: https://github.com/skalt/git-cc/releases/latest
