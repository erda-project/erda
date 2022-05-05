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
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rancher/remotedialer"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/clusterdialer"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/k8sclient"
)

type Interface interface {
	GetAccessToken(req apistructs.OAuth2TokenGetRequest) (*apistructs.OAuth2Token, error)
	GetOAuth2Token(req apistructs.OAuth2TokenGetRequest) (*apistructs.OAuth2Token, error)
	GetEdgePipelineEnvs() apistructs.ClusterDialerClientDetail
	CheckAccessToken(token string) error
	CheckAccessTokenFromHttpRequest(req *http.Request) error
	IsEdge() bool

	CanProxyToEdge(source apistructs.PipelineSource, clusterName string) bool
	GetEdgeBundleByClusterName(clusterName string) (*bundle.Bundle, error)
	ClusterIsEdge(clusterName string) (bool, error)

	// OnEdge register hook that will be invoked if you are running on edge.
	// Could register multi hooks as you need.
	// All hooks executed asynchronously.
	OnEdge(func(context.Context))

	// OnCenter register hook that will be invoked if you are running on center.
	// Could register multi hooks as you need.
	// All hooks executed asynchronously.
	OnCenter(func(context.Context))
}

func (p *provider) ClusterIsEdge(clusterName string) (bool, error) {
	isEdge, err := p.bdl.IsClusterDialerClientRegistered(apistructs.ClusterDialerClientTypePipeline, clusterName)
	if err != nil {
		return false, err
	}
	return isEdge, nil
}

func (p *provider) CanProxyToEdge(source apistructs.PipelineSource, clusterName string) bool {
	if clusterName == "" {
		return false
	}
	if p.Cfg.ClusterName == clusterName {
		return false
	}
	var findInWhitelist bool
	for _, whiteListSource := range p.Cfg.AllowedSources {
		if strings.HasPrefix(source.String(), whiteListSource) {
			findInWhitelist = true
			break
		}
	}
	if !findInWhitelist {
		return false
	}
	isEdge, err := p.bdl.IsClusterDialerClientRegistered(apistructs.ClusterDialerClientTypePipeline, clusterName)
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
	edgeDetail, err := p.bdl.GetClusterDialerClientData(apistructs.ClusterDialerClientTypePipeline, clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to get edge bundle for cluster %s, err: %v", clusterName, err)
	}
	pipelineAddr := edgeDetail.Get(apistructs.ClusterDialerDataKeyPipelineAddr)
	return bundle.New(bundle.WithDialContext(edgeDial), bundle.WithCustom(discover.EnvPipeline, pipelineAddr)), nil
}

func (p *provider) GetAccessToken(req apistructs.OAuth2TokenGetRequest) (*apistructs.OAuth2Token, error) {
	return &apistructs.OAuth2Token{
		AccessToken: p.EdgeTaskAccessToken(),
		ExpiresIn:   0,
		TokenType:   "Bearer",
	}, nil
}

