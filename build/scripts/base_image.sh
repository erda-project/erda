#!/bin/bash
# Author: recallsong
# Email: songruiguo@qq.com

set -o errexit -o pipefail
cd "$(dirname "$0")/../dockerfiles/base";

# setup base image
DOCKER_IMAGE=golang-base:20220726

if [ -n "${DOCKER_REGISTRY}" ]; then
    DOCKER_IMAGE=${DOCKER_REGISTRY}/${DOCKER_IMAGE}
fi

if [ -n "${DOCKER_PLATFORM}" ]; then
    DOCKER_PLATFORM="linux/amd64"
fi

# check parameters and print usage if need
usage() {
    echo "base_image.sh ACTION"
    echo "ACTION: "
    echo "    build       build docker image."
    echo "    push        push docker image, and build image if image not exist."
    echo "    build-push  build and push docker image."
    echo "    image       show image name."
    echo "Environment Variables: "
    echo '    DOCKER_REGISTRY format like "docker.io/<repository>". '
    echo "    DOCKER_REGISTRY_USERNAME set username for login registry if need."
    echo "    DOCKER_REGISTRY_PASSWORD set password for login registry if need."
    exit 1
}
if [ -z "$1" ]; then
    usage
fi

# build docker image
build_image()  {
    DOCKER_BUILDKIT=1 docker build --pull --platform "${DOCKER_PLATFORM}" --progress=plain -t "${DOCKER_IMAGE}" \
        --label "build-time=$(date '+%Y-%m-%d %T%z')" \
        --label "alpine=3.12" \
        --label "golang=1.16" \
        --label "librdkafka=1.5.0" \
        .
}

# push docker image
push_image() {
    if [ -z "${DOCKER_REGISTRY}" ]; then
       echo "fail to push docker image, DOCKER_REGISTRY is empty !"
       exit 1
    fi
    IMAGE_ID="$(docker images ${DOCKER_IMAGE} -q)"
    if [ -z "${IMAGE_ID}" ]; then
        build_image
    fi
    if [ -n "${DOCKER_REGISTRY_USERNAME}" ]; then
        docker login -u "${DOCKER_REGISTRY_USERNAME}" -p "${DOCKER_REGISTRY_PASSWORD}" ${DOCKER_IMAGE}
    fi
    docker push "${DOCKER_IMAGE}"
}

# build and push
build_push_image() {
    build_image
    push_image
}

case "$1" in
    "build")
        build_image
        ;;
    "push")
        push_image
        ;;
    "build-push")
        build_push_image
        ;;
    "image")
        echo ${DOCKER_IMAGE}
        ;;
    *)
        usage
esac
