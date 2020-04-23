# git-cc
> a git extension to help write [conventional commits][cc-standard]

## quickstart

```sh
go install github.com/skalt/git-cc
git cc -init -y # need to configure git-cc within this repo. Use default types.

git add .
git cc feat added conventional commits
# equivalent to `git commit -m "feat: added conventional commits"`

git cc fix something: an error
# equivalent to `git commit -m "fix(something): an error"`
git cc feat
# interactively write a conventional commit
```

## Why?

You want the same interface in any project, on any machine. You want to use your git log as your changelog.


Prior art:

[`committizen`][commitizen].

[cc-standard]: https://www.conventionalcommits.org/en/v1.0.0/

[commitizen]: https://github.com/commitizen/cz-cli
[commitlint]: https://github.com/conventional-changelog/commitlint/tree/master/%40commitlint/config-conventional