# Builds the v2.4 legacy migration image.
#
# The runtime image keeps the MySQL sandbox and client required by the
# migration, but its glibc is too old for a dynamically linked binary.
# The runner is therefore compiled here with CGO disabled.
#
# Usage:
#   docker build --platform linux/amd64 \
#     -t registry.erda.cloud/erda/erda-migration-legacy:<tag> \
#     -f dockerfiles/erda-migration-legacy/Dockerfile .

ARG GO_IMAGE=golang:1.24
ARG GOPROXY=https://goproxy.cn,direct

FROM --platform=$TARGETPLATFORM ${GO_IMAGE} AS builder

ARG GOPROXY
ENV GOPROXY=${GOPROXY}
ENV CGO_ENABLED=0

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o /out/dbmigration-legacy ./cmd/dbmigration-legacy

FROM --platform=$TARGETPLATFORM registry.erda.cloud/erda/erda-migratoin:57-20260504230057-9dafe6754dc644c375514a9a0818b650e9791ef0

COPY --from=builder /out/dbmigration-legacy /usr/bin/erda-migrate

RUN chmod 0755 /usr/bin/erda-migrate \
    && python3 -m pip install --no-cache-dir \
        Django==3.2.25 \
        mysql-connector-python==8.0.26 \
        pytz==2021.1 \
        requests==2.27.1 \
        sqlparse==0.4.4

ENTRYPOINT ["/usr/bin/erda-migrate"]
