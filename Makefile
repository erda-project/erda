# Copyright (c) 2021 Terminus, Inc.
#
# This program is free software: you can use, redistribute, and/or modify
# it under the terms of the GNU Affero General Public License, version 3
# or later ("AGPL"), as published by the Free Software Foundation.
#
# This program is distributed in the hope that it will be useful, but WITHOUT
# ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
# FITNESS FOR A PARTICULAR PURPOSE.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.

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
# GOPROXY ?= https://goproxy.cn/
# GOPRIVATE ?= ""
# GO_BUILD_ENV := PROJ_PATH=${PROJ_PATH} GOPROXY=${GOPROXY} GOPRIVATE=${GOPRIVATE}
GO_BUILD_ENV := PROJ_PATH=${PROJ_PATH} GOPRIVATE=${GOPRIVATE}

.PHONY: build-version clean tidy
build-all: build-version submodule tidy
	@set -o errexit; \
	MODULES=$$(find "./cmd" -maxdepth 10 -type d); \
	for path in $${MODULES}; \
	do \
		HAS_GO_FILE=$$(eval echo $$(bash -c "find "$${path}" -maxdepth 1 -name *.go 2>/dev/null" | wc -l)); \
		if [ $${HAS_GO_FILE} -gt 0 ]; then \
			MODULE_PATH=$${path#cmd/}; \
			echo "gona to build module: $$MODULE_PATH"; \
			MODULE_PATHS="$${MODULE_PATHS} $${path}"; \
		fi; \
	done; \
	mkdir -p "${PROJ_PATH}/bin" && \
	${GO_BUILD_ENV} go build ${VERSION_OPS} ${GO_BUILD_OPTIONS} -o "${PROJ_PATH}/bin" $${MODULE_PATHS}; \
	make cli; \
	echo "build all modules successfully!"

build: build-version submodule tidy
	cd "${BUILD_PATH}" && \
	${GO_BUILD_ENV} go build ${VERSION_OPS} ${GO_BUILD_OPTIONS} -o "${PROJ_PATH}/bin/${APP_NAME}"
	@ echo "build the ${MODULE_PATH} module successfully!"

build-cross: build-version submodule
	cd "${BUILD_PATH}" && \
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} ${GO_BUILD_ENV} go build ${VERSION_OPS} ${GO_BUILD_OPTIONS} -o "${PROJ_PATH}/bin/${GOOS}-${GOARCH}-${APP_NAME}"

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
	@if [[ -f "${BUILD_PATH}/go.mod" ]]; then \
		echo "go mod tidy: use module-level go.mod" && \
		cd "${BUILD_PATH}" && ${GO_BUILD_ENV} go mod tidy; \
	elif [[ -d "${PROJ_PATH}/vendor" ]]; then \
		echo "go mod tidy: already have vendor dir, skip tidy" ; \
	else \
		echo "go mod tidy: use project-level go.mod" && \
		cd "${PROJ_PATH}" && ${GO_BUILD_ENV} go mod tidy; \
	fi

generate:
	cd "${BUILD_PATH}" && \
	${GO_BUILD_ENV} go generate -v -x

prepare:
	cd "${PROJ_PATH}" && \
	${GO_BUILD_ENV} go generate ./apistructs && \
	${GO_BUILD_ENV} go generate ./modules/openapi/api/generate && \
	${GO_BUILD_ENV} go generate ./modules/openapi/component-protocol/generate
	make prepare-cli

submodule:
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

# normalize all go files before push to git repo
normalize:
	@go mod tidy
	@echo "run gofmt && goimports && golint ..."
	@if [ -z "$$MODULE_PATH" ]; then \
		MODULE_PATH=.; \
	fi; \
	cd $${MODULE_PATH}; \
	go test -test.timeout=10s ./...; \
	GOFILES=$$(find . -name "*.go"); \
	for path in $${GOFILES}; do \
	 	gofmt -w -l $${path}; \
	  	goimports -w -l $${path}; \
	  	golint -set_exit_status=1 $${path}; \
	done;

# check copyright header
check-copyright:
	go run tools/gotools/go-copyright/main.go

# check go imports order
check-imports:
	go run tools/gotools/go-imports-order/main.go

# docker image
build-image: prepare
	./build/scripts/docker_image.sh ${MODULE_PATH} build
push-image:
	./build/scripts/docker_image.sh ${MODULE_PATH} push
build-push-image: prepare
	./build/scripts/docker_image.sh ${MODULE_PATH} build-push

build-push-all:
	MAKE_BUILD_CMD=build-all ./build/scripts/docker_image.sh / build-push
build-push-base-image:
	./build/scripts/base_image.sh build-push

# build cli
prepare-cli:
	cd tools/cli/command/generate && go generate
.PHONY: cli
cli:
	cd tools/cli && \
	${GO_BUILD_ENV} go build ${VERSION_OPS} ${GO_BUILD_OPTIONS} -o "${PROJ_PATH}/bin/erda-cli"
	echo "build cli tool successfully!"
.PHONY: cli-linux
cli-linux:
	cd tools/cli && \
	GOOS=linux GOARCH=amd64	${GO_BUILD_ENV} go build ${VERSION_OPS} ${GO_BUILD_OPTIONS} -o "${PROJ_PATH}/bin/erda-cli-linux"
	echo "build cli tool successfully!"

.PHONY: upload-cli
upload-cli: cli cli-linux
	go run build/scripts/upload_cli/main.go ${ACCESS_KEY_ID} ${ACCESS_KEY_SECRET} cli/mac/erda "${PROJ_PATH}/bin/erda-cli"
	go run build/scripts/upload_cli/main.go ${ACCESS_KEY_ID} ${ACCESS_KEY_SECRET} cli/linux/erda "${PROJ_PATH}/bin/erda-cli-linux"
