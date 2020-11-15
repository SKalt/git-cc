# git-cc
> a git extension to help write [conventional commits][cc-standard]

## Installation

<details open><summary>With Go</summary>

```sh
go install github.com/skalt/git-cc
```

</details>
<!-- TODO: deb/apk/rpm instructions -->

## Usage

<!-- TODO: gif/video of usage  -->
```sh
# you can invoke either of the below to
# interactively write a conventional commit
git conventional-commit
git cc

git cc feat          # start after the commit-type
git cc 'feat(scope)' # start after the scope

# or validate your conventional commit
git cc feat: added conventional commits # ok! creates a commit
git cc 'feat(cli): added a conventional commit' # ok! creates a commit
git cc feat add a typo  # starts interaction at the scope
git cc -m "invalid(stuff): should return 1"
```

## Why write conventional commits through an interactive CLI? 
Figuring out what to write for an informative commit can be difficult.
The convential commits standard helps figure out what to write.
As a bonus, conventional commits facilitate generating changelogs and semantic version increments.
An interactive command-line application helps with following the standard.

<details><summary>Some parts of the conventional commit standard require quoting to work on the command-line.</summary>

- `git commit -m fix(something): ...` fails since `()` would be a syntactically-invalid subshell
- `git commit -m feat!: ...` fails since `!` expands to a bash history command
</details>

You want the same interface in any project, on any machine. You want to use your git log as your changelog.


Prior art:

- [`committizen`][commitizen]
- [`commitsar`][commitsar]

<!-- links -->
[cc-standard]: https://www.conventionalcommits.org/en/v1.0.0/

[commitizen]: https://github.com/commitizen/cz-cli
[commitlint]: https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/config-conventional
[commitsar]: https://github.com/commitsar-app/commitsar