#!/usr/bin/env bash
set -euo pipefail
shells=(
  bash
  zsh
)
generate_completions() {
  cli_name="$1"
  cmd="$*"
  for sh in "${shells[@]}"; do
    eval "SHELL=$sh $cmd $sh" > "./scripts/completions/$sh/$cli_name"
  done
}
for sh in "${shells[@]}"; do
  d=./scripts/completions/$sh
  f=$d/.gitignore
  if [ ! -f "$f" ]; then
    mkdir -p "$d"
    printf '*\n!.gitignore\n' > "$f"
  fi
done

generate_completions mise completion
generate_completions git-cc completion
generate_completions golangci-lint completion
