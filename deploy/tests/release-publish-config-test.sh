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

GO_MODULE_TEST_DIR="$(mktemp -d)"
trap 'rm -rf "$GO_MODULE_TEST_DIR"' EXIT

awk '
  /^  release:/ { in_release_job=1 }
  in_release_job && /- name: Verify release version/ { exit }
  in_release_job && /- name: Download Go modules/ { in_step=1; next }
  in_step && /- name:/ { exit }
  in_step && /run: \|/ { in_run=1; next }
  in_run { sub(/^          /, ""); print }
' "$WORKFLOW" > "$GO_MODULE_TEST_DIR/download.sh"

if [[ ! -s "$GO_MODULE_TEST_DIR/download.sh" ]]; then
  echo "Release must prepare Go modules before building and publishing" >&2
  exit 1
fi

cat > "$GO_MODULE_TEST_DIR/go" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
[[ "$*" == "mod download" ]] || exit 64
attempt="$(cat "$GO_MODULE_TEST_DIR/attempts")"
attempt=$((attempt + 1))
printf '%s\n' "$attempt" > "$GO_MODULE_TEST_DIR/attempts"
if [[ "$attempt" -le "$GO_MODULE_TEST_FAILURES" ]]; then
  echo 'stream error: INTERNAL_ERROR; received from peer' >&2
  exit 1
fi
SH
cat > "$GO_MODULE_TEST_DIR/sleep" <<'SH'
#!/usr/bin/env bash
exit 0
SH
chmod +x "$GO_MODULE_TEST_DIR/go" "$GO_MODULE_TEST_DIR/sleep"

check_go_download() {
  local failures="$1" expected_status="$2" expected_attempts="$3" status=0
  printf '0\n' > "$GO_MODULE_TEST_DIR/attempts"
  PATH="$GO_MODULE_TEST_DIR:$PATH" \
    GO_MODULE_TEST_DIR="$GO_MODULE_TEST_DIR" GO_MODULE_TEST_FAILURES="$failures" \
    bash -e -o pipefail "$GO_MODULE_TEST_DIR/download.sh" \
    > "$GO_MODULE_TEST_DIR/output" 2>&1 || status=$?
  if [[ "$status" -ne "$expected_status" \
        || "$(cat "$GO_MODULE_TEST_DIR/attempts")" -ne "$expected_attempts" ]]; then
    cat "$GO_MODULE_TEST_DIR/output" >&2
    echo "Unexpected Go download result after $failures simulated failures" >&2
    exit 1
  fi
}

check_go_download 0 0 1
check_go_download 2 0 3
check_go_download 3 1 3

echo "release publish config tests passed"
