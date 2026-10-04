#!/usr/bin/env bash
# Prints the CHANGELOG.md section for a version (without its heading), so
# the GitHub release carries the curated notes rather than the commit log.
# Exits non-zero when the version has no section or the section is empty.
set -euo pipefail

version="${1:?usage: release-notes.sh vX.Y.Z}"
changelog="${2:-CHANGELOG.md}"

notes=$(awk -v heading="## ${version}" '
  $0 == heading { found = 1; next }
  found && /^## / { exit }
  found { print }
' "$changelog")

if [ -z "$(printf '%s' "$notes" | tr -d '[:space:]')" ]; then
  echo "error: ${changelog} has no notes under '## ${version}'" >&2
  exit 1
fi

# Trim leading and trailing blank lines.
printf '%s\n' "$notes" | sed -e '/./,$!d' | sed -e ':a' -e '/^\n*$/{$d;N;ba' -e '}'
