#!/bin/bash

# This script generates semver version by the following rules in order of priority from top to bottom
# 1. the environment variable `VERSION`
# 2. take tag name when the HEAD matches any tag
# 3. take x.x when the HEAD matches branch named as release/x.x
#    cases:
#     a) release/1.0 -> 1.0-beta
#     b) release/1.0-beta2 -> 1.0-beta2
# 4. VERSION file content which indicates the next version

set -o errexit

function normalize_version() {
  local raw="${1/v/}"
  if [[ "${raw}" =~ ^([0-9]+)\.([0-9]+)$ ]]; then
    echo "${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.0"
    return
  fi
  if [[ "${raw}" =~ ^([0-9]+)\.([0-9]+)-(.+)$ ]]; then
    echo "${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.0-${BASH_REMATCH[3]}"
    return
  fi
  echo "${raw}"
}

function get_version() {
  [[ -n "${VERSION}" ]] && normalize_version "${VERSION}" && return
  [[ -f VERSION ]] && ver=$(head -n 1 VERSION) || ver=0.0
  BASE=$(normalize_version "${ver}")
  TS=$(date '+%Y%m%d%H%M%S')
  HEAD_TAG=$(git tag --points-at HEAD |head -n1)
  # remove prefix v when present
  [[ -n "${HEAD_TAG}" ]] && normalize_version "${HEAD_TAG}" && return

  BRANCH_PREFIX=$(git rev-parse --abbrev-ref HEAD)

  if [[ "${BRANCH_PREFIX}" =~ ^release/([0-9]+\.[0-9]+)(-beta([0-9]+))?$ ]]; then
    BASE=$(normalize_version "${BASH_REMATCH[1]}")
    if [[ -n "${BASH_REMATCH[3]}" ]]; then
      echo "${BASE}-beta.${BASH_REMATCH[3]}.${TS}"
    else
      echo "${BASE}-beta.${TS}"
    fi
    return
  fi

  echo "${BASE}-alpha.${TS}"
}

function get_tag() {
  ver=$(get_version)
  if ! [[ "${ver}" =~ - ]]; then
    ver="${ver}-stable"
  fi
  echo "${ver}-$(date '+%Y%m%d%H%M%S')-$(git rev-parse --short HEAD)"
}

function get_channel() {
  ver=$(get_version)
  if [[ "${ver}" =~ -alpha ]]; then
    echo "alpha"
  elif [[ "${ver}" =~ -beta ]]; then
    echo "beta"
  else
    echo "stable"
  fi
}

case $1 in
tag)
  get_tag
  ;;
channel)
  get_channel
  ;;
*)
  get_version
esac
