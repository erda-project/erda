# Copyright (c) 2021 Terminus, Inc.
#
# This program is free software: you can use, redistribute, and/or modify
# it under the terms of the GNU Affero General Public License, version 3
# or later ("AGPL"), as published by the Free Software Foundation.
#
# This program is distributed in the hope that it will be useful, but WITHOUT
# ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
# FITNESS FOR A PARTICULAR PURPOSE.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program. If not, see <http://www.gnu.org/licenses/>.

FROM golang:1.15 AS builder

RUN echo $PWD

COPY . /go/src/github.com/erda-project/erda
WORKDIR /go/src/github.com/erda-project/erda

# go build
RUN GOOS=linux GOARCH=amd64 go build -o /opt/action/run cmd/migrator/internal/cmd/main.go

RUN chmod 777 /opt/action/run

FROM centos/mysql-57-centos7

MAINTAINER chenzhongrun "zhongrun.czr@alibaba-inc.com"

USER root

ENV MYSQL_ROOT_PASSWORD="12345678"
ENV MYSQL_ALLOW_EMPTY_PASSWORD yes

COPY --from=builder /opt/action/run /opt/action/run
COPY --from=builder /go/src/github.com/erda-project/erda/cmd/migrator/my.cnf /etc/my.cnf

RUN chmod 777 /opt/action/*

RUN mkdir -p /opt/action/migrations

VOLUME ["/opt/action/migrations"]

ENTRYPOINT ["/opt/action/run"]
