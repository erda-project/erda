# syntax = registry.erda.cloud/retag/dockerfile:latest
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
    make ${MAKE_BUILD_CMD} MODULE_PATH=${MODULE_PATH} DOCKER_IMAGE=${DOCKER_IMAGE} GOPROXY=${GOPROXY}



FROM ${BASE_DOCKER_IMAGE} as app-handler
ARG MODULE_PATH
ENV PROJ_ROOT="/go/src/github.com/erda-project/erda"
COPY --from=build "${PROJ_ROOT}" /erda
WORKDIR /erda
# handle app
RUN ./build/scripts/build_all/app_handler.sh /erda "${MODULE_PATH}" /erda-handled



FROM ${BASE_DOCKER_IMAGE}

ARG MODULE_PATH

COPY --from=app-handler /erda-handled /erda
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/build/java-agent" "/opt/spot-original/java-agent"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/cmd/${MODULE_PATH}/conf/container-init.sh" "/bin/container-init.sh"

WORKDIR /erda
CMD ["sh", "-c", "/erda/cmd/${MODULE_PATH}/bin"]
