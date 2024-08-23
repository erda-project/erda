# syntax = registry.erda.cloud/retag/dockerfile:latest
ARG BASE_DOCKER_IMAGE
FROM ${BASE_DOCKER_IMAGE} as build

RUN mkdir -p "$GOPATH/src/github.com/erda-project/erda/"
COPY . "$GOPATH/src/github.com/erda-project/erda/"
WORKDIR "$GOPATH/src/github.com/erda-project/erda/"

ARG CONFIG_PATH
ARG MODULE_PATH
ARG DOCKER_IMAGE
ARG MAKE_BUILD_CMD
RUN --mount=type=cache,target=/root/.cache/go-build\
    --mount=type=cache,target=/go/pkg/mod \
    make ${MAKE_BUILD_CMD} MODULE_PATH=${MODULE_PATH} DOCKER_IMAGE=${DOCKER_IMAGE}

ARG BASE_DOCKER_IMAGE
FROM ${BASE_DOCKER_IMAGE}

WORKDIR /app

ARG APP_NAME
ARG CONFIG_PATH
ENV APP_NAME=${APP_NAME}

COPY --from=build "$GOPATH/src/github.com/erda-project/erda/bin/${APP_NAME}" "/app/${APP_NAME}"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/conf/${CONFIG_PATH}" "/app/conf/${CONFIG_PATH}"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/pkg/erda-configs" "/app/erda-configs"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/conf/common" "/app/conf/common"

RUN cp -r /app /opt/action && mv /opt/action/${APP_NAME} /opt/action/run

CMD ["sh", "-c", "/app/${APP_NAME}"]
