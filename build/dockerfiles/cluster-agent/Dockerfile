# syntax = docker/dockerfile:1.2
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
    make ${MAKE_BUILD_CMD} MODULE_PATH=${MODULE_PATH} DOCKER_IMAGE=${DOCKER_IMAGE} GO_BUILD_OPTIONS="-tags musl"

ARG BASE_DOCKER_IMAGE
FROM ${BASE_DOCKER_IMAGE}

ENV KUBECTL_VERSION v1.19.7
ARG ARCH=amd64

RUN apk add --no-cache jq py3-pip && \
    apk add --update nodejs nodejs-npm && \
    npm i -g jackson-converter@1.0.10 && \
    pip3 install dicttoxml xmindparser && \
    apk add shadow && \
    curl -sLf https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl > /usr/bin/kubectl && \
    chmod +x /usr/bin/kubectl

WORKDIR /app

ARG APP_NAME
ARG CONFIG_PATH
ENV APP_NAME=${APP_NAME}
ENV LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib:/usr/local/lib64:/lib:/lib64

COPY --from=build "$GOPATH/src/github.com/erda-project/erda/bin/${APP_NAME}" "/app/${APP_NAME}"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/conf/${CONFIG_PATH}" "/app/conf/${CONFIG_PATH}"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/pkg/erda-configs" "/app/erda-configs"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/conf/common" "/app/conf/common"

COPY --from=build "$GOPATH/src/github.com/erda-project/erda/build/scripts/cluster-agent/*" "/usr/bin/"
RUN chmod 777 /usr/bin/kubectl-shell.sh /usr/bin/shell-setup.sh

CMD ["sh", "-c", "/app/${APP_NAME}"]
