FROM registry.cn-hangzhou.aliyuncs.com/terminus/golang-librdkafka:alpine3.10-golang1.13.5-librdkafka1.3.0-tz as build

RUN mkdir -p "$GOPATH/src/github.com/erda-project/erda/"
COPY . "$GOPATH/src/github.com/erda-project/erda/"
WORKDIR "$GOPATH/src/github.com/erda-project/erda/"

RUN echo "Asia/Shanghai" > /etc/timezone

ARG APP_NAME
ARG CONFIG_PATH
ARG MODULE_PATH
ARG DOCKER_IMAGE
RUN make build APP_NAME=${APP_NAME} MODULE_PATH=${MODULE_PATH} DOCKER_IMAGE=${DOCKER_IMAGE} 

FROM registry.cn-hangzhou.aliyuncs.com/terminus/golang-librdkafka:alpine3.10-golang1.13.5-librdkafka1.3.0-tz

WORKDIR /app

ARG APP_NAME
ARG CONFIG_PATH
ENV APP_NAME=${APP_NAME}
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/bin/${APP_NAME}" "/app/${APP_NAME}"
COPY --from=build "$GOPATH/src/github.com/erda-project/erda/conf/${CONFIG_PATH}" "/app/conf/${CONFIG_PATH}"

CMD ["sh", "-c", "/app/${APP_NAME}"]
