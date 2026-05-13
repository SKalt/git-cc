---
name: git-cc
description: Use when writing git commits with conventional commit format and the git-cc tool is available (git cc command)
---

# git-cc

CLI tool for writing and validating [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/).

> [!WARNING]
> **NEVER use interactive mode.** `git cc` without a message argument, `git cc --redo`, or `git cc` with a partial/invalid message opens a TUI that possibly block the session indefinitely. Always pass a complete conventional commit as an argument or via `-m`. If you need the TUI, use interactive CLI skills or tools.

## Basic CLI Usage

### Quick commit (preferred)

Pass a complete conventional commit — git-cc validates and delegates to `git commit`:

```bash
git cc feat: add login page
git cc 'fix(parser): handle empty input'
git cc 'refactor(auth)!: drop legacy token format'
```

### Validate via -m

```bash
git cc -m "feat: added search"
git cc -m "feat(api): add user endpoint"
```

Multiple `-m` flags concatenate as paragraphs (same as `git commit`).

### Dry run

```bash
git cc 'feat: something' --dry-run  # prints message, doesn't commit
```

### Pass-through flags

git-cc delegates these flags to `git commit`:

```bash
git cc 'feat: something' -a         # --all
git cc 'feat: something' -s         # --signoff
git cc 'feat: something' -n         # --no-verify
git cc 'feat: something' --no-edit  # skip editor for body
git cc 'feat: something' --author "..." --date "..."
git cc 'feat: something' --allow-empty
```

### Configuration

Config file: `commit_convention.{yaml,yml,toml}` (searched in `$PWD/`, repo root, `.config/`, `$XDG_CONFIG_HOME/`).

Generate a config:

```bash
git cc --init                  # creates .config/commit_convention.yaml
git cc --init --config-format toml
```

Config supports custom commit types, scopes (with descriptions), header max length, and enforcement:

```yaml
# commit_convention.yaml
scopes:
  - parser: parses conventional commits
  - cli: command-line UI
  - dist: release and distribution
header_max_length: 72
enforce_header_max_length: false
```

## Lightweight Commit Message Guide

These are the conventions git-cc validates against. For broader commit discipline (atomicity, splitting, branching strategy), use the `atomic-git-commits` skill.

### Format

```
type(scope)!: description

body (optional)

BREAKING CHANGE: description (optional)
```

### Type (required)

A noun describing the kind of change. Default types from the Angular convention:

| Type | Use when |
|------|----------|
| `feat` | Adding a new feature |
| `fix` | Fixing a bug |
| `docs` | Documentation only |
| `style` | Formatting, no logic change |
| `perf` | Performance improvement |
| `test` | Adding or correcting tests |
| `build` | Build system or dependencies |
| `chore` | Maintenance, tooling |
| `ci` | CI configuration |
| `refactor` | Restructuring without behavior change |
| `revert` | Reverting a prior commit |

Customize types per project via the config file.

### Scope (optional)

A noun in parentheses describing the affected code area. Example scopes: `parser`, `cli`, `api`, `auth`, `ui`. Define project scopes in the config file so git-cc validates them.

### Breaking change (`!`)

Append `!` before the colon to mark a breaking change:

```
refactor(api)!: remove deprecated /v1 endpoints
```

Optionally add a `BREAKING CHANGE:` footer with details.

### Description (required)

A short imperative summary of the change. Keep the full header (`type(scope)!: description`) ≤ 72 characters to stay within 80 columns in `git log --oneline`. git-cc defaults to 72 and can enforce this if `enforce_header_max_length` is set in config.

### Body (optional)

Free-form explanation of the change. Use `--no-edit` to skip, or let git-cc delegate to `$EDITOR` for the body after committing the header.
