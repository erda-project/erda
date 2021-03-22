#!/bin/bash

set -o errexit -o nounset -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

cp ../../../vendor/github.com/erda-project/erda/apistructs/sysconf.go .
sed -i 's/package apistructs/package sysconf/' sysconf.go
