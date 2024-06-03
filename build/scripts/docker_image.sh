#!/bin/bash
# Author: recallsong
# Email: songruiguo@qq.com

set -o errexit -o pipefail

echo "GO_BUILD_OPTIONS=${GO_BUILD_OPTIONS}"

# check parameters and print usage if need
usage() {
    echo "docker_build.sh MODULE [ACTION]"
    echo "MODULE: "
    echo "    module path relative to cmd/"
    echo "ACTION: "
    echo "    build       build docker image. this is default action."
    echo "    push        push docker image, and build image if image not exist."
    echo "    build-push  build and push docker image."
    echo "Environment Variables: "
    echo '    DOCKER_REGISTRY format like "docker.io/<repository>". '
    echo "    DOCKER_REGISTRY_USERNAME set username for login registry if need."
    echo "    DOCKER_REGISTRY_PASSWORD set password for login registry if need."
    exit 1
}
if [ -z "$1" ]; then
    usage
fi
MODULE_PATH=$1
ACTION=$2

# cd to root directory
cd $(git rev-parse --show-toplevel)

# image version and url
ARCH="${ARCH:-$(go env GOARCH)}"
VERSION="$(build/scripts/make-version.sh)"
IMAGE_TAG="${IMAGE_TAG:-$(build/scripts/make-version.sh tag)}"
DOCKERFILE_DEFAULT="build/dockerfiles/Dockerfile"
BASE_DOCKER_IMAGE="registry.erda.cloud/erda/erda-base:20240603"
DOCKERFILE=${DOCKERFILE_DEFAULT}

# setup single module envionment variables
setup_single_module_env() {
    MAKE_BUILD_CMD="build-one"

    # application name
    APP_NAME="$(echo ${MODULE_PATH} | sed 's/^\(.*\)[/]//')"

    # Dockerfile path and image name
    if [ -f "build/dockerfiles/Dockerfile-${APP_NAME}" ];then
        DOCKERFILE="build/dockerfiles/Dockerfile-${APP_NAME}"
    elif [ -d "build/dockerfiles/${APP_NAME}" ];then
        DOCKERFILE="build/dockerfiles/${APP_NAME}/Dockerfile"
    fi
    IMAGE_TAG="${IMAGE_TAG}-${ARCH}"
    DOCKER_IMAGE="${APP_NAME}:${IMAGE_TAG}"
}

# setup envionment variables for build all
setup_build_all_env() {
    MAKE_BUILD_CMD="build-all"
    DOCKER_IMAGE="erda:${IMAGE_TAG}"
    CONFIG_PATH=""
    MODULE_PATH=""
}

# setup build env
case "${MAKE_BUILD_CMD}" in
    "build-all")
        # build all application in one image
        setup_build_all_env
        ;;
    *)
        setup_single_module_env
        ;;
esac

if [ -n "${DOCKER_REGISTRY}" ]; then
    DOCKER_IMAGE=${DOCKER_REGISTRY}/${DOCKER_IMAGE}
fi

# print details
print_details() {
    echo "Module Path    : ${MODULE_PATH}"
    echo "App Name       : ${APP_NAME}"
    echo "Config Path    : ${CONFIG_PATH}"
    echo "Dockerfile     : ${DOCKERFILE}"
    echo "Docker Image   : ${DOCKER_IMAGE}"
    echo "Build Command  : ${MAKE_BUILD_CMD}"
    echo "Arch           : ${ARCH}"
    echo "Docker Platform: linux/${ARCH}"
}
print_details

# login
docker_login() {
    if [[ -n "${DOCKER_REGISTRY_USERNAME}" ]] && [[ -n "${DOCKER_REGISTRY}" ]] && [[ -n "${DOCKER_REGISTRY_PASSWORD}" ]]; then
        docker login -u "${DOCKER_REGISTRY_USERNAME}" -p "${DOCKER_REGISTRY_PASSWORD}" ${DOCKER_REGISTRY}
    fi
}

# build docker image
build_image()  {
    DOCKER_BUILDKIT=1 docker build --pull --platform "linux/${ARCH}" --progress=plain -t "${DOCKER_IMAGE}" \
        --label "branch=$(git rev-parse --abbrev-ref HEAD)" \
        --label "commit=$(git rev-parse HEAD)" \
        --label "build-time=$(date '+%Y-%m-%d %T%z')" \
        --build-arg "MODULE_PATH=${MODULE_PATH}" \
        --build-arg "APP_NAME=${APP_NAME}" \
        --build-arg "DOCKER_IMAGE=${DOCKER_IMAGE}" \
        --build-arg "BASE_DOCKER_IMAGE=${BASE_DOCKER_IMAGE}" \
        --build-arg "MAKE_BUILD_CMD=${MAKE_BUILD_CMD}" \
        --build-arg "GO_BUILD_OPTIONS=${GO_BUILD_OPTIONS}" \
        --build-arg "GOPROXY=${GOPROXY}" \
        --build-arg "ARCH=${ARCH}" \
        -f "${DOCKERFILE}" .
}

build_multi_arch() {
    DOCKER_IMAGE="${DOCKER_REGISTRY}/${APP_NAME}:${IMAGE_TAG}"
    DOCKER_BUILDKIT=1 docker buildx build \
      --platform linux/amd64 \
      --platform linux/arm64 \
      --label "branch=$(git rev-parse --abbrev-ref HEAD)" \
      --label "commit=$(git rev-parse HEAD)" \
      --label "build-time=$(date '+%Y-%m-%d %T%z')" \
      --build-arg "MODULE_PATH=${MODULE_PATH}" \
      --build-arg "APP_NAME=${APP_NAME}" \
      --build-arg "DOCKER_IMAGE=${DOCKER_IMAGE}" \
      --build-arg "BASE_DOCKER_IMAGE=${BASE_DOCKER_IMAGE}" \
      --build-arg "MAKE_BUILD_CMD=${MAKE_BUILD_CMD}" \
      --build-arg "GO_BUILD_OPTIONS=${GO_BUILD_OPTIONS}" \
      --build-arg "GOPROXY=${GOPROXY}" \
      -f "${DOCKERFILE}" \
      -t "${DOCKER_IMAGE}" \
      --push .
    echo "action meta: image=${DOCKER_IMAGE}"
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
    docker push "${DOCKER_IMAGE}"
}

# build and push
build_push_image() {
    build_image
    push_image
    echo "action meta: image=${DOCKER_IMAGE}"
    echo "action meta: tag=${IMAGE_TAG}"
}

case "${ACTION}" in
    "build")
        docker_login && build_image
        ;;
    "push")
        docker_login && push_image
        ;;
    "build-push")
        docker_login && build_push_image
        ;;
    "build-multi-arch")
        docker_login && build_multi_arch
        ;;
    "")
        docker_login && build_image
        ;;
    *)
        usage
esac
