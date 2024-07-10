#!/usr/bin/env bash

set -o errexit -o pipefail

# goto current bash script dir
cd "$(dirname "$0")"

rm -fr bin

ARCHES="amd64 arm64"

for arch in $ARCHES; do
    GOOS=darwin GOARCH=$arch go build -o bin/action-runner-$arch
done
