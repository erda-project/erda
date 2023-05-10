# syntax = docker/dockerfile:1.2
ARG BASE_DOCKER_IMAGE
FROM --platform=$TARGETPLATFORM ${BASE_DOCKER_IMAGE} as build

RUN mkdir -p "$GOPATH/src/github.com/erda-project/erda/"
COPY . "$GOPATH/src/github.com/erda-project/erda/"
WORKDIR "$GOPATH/src/github.com/erda-project/erda/"

ARG MODULE_PATH
ARG DOCKER_IMAGE
ARG MAKE_BUILD_CMD
ARG GOPROXY
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    make ${MAKE_BUILD_CMD} MODULE_PATH=${MODULE_PATH} DOCKER_IMAGE=${DOCKER_IMAGE} GOPROXY=${GOPROXY}


FROM --platform=$TARGETPLATFORM ${BASE_DOCKER_IMAGE} as app-handler
ARG MODULE_PATH
ENV PROJ_ROOT="/go/src/github.com/erda-project/erda"

COPY --from=build "${PROJ_ROOT}" /erda
WORKDIR /erda
# handle app
RUN ./build/scripts/build_all/app_handler.sh /erda "${MODULE_PATH}" /erda-handled


FROM --platform=$TARGETPLATFORM ${BASE_DOCKER_IMAGE}
ARG MODULE_PATH
ARG TARGETARCH
ENV MODULE_PATH=${MODULE_PATH}
ENV KUBECTL_VERSION=v1.24.6

COPY --from=app-handler /erda-handled /erda
COPY --from=app-handler /erda/build/scripts/cluster-agent/* /usr/bin/

RUN curl -sLf https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl > /usr/bin/kubectl \
    && chmod +x /usr/bin/kubectl \
    && chmod 777 /usr/bin/kubectl-shell.sh /usr/bin/shell-setup.sh

WORKDIR /erda
CMD ["bash", "-c", "/erda/cmd/${MODULE_PATH}/bin"]