func (p *provider) GetOAuth2Token(req apistructs.OAuth2TokenGetRequest) (*apistructs.OAuth2Token, error) {
	return &apistructs.OAuth2Token{
		AccessToken: p.EdgeTaskAccessToken(),
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
	if token != p.EdgeTaskAccessToken() {
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

// RegisterEdgeToDialer registers the edge to the dialer
// watch the edge cluster key and update the edge cluster key
func (p *provider) RegisterEdgeToDialer(ctx context.Context) {
	if !p.Cfg.IsEdge {
		p.Log.Warnf("current Pipeline is not deployed on edge side, skip register to center dialer")
		return
	}

	ep, err := p.parseDialerEndpoint()
	if err != nil {
		p.Log.Fatalf("failed to parse dialer endpoint: %v", err)
	}
	for {
		if p.ClusterAccessKey() == "" {
			time.Sleep(time.Second)
			continue
		}
		headers, err := p.makeEdgeConnHeaders()
		if err != nil {
			p.Log.Fatalf("failed to make edge connection headers: %v", err)
		}
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
		apistructs.ClusterDialerHeaderKeyAuthorization.String(): {p.ClusterAccessKey()},
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

// waitingEdgeReady block the current process until the edge side pipeline is ready
func (p *provider) waitingEdgeReady(ctx context.Context) {
	if !p.Cfg.IsEdge {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(1 * time.Second):
			if p.ClusterAccessKey() != "" && p.EdgeTaskAccessToken() != "" {
				p.Log.Infof("edge pipeline is ready")
				return
			}
			p.Log.Warnf("waiting for edge ready...")
		}
	}
}

func (p *provider) watchClusterCredential(ctx context.Context) {
	// if specified cluster access key, preferred to use it.
	if p.ClusterAccessKey() != "" {
		p.setAccessTokenIfNotExist(p.ClusterAccessKey())
		return
	}

	var (
		retryWatcher *watchtools.RetryWatcher
		err          error
	)

	// Wait cluster credential secret ready.
	for {
		retryWatcher, err = p.getInClusterRetryWatcher(p.Cfg.ErdaNamespace)
		if err != nil {
			p.Log.Errorf("get retry warcher, %v", err)
		} else if retryWatcher != nil {
			break
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(rand.Int()%10) * time.Second):
			p.Log.Warnf("failed to get retry watcher, try again")
		}
	}

	defer retryWatcher.Stop()

	p.Log.Info("start retry watcher")

	for {
		select {
		case event := <-retryWatcher.ResultChan():
			sec, ok := event.Object.(*corev1.Secret)
			if !ok {
				p.Log.Errorf("illegal secret object, igonre")
				continue
			}

			p.Log.Debugf("event type: %v, object: %+v", event.Type, sec)

			switch event.Type {
			case watch.Deleted:
				// ignore delete event, if cluster offline, reconnect will be failed.
				continue
			case watch.Added, watch.Modified:
				ak, ok := sec.Data[apistructs.ClusterAccessKey]
				// If accidentally deleted credential key, use the latest access key
				if !ok {
					p.Log.Errorf("cluster info doesn't contain cluster access key value")
					continue
				}

				// Access key values doesn't change, skip reconnect
				if string(ak) == p.ClusterAccessKey() {
					p.Log.Debug("cluster access key doesn't change, skip")
					continue
				}

				if p.ClusterAccessKey() == "" {
					p.Log.Infof("get cluster access key: %s", string(ak))
				} else {
					p.Log.Infof("cluster access key change from %s to %s", p.ClusterAccessKey(), string(ak))
				}

				// change value
				p.setAccessKey(string(ak))
				p.setAccessTokenIfNotExist(string(ak))
			}
		case <-ctx.Done():
			return
		}
	}
}

func (p *provider) getInClusterRetryWatcher(ns string) (*watchtools.RetryWatcher, error) {
	cs, err := k8sclient.New(p.Cfg.ClusterName, k8sclient.WithPreferredToUseInClusterConfig())
	if err != nil {
		return nil, fmt.Errorf("create clientset error: %v", err)
	}

	// Get or create secret
	secInit, err := cs.ClientSet.CoreV1().Secrets(ns).Get(context.Background(), apistructs.ErdaClusterCredential, v1.GetOptions{})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("get secret error: %v", err)
		}
		// try to create init cluster credential secret
		secInit, err = cs.ClientSet.CoreV1().Secrets(ns).Create(context.Background(), &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{Name: apistructs.ErdaClusterCredential},
			Data:       map[string][]byte{apistructs.ClusterAccessKey: []byte("init")},
		}, v1.CreateOptions{})

		if err != nil {
			return nil, fmt.Errorf("create init cluster credential secret error: %v", err)
		}
	}

	// create retry watcher
	retryWatcher, err := watchtools.NewRetryWatcher(secInit.ResourceVersion, &cache.ListWatch{
		WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
			return cs.ClientSet.CoreV1().Secrets(ns).Watch(context.Background(), v1.ListOptions{
				FieldSelector: fmt.Sprintf("metadata.name=%s", apistructs.ErdaClusterCredential),
			})
		},
	})

	if err != nil {
		return nil, fmt.Errorf("create retry watcher error: %v", err)
	}

	return retryWatcher, nil
}

func (p *provider) ClusterAccessKey() string {
	p.Lock()
	ac := p.Cfg.ClusterAccessKey
	p.Unlock()
	return ac
}

func (p *provider) setAccessKey(ac string) {
	p.Lock()
	defer p.Unlock()
	p.Cfg.ClusterAccessKey = ac
}

func (p *provider) EdgeTaskAccessToken() string {
	p.Lock()
	token := p.Cfg.AccessToken
	p.Unlock()
	return token
}

func (p *provider) setAccessTokenIfNotExist(token string) {
	p.Lock()
	if p.Cfg.AccessToken == "" {
		p.Cfg.AccessToken = token
	}
	p.Unlock()
}
