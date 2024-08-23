# syntax = registry.erda.cloud/retag/dockerfile:latest
ARG BASE_DOCKER_IMAGE
FROM ${BASE_DOCKER_IMAGE} as build

RUN mkdir -p "$GOPATH/src/github.com/erda-project/erda/"
COPY . "$GOPATH/src/github.com/erda-project/erda/"
WORKDIR "$GOPATH/src/github.com/erda-project/erda/"


ARG DOCKER_IMAGE
ARG GOPROXY
RUN --mount=type=cache,target=/root/.cache/go-build\
    --mount=type=cache,target=/go/pkg/mod \
    make cli-linux DOCKER_IMAGE=${DOCKER_IMAGE} GOPROXY=${GOPROXY}

ARG BASE_DOCKER_IMAGE
FROM ${BASE_DOCKER_IMAGE}

COPY --from=build "$GOPATH/src/github.com/erda-project/erda/bin/erda-cli-linux" "/usr/bin/erda-cli"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/build/scripts/cli/run.sh" "/usr/bin/run.sh"

RUN chmod 755 /usr/bin/erda-cli
RUN chmod 755 /usr/bin/run.sh

WORKDIR /root
ENV workspace=/root
ENV home=/root

ENTRYPOINT ["/usr/bin/run.sh"]
