# git-cc
a git extension to help write [conventional commits][cc-standard]

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/skalt/git-cc)
![Go report card](https://goreportcard.com/badge/github.com/skalt/git-cc)


## Usage

![demo video](./assets/demo.gif)

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


## Installation

Go to [the project releases page][releases page] to download the appropriate package. Alternately, use the installer script at [./scripts/download_release.sh](./scripts/download_release.sh).  Note that (and please verify) the installer script checks the shasum of the downloaded executable for you before you run it.

<details><summary>on linux or mac/darwin</summary>

```sh
./download_release.sh "tar.gz"     &&
  tar -xf ./git-cc*.tar.gz -C /tmp &&
  mv /tmp/git-cc /usr/local/bin/   &&
  rm -rf ./git-cc*.tar.gz
```
</details>

<details><summary>on debian</summary>

```sh
./download_release.sh "deb"  &&
  sudo dpkg -i ./git-cc*.deb &&
  rm ./git-cc*.deb
```
</details>
<details><summary>on centos/RHEL/Amazon Linux</summary>

```sh
./download_release.sh "rpm" &&
  sudo rpm -i ./git-cc*.rpm &&
  rm ./git-cc*.rpm
```
</details>
<details open><summary>From source with go</summary>

To compile from source, run `make install` inside the source directory.
You'll need to have a `go >= 1.13` toolchain and to have your `$GOPATH/bin` on your `$PATH`.
</details>

## Prior art:

- [`committizen`][commitizen]
- [`commitsar`][commitsar]

<!-- links -->
[cc-standard]: https://www.conventionalcommits.org/en/v1.0.0/

[commitizen]: https://github.com/commitizen/cz-cli
[commitlint]: https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/config-conventional
[commitsar]: https://github.com/commitsar-app/commitsar
[releases page]: https://github.com/skalt/git-cc/releases/latest
