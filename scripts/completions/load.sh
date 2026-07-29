#!/bin/sh
if [ -z "${MISE_CONFIG_ROOT:-}" ] || [ -z "${MISE_SHELL}" ]; then
  echo "not loading completions: missing MISE_CONFIG_ROOT and MISE_SHELL"
fi
for f in "${MISE_CONFIG_ROOT}/scripts/completions/${MISE_SHELL}/"*; do
  # shellcheck disable=SC1090
  if [  -f "$f" ]; then . "$f"; fi
done
