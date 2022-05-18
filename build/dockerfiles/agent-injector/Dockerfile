# syntax = docker/dockerfile:1.2
ARG BASE_DOCKER_IMAGE
FROM ${BASE_DOCKER_IMAGE} as build

RUN mkdir -p "$GOPATH/src/github.com/erda-project/erda/"
COPY . "$GOPATH/src/github.com/erda-project/erda/"
WORKDIR "$GOPATH/src/github.com/erda-project/erda/"
RUN rm -fr extensions

ARG CONFIG_PATH
ARG DOCKER_IMAGE
ARG MAKE_BUILD_CMD
ARG GOPROXY
ARG MODULE_PATH
RUN --mount=type=cache,target=/root/.cache/go-build\
    --mount=type=cache,target=/go/pkg/mod \
    make ${MAKE_BUILD_CMD} MODULE_PATH=${MODULE_PATH} DOCKER_IMAGE=${DOCKER_IMAGE} GO_BUILD_OPTIONS="-tags musl" GOPROXY=${GOPROXY}

ARG BASE_DOCKER_IMAGE
FROM ${BASE_DOCKER_IMAGE} as app-handler
ARG MODULE_PATH
ENV PROJ_ROOT="/go/src/github.com/erda-project/erda"
COPY --from=build "${PROJ_ROOT}" /erda
WORKDIR /erda
# handle app
RUN ./build/scripts/build_all/app_handler.sh /erda "${MODULE_PATH}" /erda-handled

ARG BASE_DOCKER_IMAGE
FROM ${BASE_DOCKER_IMAGE}

WORKDIR /app

ARG APP_NAME
ARG CONFIG_PATH
ENV APP_NAME=${APP_NAME}
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib:/usr/local/lib64:/lib:/lib64

COPY --from=app-handler /erda-handled /erda
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/build/java-agent" "/opt/spot-original/java-agent"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/build/scripts/monitor-agent-injector/container-init.sh" "/bin/container-init.sh"
