#!/usr/bin/env bash
set -eo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

# create docker buildx builder
docker buildx create --name erda-base-builder --platform linux/amd64,linux/arm64 --use || true

images=()

archs=(amd64 arm64)
for arch in ${archs[@]}; do
    image=registry.erda.cloud/erda/${arch}/erda-base:$(date +"%Y%m%d")
    images+=(${image})
    docker buildx build --platform linux/${arch} -t ${image} -f Dockerfile --push .
    # generate Software Bill Of Materials (SBOM)
    docker sbom ${image} --output sbom-${arch}.txt
done

for image in ${images[@]}; do
    echo "action meta: erda-base=$image"
done
