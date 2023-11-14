# git-cc

A git extension to help write [conventional commits][cc-standard]

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

`git-cc` searches for a configuration file named `commit_convention.{yaml,yml,toml}`.
Note that `git-cc` prefers the extension `yaml` over `yml`, and `yml` over `toml`.


`git-cc` searches the following directories for a configuration file in this order:

```
${PWD}/
${REPO_ROOT}/         # ignored if not inside a git repo
${REPO_ROOT}/.config/ # ignored if not inside a git repo
${XDG_CONFIG_HOME}/
```

See [`./config/commit_convention.yaml`](./.config/commit_convention.yaml) for an example configuration file.

## Why write conventional commits through an interactive CLI?

Figuring out what to write for an informative commit can be difficult.
The conventional commits standard helps figure out what to write.
As a bonus, conventional commits facilitate generating changelogs and semantic version increments.
An interactive command-line application helps with following the standard.

## Licensing

For the details, see [./LICENSE.md](./LICENSE.md).

`git-cc` is source-provided software, not open-source software: it's offered under either [the PolyForm Free Trial 1.0.0 license](./licenses/PolyForm-Free-Trial-1.0.0.md), [the PolyForm Noncommercial 1.0.0 license](./licenses/PolyForm-Noncommercial-1.0.0.md), or a per-developer commercial license.
This means anyone can use the software for noncommercial purposes and trial the software for commercial use for a month.
Anyone using the software to make money after their trial expires needs to purchase a commercial license.

You can purchase a commercial license [via a one-time GitHub sponsorship](https://github.com/sponsors/skalt/sponsorships?tier_id=335824).


<details><summary>
  Prior licenses apply to previous versions of <code>git-cc</code>.
</summary>

- From 2021-08-25 to 2023-11-14, `git-cc` was offered under the Indie Code Catalog Standard Deal, version 4.0.1 or higher. The licenses can be found at commit [`b682f27e`](https://github.com/SKalt/git-cc/commit/b682f27e1fb9d3c06f9bfbcc9403e4343e739a11).
IndieCC shut down, prompting relicensing.

- From 2020-11-23 to 2021-08-25, `git-cc` was licensed under the strictEq free and paid licenses. The licenses can be found at commit [`7ca8e95`](https://github.com/SKalt/git-cc/blob/7ca8e95c1dc25ee87eeeb0d73f0b2ddef1aa9f29/LICENSE.md). StrictEq rebranded to indieCC, prompting relicensing.

If you bought an indieCC or strictEq paid license, that license still grants you rights to copy and use updated versions `git-cc` even after the license change.

</details>


## Installation

### Manual installation

Go to [the project releases page][releases page] to download the appropriate packaging format.
Please verify the shasum of the downloaded executable for you before you run it.

### Using `brew`

```sh
brew tap skalt/git-cc
brew install git-cc
```

### Using the installer script for Linux or Mac

First, take a second to verify that the installer script would correctly check the shasum of the downloaded package.
Then, run the following commands:

<!-- TODO: automate populating outputs with `cog` or similar -->

```sh
repo=skalt/git-cc
branch=master
curl -sL https://raw.githubusercontent.com/$repo/$branch/scripts/install.sh > /tmp/install.sh
shasum -a 256 /tmp/install.sh | sed 's/^/# /g'
# 805354a9b0db5648af785086af5a4af036d5281ea5999bda055c888186d36cd7  /tmp/install.sh
```

```sh
chmod +x /tmp/install.sh
/tmp/install.sh --help | sed 's/^/# /g'
```
```txt
# USAGE: ./install.sh [-h|--help] [--download-only|--dry-run] [FMT]
# download a release of git-cc for your OS and instruction set architecture.
#
# ARGS:
#   -h|--help        print this message and exit
#   --download-only  download as FMT, but do not install
#   --dry-run        print rather than follow the download url for the binary
#   FMT              The download format. Valid values are
#                      - tar.gz (default)
#                      - apk
#                      - brew
#                      - deb
#                      - exe
#                      - rpm
```

```sh
/tmp/install.sh
```

### From source with go

To compile from source, run `make install` inside the source directory.
You'll need to have a `go >= 1.19` toolchain and to have your `$GOPATH/bin` on your `$PATH`.


## Prior art:

- [`@commitlint/prompt-cli`][commitlint]
- [`committizen`][commitizen]
- [`commitsar`][commitsar]

<!-- links -->

[commitlint]: https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/prompt-cli
[cc-standard]: https://www.conventionalcommits.org/en/v1.0.0/
[commitizen]: https://github.com/commitizen/cz-cli
[commitlint]: https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/config-conventional
[commitsar]: https://github.com/commitsar-app/commitsar
[releases page]: https://github.com/skalt/git-cc/releases/latest
