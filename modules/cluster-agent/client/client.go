// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/remotedialer"

	"github.com/erda-project/erda/modules/cluster-agent/config"
)

var connected = make(chan struct{})

const (
	tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

func getClusterInfo(apiserverAddr string) (map[string]interface{}, error) {
	caData, err := ioutil.ReadFile(rootCAFile)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s", rootCAFile)
	}

	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s", tokenFile)
	}
	return map[string]interface{}{
		"address": apiserverAddr,
		"token":   strings.TrimSpace(string(token)),
		"caCert":  base64.StdEncoding.EncodeToString(caData),
	}, nil
}

func Start(ctx context.Context, cfg *config.Config) error {
	headers := http.Header{
		"X-Erda-Cluster-Key": {cfg.ClusterKey},
		// TODO: support encode with secretKey
		"Authorization": {cfg.ClusterKey},
	}
	if cfg.CollectClusterInfo {
		clusterInfo, err := getClusterInfo(cfg.K8SApiServerAddr)
		if err != nil {
			return err
		}
		bytes, err := json.Marshal(clusterInfo)
		if err != nil {
			return err
		}
		headers["X-Erda-Cluster-Info"] = []string{base64.StdEncoding.EncodeToString(bytes)}
	}

	u, err := url.Parse(cfg.ClusterDialEndpoint)
	if err != nil {
		return err
	}
	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	}

	for {
		remotedialer.ClientConnect(ctx, u.String(), headers, nil, func(proto, address string) bool {
			switch proto {
			case "tcp":
				return true
			case "unix":
				return address == "/var/run/docker.sock"
			case "npipe":
				return address == "//./pipe/docker_engine"
			}
			return false
		}, onConnect)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(rand.Int()%10) * time.Second):
			// retry connect after sleep a random time
		}
	}

}

func Connected() <-chan struct{} {
	return connected
}

func onConnect(context.Context, *remotedialer.Session) error {
	connected <- struct{}{}
	return nil
}
