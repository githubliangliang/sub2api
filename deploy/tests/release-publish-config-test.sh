#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORKFLOW="$ROOT_DIR/.github/workflows/release.yml"
CONFIG="$ROOT_DIR/.goreleaser.yaml"

if ! awk '
  /- name: Run GoReleaser/ { in_step=1; next }
  in_step && /- name:/ { exit }
  in_step && /GORELEASER_CURRENT_TAG:[[:space:]]*\$\{\{ env\.RELEASE_TAG \}\}/ { found=1 }
  END { exit found ? 0 : 1 }
' "$WORKFLOW"; then
  echo "GoReleaser must pin GORELEASER_CURRENT_TAG to RELEASE_TAG" >&2
  exit 1
fi

if ! awk '
  /^release:/ { in_release=1; next }
  in_release && /^[^[:space:]]/ { exit }
  in_release && /^[[:space:]]+replace_existing_artifacts:[[:space:]]+true([[:space:]]|$)/ { found=1 }
  END { exit found ? 0 : 1 }
' "$CONFIG"; then
  echo "GoReleaser must replace existing release assets on retry" >&2
  exit 1
fi

echo "release publish config tests passed"
