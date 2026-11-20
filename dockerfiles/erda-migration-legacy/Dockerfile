# Builds the v2.4 legacy migration image by reusing the binary that the
# standard `make build-all` pipeline already produces in the unified image.
# The base image keeps the MySQL sandbox and client that the migration needs;
# only /usr/bin/erda-migrate is replaced.
#
# Usage:
#   docker build \
#     --build-arg ERDA_IMAGE=registry.erda.cloud/erda/erda:<tag> \
#     -t registry.erda.cloud/erda/erda-migration-legacy:<tag> .

ARG ERDA_IMAGE
FROM --platform=$TARGETPLATFORM ${ERDA_IMAGE} AS erda-binaries

FROM --platform=$TARGETPLATFORM registry.erda.cloud/erda/erda-migratoin:57-20260504230057-9dafe6754dc644c375514a9a0818b650e9791ef0

COPY --from=erda-binaries /erda/cmd/dbmigration-legacy/bin /usr/bin/erda-migrate

RUN chmod 0755 /usr/bin/erda-migrate

ENTRYPOINT ["/usr/bin/erda-migrate"]
