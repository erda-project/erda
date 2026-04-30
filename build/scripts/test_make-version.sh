#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
MAKE_VERSION_SCRIPT="${ROOT_DIR}/build/scripts/make-version.sh"

run_case() {
  local input_version="$1"
  local expected="$2"

  local out
  out="$(VERSION="${input_version}" bash "${MAKE_VERSION_SCRIPT}")"
  if [[ "${out}" != "${expected}" ]]; then
    echo "FAIL: VERSION='${input_version}' -> got '${out}', expected '${expected}'" 1>&2
    exit 1
  fi
}

# Prerelease strings may contain 'v' not only at the beginning.
# normalize_version() must strip only a leading v/V prefix.
run_case "1.2-preview.1" "1.2.0-preview.1"
run_case "v1.2-preview.1" "1.2.0-preview.1"
run_case "1.2-vpreview.1" "1.2.0-vpreview.1"

echo "OK: normalize_version prerelease 'v' handling"

