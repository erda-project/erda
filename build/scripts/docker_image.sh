#!/bin/bash
# Author: recallsong
# Email: songruiguo@qq.com

set -o errexit -o pipefail

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
    echo "    DOCKER_REGISTRY format like \"registry.example.org/username\" ."
    echo "    DOCKER_REGISTRY_USERNAME set username for login registry if need."
    echo "    DOCKER_REGISTRY_PASSWORD set password for login registry if need."
    exit 1
}
if [ -z "$1" ]; then
    usage
fi

# cd to root directory
cd $(git rev-parse --show-toplevel)

# module path
MODULE_PATH=$1
APP_NAME="$(echo ${MODULE_PATH} | sed 's/^\(.*\)[/]//')"

# Dockerfile path
DOCKERFILE_DEFAULT="build/dockerfiles/Dockerfile"
DOCKERFILE=${DOCKERFILE_DEFAULT}
if [ -f "build/dockerfiles/Dockerfile-${APP_NAME}" ];then
    DOCKERFILE="build/dockerfiles/Dockerfile-${APP_NAME}"
elif [ -d "build/dockerfiles/${APP_NAME}" ];then
    DOCKERFILE="build/dockerfiles/${APP_NAME}/Dockerfile"
fi

# config file or directory path
if [ -f "conf/${APP_NAME}.yaml" ];then
    CONFIG_PATH="${APP_NAME}.yaml"
elif [ -f "conf/${APP_NAME}.yml" ];then
    CONFIG_PATH="${APP_NAME}.yml"
elif [ -f "conf/${MODULE_PATH}.yaml" ];then
    CONFIG_PATH="${MODULE_PATH}.yaml"
elif [ -f "conf/${MODULE_PATH}.yml" ];then
    CONFIG_PATH="${MODULE_PATH}.yml"
elif [ -d "conf/${MODULE_PATH}" ];then
    CONFIG_PATH="${MODULE_PATH}"
elif [ -d "conf/${APP_NAME}" ];then
    CONFIG_PATH="${APP_NAME}"
else
    CONFIG_PATH=""
fi

# image version and url
VERSION="$(head -n 1 VERSION)"
VERSION="${VERSION}-$(date '+%Y%m%d')-$(git rev-parse --short HEAD)"
DOCKER_IMAGE=${APP_NAME}:${VERSION}
if [ -n "${DOCKER_REGISTRY}" ]; then
    DOCKER_IMAGE=${DOCKER_REGISTRY}/${DOCKER_IMAGE}
fi
BASE_DOCKER_IMAGE="$(build/scripts/base_image.sh image)"

# print details
print_details() {
    echo "Module Path : ${MODULE_PATH}"
    echo "App Name    : ${APP_NAME}"
    echo "Config Path : ${CONFIG_PATH}"
    echo "Dockerfile  : ${DOCKERFILE}"
    echo "Docker Image: ${DOCKER_IMAGE}"
}
print_details

# build docker image
build_image()  {
    if [[ -z "${DOCKER_REGISTRY}" && ${DOCKERFILE} == ${DOCKERFILE_DEFAULT} ]]; then
        BASE_IMAGE_ID="$(docker images ${BASE_DOCKER_IMAGE} -q)"
        if [ -z "${BASE_IMAGE_ID}" ]; then
            echo "base image '${BASE_DOCKER_IMAGE}' not exist, start build base image ..."
            build/scripts/base_image.sh build
        fi
    fi
    docker build -t "${DOCKER_IMAGE}" \
        --label "branch=$(git rev-parse --abbrev-ref HEAD)" \
        --label "commit=$(git rev-parse HEAD)" \
        --label "build-time=$(date '+%Y-%m-%d %T%z')" \
        --build-arg "MODULE_PATH=${MODULE_PATH}" \
        --build-arg "APP_NAME=${APP_NAME}" \
        --build-arg "CONFIG_PATH=${CONFIG_PATH}" \
        --build-arg "DOCKER_IMAGE=${DOCKER_IMAGE}" \
        --build-arg "BASE_DOCKER_IMAGE=${BASE_DOCKER_IMAGE}" \
        -f "${DOCKERFILE}" .
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

case "$2" in
    "build")
        build_image
        ;;
    "push")
        push_image
        ;;
    "build-push")
        build_push_image
        ;;
    "")
        build_image
        ;;
    *)
        usage
esac
