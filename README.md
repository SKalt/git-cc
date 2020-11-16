# git-cc
> a git extension to help write [conventional commits][cc-standard]


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
The convential commits standard helps figure out what to write.
As a bonus, conventional commits facilitate generating changelogs and semantic version increments.
An interactive command-line application helps with following the standard.


## Installation

Go to [the project releases page][releases page] to download the appropriate package.

To compile from source, run `make install` inside the source directory.


## Prior art:

- [`committizen`][commitizen]
- [`commitsar`][commitsar]

<!-- links -->
[cc-standard]: https://www.conventionalcommits.org/en/v1.0.0/

[commitizen]: https://github.com/commitizen/cz-cli
[commitlint]: https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/config-conventional
[commitsar]: https://github.com/commitsar-app/commitsar
[releases page]: https://github.com/skalt/git-cc/releases/latest
