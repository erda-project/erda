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

package cluster_agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
)

type config struct {
	Debug               bool   `default:"false" desc:"enable debug logging"`
	ClusterDialEndpoint string `desc:"cluster dialer endpoint"`
	ClusterKey          string `desc:"cluster key"`
	SecretKey           string `desc:"secret key"`
	K8SApiServerAddr    string `desc:"kube-apiserver address in cluster"`
}

const (
	tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
)

type provider struct {
	Cfg *config // auto inject this field
}

func (p *provider) Init(ctx servicehub.Context) error {
	if p.Cfg.Debug {
		logrus.SetLevel(logrus.DebugLevel)
		remotedialer.PrintTunnelData = true
	}
	return nil
}

func (p *provider) getClusterInfo() (map[string]interface{}, error) {
	caData, err := ioutil.ReadFile(rootCAFile)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s", rootCAFile)
	}

	token, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, errors.Wrapf(err, "reading %s", tokenFile)
	}
	return map[string]interface{}{
		"address": p.Cfg.K8SApiServerAddr,
		"token":   strings.TrimSpace(string(token)),
		"caCert":  base64.StdEncoding.EncodeToString(caData),
	}, nil
}

func (p *provider) Run(ctx context.Context) error {
	clusterInfo, err := p.getClusterInfo()
	if err != nil {
		return err
	}
	bytes, err := json.Marshal(clusterInfo)
	if err != nil {
		return err
	}
	headers := http.Header{
		"X-Erda-Cluster-Info": {base64.StdEncoding.EncodeToString(bytes)},
		"X-Erda-Cluster-Key":  []string{p.Cfg.ClusterKey},
		// TODO: support encode with secretKey
		"Authorization": []string{p.Cfg.ClusterKey},
	}
	for {
		remotedialer.ClientConnect(ctx, p.Cfg.ClusterDialEndpoint, headers, nil, func(proto, address string) bool {
			switch proto {
			case "tcp":
				return true
			case "unix":
				return address == "/var/run/docker.sock"
			case "npipe":
				return address == "//./pipe/docker_engine"
			}
			return false
		}, nil)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Duration(rand.Int()%10) * time.Second):
			// retry connect after sleep a random time
		}
	}
}

func init() {
	servicehub.Register("cluster-agent", &servicehub.Spec{
		Services:     []string{"cluster-agent"},
		Dependencies: []string{"http-server"},
		Description:  "cluster agent",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
