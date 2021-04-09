// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package docker

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/docker/docker/api/types/versions"
	docker "github.com/docker/docker/client"
	"github.com/docker/go-connections/sockets"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	defaultHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}
)

type Client struct {
	*docker.Client
}

func NewClient(host string, tlsConfig *tls.Config) (*Client, error) {
	if host == "" {
		return nil, errors.Errorf("docker: host is empty")
	}

	proto, addr, _, err := docker.ParseHost(host)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	sockets.ConfigureTransport(transport, proto, addr)
	httpClient := &http.Client{
		Transport: transport,
	}
	c, err := docker.NewClient(host, docker.DefaultVersion, httpClient, defaultHeaders)
	if err != nil {
		return nil, err
	}

	// 若server的version比client的低，则client降低到server的version
	ping, err := c.Ping(context.Background())
	if err != nil {
		return nil, errors.Errorf("docker: fail to ping server: %s", err)
	}
	if versions.LessThan(ping.APIVersion, docker.DefaultVersion) {
		logrus.Debugf("docker: client version set to %s", ping.APIVersion)
		c.UpdateClientVersion(ping.APIVersion)
	}

	logrus.Debugf("docker: create new client")
	return &Client{c}, nil
}
