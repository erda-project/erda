#!/bin/bash

set -o errexit -o pipefail

# check parameters and print usage if need
usage() {
    echo "protoc.sh MODULE PB_PATH"
    echo "PB_PATH: "
    echo "    directory path that contains *.proto files."
    echo "MODULE: "
    echo "    directory path relative to modules/."
    exit 1
}

PB_PATH=$2
if [ -z "$PB_PATH" ]; then
    usage
fi
MODULE_PATH=$1

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "${ROOT_DIR}/modules/${MODULE_PATH}"

PKG_PATH=$(go run github.com/erda-project/erda-infra/tools/gopkg github.com/erda-project/erda-infra)
${PKG_PATH}/tools/protoc.sh init "$PB_PATH/*.proto"