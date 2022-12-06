#!/usr/bin/env bash
set -eo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

image=registry.erda.cloud/erda/erda-base:$(date +"%Y%m%d")

docker buildx build --no-cache --pull --platform linux/amd64 -t ${image} --push . -f Dockerfile

# generate Software Bill Of Materials (SBOM)
docker sbom ${image} --output sbom.txt

echo "action meta: erda-base=$image"
