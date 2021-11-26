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

function get_version() {
  [[ -n "${VERSION}" ]] && echo "${VERSION/v/}" && return
  [[ -f VERSION ]] && ver=$(head -n 1 VERSION) || ver=0.0
  ALPHA=${ver}-alpha
  HEAD_TAG=$(git tag --points-at HEAD |head -n1)
  # remove prefix v when present
  [[ -n "${HEAD_TAG}" ]] && echo "${HEAD_TAG/v/}" && return

  BRANCH_PREFIX=$(git rev-parse --abbrev-ref HEAD)

  if [[ "${BRANCH_PREFIX}" =~ release/[[:digit:]]+\.* ]]; then
    VERSION=${BRANCH_PREFIX//release\//}
    # some case branch already have greek version, like: release/1.0-beta2
    if ! [[ "${BRANCH_PREFIX}" =~ - ]]; then
      VERSION="${VERSION}-beta"
    fi
  fi

  echo ${VERSION:-${ALPHA}}
}

function get_tag() {
  ver=$(get_version)
  if ! [[ "${ver}" =~ - ]]; then
    ver="${ver}-stable"
  fi
  echo "${ver}-$(date '+%Y%m%d%H%M%S')-$(git rev-parse --short HEAD)"
}

case $1 in
tag)
  get_tag
  ;;
*)
  get_version
esac
