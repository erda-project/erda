#!/usr/bin/env bash
set -eo pipefail

image=registry.erda.cloud/erda/erda-base:$(date +"%Y%m%d")

docker buildx build --platform linux/amd64 -t ${image} --push . -f Dockerfile

echo "action meta: erda-base=$image"
