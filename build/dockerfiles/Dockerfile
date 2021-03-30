ARG BASE_DOCKER_IMAGE
FROM ${BASE_DOCKER_IMAGE} as build

RUN mkdir -p "$GOPATH/src/github.com/erda-project/erda/"
COPY . "$GOPATH/src/github.com/erda-project/erda/"
WORKDIR "$GOPATH/src/github.com/erda-project/erda/"

ARG APP_NAME
ARG CONFIG_PATH
ARG MODULE_PATH
ARG DOCKER_IMAGE
RUN make build APP_NAME=${APP_NAME} MODULE_PATH=${MODULE_PATH} DOCKER_IMAGE=${DOCKER_IMAGE} 

ARG BASE_DOCKER_IMAGE
FROM ${BASE_DOCKER_IMAGE}

RUN apk add --no-cache jq

WORKDIR /app

ARG APP_NAME
ARG CONFIG_PATH
ENV APP_NAME=${APP_NAME}
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/bin/${APP_NAME}" "/app/${APP_NAME}"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/conf/${CONFIG_PATH}" "/app/conf/${CONFIG_PATH}"

CMD ["sh", "-c", "/app/${APP_NAME}"]
