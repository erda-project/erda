// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package edgepipeline_register

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/rancher/remotedialer"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/discover"
)

type Interface interface {
	//RegisterEdgeToDialer(ctx context.Context)
}

// RegisterEdgeToDialer only registers the edge to the dialer
func (p *provider) RegisterEdgeToDialer(ctx context.Context) {
	if !p.Cfg.IsEdge {
		p.Log.Warnf("current Pipeline is not deployed on edge side, skip register to center dialer")
		return
	}
	if p.Cfg.ClusterAccessKey == "" {
		p.Log.Panicf("cluster access key is empty, couldn't make edge registration")
	}
	headers, err := p.makeEdgeConnHeaders()
	if err != nil {
		p.Log.Panicf("failed to make edge connection headers: %v", err)
	}
	ep, err := p.parseDialerEndpoint()
	if err != nil {
		p.Log.Panicf("failed to parse dialer endpoint: %v", err)
	}
	for {
		// if client connect successfully, will block until client disconnect
		err = remotedialer.ClientConnect(ctx, ep, headers, nil, p.ConnectAuthorizer, nil)
		if err != nil {
			p.Log.Errorf("failed to connect to dialer: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(p.Cfg.RetryConnectDialerInterval):
			// retry connect dialer
		}
	}
}

func (p *provider) ConnectAuthorizer(proto string, address string) bool {
	switch proto {
	case "tcp":
		return true
	case "unix":
		return address == "/var/run/docker.sock"
	case "npipe":
		return address == "//./pipe/docker_engine"
	}
	return false
}

func (p *provider) makeEdgeConnDetail() (string, error) {
	edgeDetail := apistructs.ClusterDialerClientDetail{
		apistructs.ClusterDialerDataKeyPipelineAddr: p.Cfg.PipelineAddr,
		apistructs.ClusterDialerDataKeyPipelineHost: p.Cfg.PipelineHost,
	}
	edgeDetailBytes, err := edgeDetail.Marshal()
	if err != nil {
		return "", err
	}
	return string(edgeDetailBytes), nil
}

func (p *provider) makeEdgeConnHeaders() (http.Header, error) {
	edgeDetail, err := p.makeEdgeConnDetail()
	if err != nil {
		return nil, err
	}
	edgeHeaders := http.Header{
		apistructs.ClusterDialerHeaderKeyClusterKey.String():    {p.Cfg.ClusterName},
		apistructs.ClusterDialerHeaderKeyClientType.String():    {apistructs.ClusterDialerClientTypePipeline.String()},
		apistructs.ClusterDialerHeaderKeyAuthorization.String(): {p.Cfg.ClusterAccessKey},
		apistructs.ClusterDialerHeaderKeyClientDetail.String():  {edgeDetail},
	}
	return edgeHeaders, nil
}

func (p *provider) parseDialerEndpoint() (string, error) {
	u, err := url.Parse(p.Cfg.ClusterDialEndpoint)
	if err != nil {
		return "", err
	}

	//inCluster, visit dialer inner service first.
	if !p.Cfg.IsEdge && discover.ClusterDialer() != "" {
		return "ws://" + discover.ClusterDialer() + u.Path, nil
	}

	switch u.Scheme {
	case "https":
		u.Scheme = "wss"
	case "http":
		u.Scheme = "ws"
	}

	return u.String(), nil
}
