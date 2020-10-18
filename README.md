# git-cc
> a git extension to help write [conventional commits][cc-standard]

## install

```sh
go install github.com/skalt/git-cc

git cc feat added conventional commits
# --> "feat: added conventional commits"
git cc feat: added conventional commits
# --> "feat: added conventional commits"
git cc fix something: an error
# --> "fix(something): an error"

git cc feat
# interactively write a conventional commit from the description onward
git cc
# interactively write a conventional commit!
```
<!-- 
## Why Conventional Commits?
The cool kids are doing it.
Plus, `cc`s make good, `grep`pable changelogs. -->

## Why An Interactive Cli? 
Figuring out what to write for an informative commit can be difficult.
The convential commits standard helps figure out what to write.
An interactive prompts helps with following the standard.

<details><summary>Some parts of the conventional commit standard require quoting to work on the command-line.</summary>

- `git commit -m fix(something): ...` fails since `()` would be a syntactically-invalid subshell
- `git commit -m feat!: ...` fails since `!` expands to a bash history command
</details>

You want the same interface in any project, on any machine. You want to use your git log as your changelog.


Prior art:

- [`committizen`][commitizen]
- [`commitsar`][commitsar]

[cc-standard]: https://www.conventionalcommits.org/en/v1.0.0/

[commitizen]: https://github.com/commitizen/cz-cli
[commitlint]: https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/config-conventional
[commitsar]: https://github.com/commitsar-app/commitsar