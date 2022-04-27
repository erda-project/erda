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
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rancher/remotedialer"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/clusterdialer"
	"github.com/erda-project/erda/pkg/discover"
)

type Interface interface {
	GetAccessToken(req apistructs.OAuth2TokenGetRequest) (*apistructs.OAuth2Token, error)
	GetOAuth2Token(req apistructs.OAuth2TokenGetRequest) (*apistructs.OAuth2Token, error)
	GetEdgePipelineEnvs() apistructs.ClusterDialerClientDetail
	CheckAccessToken(token string) error
	CheckAccessTokenFromHttpRequest(req *http.Request) error
	IsEdge() bool

	ShouldDispatchToEdge(source, clusterName string) bool
	GetEdgeBundleByClusterName(clusterName string) (*bundle.Bundle, error)

	// OnEdge register hook that will be invoked if you are running on edge.
	// Could register multi hooks as you need.
	// All hooks executed asynchronously.
	OnEdge(func(context.Context))

	// OnCenter register hook that will be invoked if you are running on center.
	// Could register multi hooks as you need.
	// All hooks executed asynchronously.
	OnCenter(func(context.Context))
}

func (p *provider) ShouldDispatchToEdge(source, clusterName string) bool {
	if clusterName == "" {
		return false
	}
	if p.Cfg.ClusterName == clusterName {
		return false
	}
	var findInWhitelist bool
	for _, whiteListSource := range p.Cfg.AllowedSources {
		if strings.HasPrefix(source, whiteListSource) {
			findInWhitelist = true
			break
		}
	}
	if !findInWhitelist {
		return false
	}
	isEdge, err := p.bdl.IsClusterDialerClientRegistered(clusterName, apistructs.ClusterDialerClientTypePipeline.String())
	if !isEdge || err != nil {
		return false
	}

	return true
}

func (p *provider) GetDialContextByClusterName(clusterName string) clusterdialer.DialContextFunc {
	clusterKey := apistructs.ClusterDialerClientTypePipeline.MakeClientKey(clusterName)
	return clusterdialer.DialContext(clusterKey)
}

func (p *provider) GetEdgeBundleByClusterName(clusterName string) (*bundle.Bundle, error) {
	edgeDial := p.GetDialContextByClusterName(clusterName)
	edgeDetail, err := p.bdl.GetClusterDialerClientData(apistructs.ClusterDialerClientTypePipeline.String(), clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to get edge bundle for cluster %s, err: %v", clusterName, err)
	}
	pipelineAddr := edgeDetail.Get(apistructs.ClusterDialerDataKeyPipelineAddr)
	return bundle.New(bundle.WithDialContext(edgeDial), bundle.WithCustom(discover.EnvPipeline, pipelineAddr)), nil
}

func (p *provider) GetAccessToken(req apistructs.OAuth2TokenGetRequest) (*apistructs.OAuth2Token, error) {
	return &apistructs.OAuth2Token{
		AccessToken: p.Cfg.accessToken,
		ExpiresIn:   0,
		TokenType:   "Bearer",
	}, nil
}

func (p *provider) GetOAuth2Token(req apistructs.OAuth2TokenGetRequest) (*apistructs.OAuth2Token, error) {
	return &apistructs.OAuth2Token{
		AccessToken: p.Cfg.accessToken,
		ExpiresIn:   0,
		TokenType:   "Bearer",
	}, nil
}

func (p *provider) CheckAccessTokenFromHttpRequest(req *http.Request) error {
	if p.Cfg.IsEdge {
		token := req.Header.Get("Authorization")
		return p.CheckAccessToken(token)
	}
	return nil
}

func (p *provider) CheckAccessToken(token string) error {
	if token != p.Cfg.accessToken {
		return fmt.Errorf("invalid access token")
	}
	return nil
}

func (p *provider) GetEdgePipelineEnvs() apistructs.ClusterDialerClientDetail {
	return apistructs.ClusterDialerClientDetail{
		apistructs.ClusterDialerDataKeyPipelineAddr: p.Cfg.PipelineAddr,
		apistructs.ClusterDialerDataKeyPipelineHost: p.Cfg.PipelineHost,
	}
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

func (p *provider) IsEdge() bool {
	return p.Cfg.IsEdge
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
