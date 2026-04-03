#!/usr/bin/env bash

set -o errexit -o pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/git_context.sh"

cd "$(git rev-parse --show-toplevel)"
REPO_ROOT="$(pwd)"
CONTAINER_REPO_PATH="/go/src/github.com/erda-project/erda"
BUILD_CONTEXT="${REPO_ROOT}"
TEMP_BUILD_CONTEXT=""

cleanup() {
    if [[ -n "${TEMP_BUILD_CONTEXT}" && -d "${TEMP_BUILD_CONTEXT}" ]]; then
        rm -rf "${TEMP_BUILD_CONTEXT}"
    fi
}
trap cleanup EXIT

prepare_build_context() {
    if ! is_git_worktree "${REPO_ROOT}"; then
        return
    fi

    TEMP_BUILD_CONTEXT="$(mktemp -d)"
    prepare_git_build_context "${REPO_ROOT}" "${TEMP_BUILD_CONTEXT}" "${CONTAINER_REPO_PATH}"
    BUILD_CONTEXT="${TEMP_BUILD_CONTEXT}"
}

BASE_DOCKER_IMAGE=registry.erda.cloud/erda/erda-base:20250812
IMAGE_TAG="${IMAGE_TAG:-$(build/scripts/make-version.sh tag)}"
DOCKERFILE=./build/dockerfiles

if [[ "$MAKE_BUILD_CMD" == build-one ]]; then
    APP_NAME="$(echo "$MODULE_PATH" | awk -F/ '{print $NF}')"
    DOCKER_IMAGE="${APP_NAME}:${IMAGE_TAG}"
    if [[ -d "$DOCKERFILE/$APP_NAME" ]]; then
        DOCKERFILE="$DOCKERFILE/$APP_NAME"
    fi
elif [[ "$MAKE_BUILD_CMD" == build-all ]]; then
    DOCKER_IMAGE="erda:$IMAGE_TAG"
    CONFIG_PATH=""
    MODULE_PATH=""
else
    echo "invalid MAKE_BUILD_CMD: $MAKE_BUILD_CMD"
    exit 1
fi

if [[ -n "$DOCKER_REGISTRY" ]]; then
    DOCKER_IMAGE="$DOCKER_REGISTRY/$DOCKER_IMAGE"
    if [[ -n "$DOCKER_REGISTRY_USERNAME" && -n "$DOCKER_REGISTRY_PASSWORD" ]]; then
        docker login -u "$DOCKER_REGISTRY_USERNAME" -p "$DOCKER_REGISTRY_PASSWORD" "$DOCKER_REGISTRY"
    fi
fi

prepare_build_context

buildctl \
    --addr tcp://buildkitd.default.svc.cluster.local:1234 \
    --tlscacert=/.buildkit/ca.pem \
    --tlscert=/.buildkit/cert.pem \
    --tlskey=/.buildkit/key.pem \
     build \
    --frontend dockerfile.v0 \
    --local context="$BUILD_CONTEXT" \
    --local dockerfile="$BUILD_CONTEXT/$DOCKERFILE" \
    --opt label:"branch=$(git rev-parse --abbrev-ref HEAD)" \
    --opt label:"commit=$(git rev-parse HEAD)" \
    --opt label:"build-time=$(date '+%Y-%m-%d %T%z')" \
    --opt platform="$PLATFORMS" \
    --opt build-arg:"MODULE_PATH=$MODULE_PATH" \
    --opt build-arg:"APP_NAME=$APP_NAME" \
    --opt build-arg:"BASE_DOCKER_IMAGE=$BASE_DOCKER_IMAGE" \
    --opt build-arg:"MAKE_BUILD_CMD=$MAKE_BUILD_CMD" \
    --opt build-arg:"GO_BUILD_OPTIONS=$GO_BUILD_OPTIONS" \
    --opt build-arg:"GOPROXY=$GOPROXY" \
    --opt build-arg:"DOCKER_IMAGE=$DOCKER_IMAGE" \
    --output type=image,name=$DOCKER_IMAGE,push=true

echo "action meta: image=$DOCKER_IMAGE"
echo "action meta: tag=$IMAGE_TAG"
