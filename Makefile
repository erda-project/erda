# Author: recallsong
# Email: songruiguo@qq.com

# project info
PROJ_PATH := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
BUILD_PATH ?= ${PROJ_PATH}/cmd/${MODULE_PATH}
APP_NAME ?= $(shell echo ${BUILD_PATH} | sed 's/^\(.*\)[/]//')
VERSION := $(shell head -n 1 VERSION)
# build info
GOARCH ?= $(shell go env GOARCH)
GOOS ?= $(shell go env GOOS)
GO_VERSION := $(shell go version)
GO_SHORT_VERSION := $(shell go version | awk '{print $$3}')
BUILD_TIME := $(shell date "+%Y-%m-%d %H:%M:%S")
COMMIT_ID := $(shell git rev-parse HEAD 2>/dev/null)
DOCKER_IMAGE ?=
VERSION_PKG := github.com/erda-project/erda-infra/base/version
VERSION_OPS := -ldflags "\
		-X '${VERSION_PKG}.Version=${VERSION}' \
		-X '${VERSION_PKG}.BuildTime=${BUILD_TIME}' \
        -X '${VERSION_PKG}.CommitID=${COMMIT_ID}' \
        -X '${VERSION_PKG}.GoVersion=${GO_VERSION}' \
		-X '${VERSION_PKG}.DockerImage=${DOCKER_IMAGE}'"
# build env
GOPROXY ?= https://goproxy.cn/
# GOPRIVATE ?= ""
GO_BUILD_ENV := PROJ_PATH=${PROJ_PATH} GOPROXY=${GOPROXY} GOPRIVATE=${GOPRIVATE}

.PHONY: build-version clean tidy
build: build-version summodule tidy
	cd "${BUILD_PATH}" && \
	${GO_BUILD_ENV} go build ${VERSION_OPS} -o "${PROJ_PATH}/bin/${APP_NAME}"

build-cross: build-version summodule
	cd "${BUILD_PATH}" && \
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} ${GO_BUILD_ENV} go build ${VERSION_OPS} -o "${PROJ_PATH}/bin/${GOOS}-${GOARCH}-${APP_NAME}"

build-for-linux:
	GOOS=linux GOARCH=amd64 make build-cross

build-version:
	@echo ------------ Start Build Version Details ------------
	@echo AppName: ${APP_NAME}
	@echo Arch: ${GOARCH}
	@echo OS: ${GOOS}
	@echo Version: ${VERSION}
	@echo BuildTime: ${BUILD_TIME}
	@echo GoVersion: ${GO_VERSION}
	@echo CommitID: ${COMMIT_ID}
	@echo DockerImage: ${DOCKER_IMAGE}
	@echo ------------ End   Build Version Details ------------

tidy:
	cd "${BUILD_PATH}" && \
    ${GO_BUILD_ENV} go mod tidy

generate:
	cd "${BUILD_PATH}" && \
	${GO_BUILD_ENV} go generate -v -x

summodule:
	git submodule update --init

clean:
	rm -rf "${PROJ_PATH}/bin"
	rm go.sum

run: build
	./bin/${APP_NAME}

exec:
	./bin/${APP_NAME}

# print the dependency graph of the configured providers
run-g: build
	./bin/${APP_NAME} -g

# print the compiled providers and help information
run-ps: build
	./bin/${APP_NAME} --providers