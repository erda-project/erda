#!/usr/bin/env bash
set -eo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

# create docker buildx builder
docker buildx create --name erda-base-builder --platform linux/amd64,linux/arm64 --use || true

image=registry.erda.cloud/erda/erda-base:$(date +"%Y%m%d")

docker buildx build --platform linux/amd64,linux/arm64 -t ${image} -f Dockerfile --push .

archs=(amd64 arm64)
for arch in ${archs[@]}; do
    # generate Software Bill Of Materials (SBOM)
    docker scout sbom ${image} --format list --platform linux/${arch} --output sbom-${arch}.txt
done

echo "action meta: erda-base=$image"

docker buildx rm erda-base-builder
