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

package server

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cluster-dialer/auth"
	"github.com/erda-project/erda/modules/cluster-dialer/config"
	"github.com/erda-project/erda/pkg/http/httputil"
)

var (
	l                     sync.Mutex
	clients               = map[string]*http.Client{}
	counter               int64
	registerTimeout       = 30 * time.Second
	registerCheckInterval = 1 * time.Second
)

const (
	portalSchemeHeader  = "X-Portal-Scheme"
	portalHostHeader    = "X-Portal-Host"
	portalDestHeader    = "X-Portal-Dest"
	portalTimeoutHeader = "X-Portal-Timeout"
)

type cluster struct {
	Address string `json:"address"`
	Token   string `json:"token"`
	CACert  string `json:"caCert"`
}

func clusterRegister(server *remotedialer.Server, rw http.ResponseWriter, req *http.Request, needClusterInfo bool) {
	registerFunc := func(clusterKey string, clusterInfo cluster) {
		ctx, cancel := context.WithTimeout(context.Background(), registerTimeout)
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				logrus.Errorf("register cluster info timeout [%s]", clusterKey)
				return
			default:
				if !server.HasSession(clusterKey) {
					logrus.Infof("session not found, try again [%s]", clusterKey)
					<-time.After(registerCheckInterval)
					continue
				}
				logrus.Infof("session exsited, patch cluster info to cluster-manager [%s]", clusterKey)
				// register to cluster manager
				bdl := bundle.New(bundle.WithClusterManager())
				c, err := bdl.GetCluster(clusterKey)
				if err != nil {
					logrus.Errorf("failed to get cluster from cluster-manager: %s, err: %v", clusterKey, err)
					remotedialer.DefaultErrorWriter(rw, req, 500, err)
					return
				}

				if c.ManageConfig != nil {
					if c.ManageConfig.Type != apistructs.ManageProxy {
						logrus.Warnf("cluster is not proxy type [%s]", clusterKey)
						return
					}
					if clusterInfo.Token == c.ManageConfig.Token && clusterInfo.Address == c.ManageConfig.Address &&
						clusterInfo.CACert == c.ManageConfig.CaData {
						logrus.Infof("cluster info isn't change [%s]", clusterKey)
						return
					}
				}

				if err = bdl.PatchCluster(&apistructs.ClusterPatchRequest{
					Name: clusterKey,
					ManageConfig: &apistructs.ManageConfig{
						Type:             apistructs.ManageProxy,
						Address:          clusterInfo.Address,
						CaData:           clusterInfo.CACert,
						Token:            clusterInfo.Token,
						CredentialSource: apistructs.ManageProxy,
					},
				}, map[string][]string{httputil.InternalHeader: {"cluster-dialer"}}); err != nil {
					logrus.Errorf("failed to patch cluster [%s], err: %v", clusterKey, err)
					remotedialer.DefaultErrorWriter(rw, req, 500, err)
					return
				}

				logrus.Infof("patch cluster info success [%s]", clusterKey)

				return
			}
		}
	}

	if needClusterInfo {
		// Get cluster info from agent request
		clusterKey := req.Header.Get("X-Erda-Cluster-Key")
		if clusterKey == "" {
			remotedialer.DefaultErrorWriter(rw, req, 400, errors.New("missing header:Authorization"))
			return
		}
		info := req.Header.Get("X-Erda-Cluster-Info")
		if info == "" {
			remotedialer.DefaultErrorWriter(rw, req, 400, errors.New("missing header:X-Erda-Cluster-Info"))
			return
		}
		var clusterInfo cluster
		bytes, err := base64.StdEncoding.DecodeString(info)
		if err != nil {
			remotedialer.DefaultErrorWriter(rw, req, 400, err)
			return
		}
		if err := json.Unmarshal(bytes, &clusterInfo); err != nil {
			remotedialer.DefaultErrorWriter(rw, req, 400, err)
			return
		}
		if clusterInfo.Address == "" {
			err = errors.New("invalid cluster info, address empty")
			remotedialer.DefaultErrorWriter(rw, req, 400, err)
			return
		}
		if clusterInfo.Token == "" {
			err = errors.New("invalid cluster info, token empty")
			remotedialer.DefaultErrorWriter(rw, req, 400, err)
			return
		}
		if clusterInfo.CACert == "" {
			err = errors.New("invalid cluster info, caCert empty")
			remotedialer.DefaultErrorWriter(rw, req, 400, err)
			return
		}
		go registerFunc(clusterKey, clusterInfo)
	}

	server.ServeHTTP(rw, req)
}

func netportal(server *remotedialer.Server, rw http.ResponseWriter, req *http.Request, timeout time.Duration) {
	clusterKey := req.Header.Get(portalHostHeader)
	addrInCluster := req.Header.Get(portalDestHeader)
	schemeInCluster := req.Header.Get(portalSchemeHeader)
	timeoutStr := req.Header.Get(portalTimeoutHeader)
	if timeoutStr != "" {
		timeoutInt, err := strconv.Atoi(timeoutStr)
		if err == nil {
			timeout = time.Duration(timeoutInt) * time.Second
		}
	}
	if schemeInCluster == "" {
		schemeInCluster = "http"
	}
	url := fmt.Sprintf("%s://%s%s", schemeInCluster, addrInCluster, req.URL.EscapedPath())
	if req.URL.RawQuery != "" {
		url = fmt.Sprintf("%s?%s", url, req.URL.RawQuery)
	}
	client := getClusterClient(server, clusterKey, timeout)
	id := atomic.AddInt64(&counter, 1)
	logrus.Infof("[%d] REQ timeout=%s %s", id, timeout, url)
	proxyReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		logrus.Errorf("[%d] NEW REQ %s: %v", id, url, err)
		remotedialer.DefaultErrorWriter(rw, req, 500, err)
		return
	}
	req.Header.Del(portalHostHeader)
	req.Header.Del(portalDestHeader)
	req.Header.Del(portalSchemeHeader)
	req.Header.Del(portalTimeoutHeader)
	proxyReq.Header = req.Header
	start := time.Now()
	resp, err := client.Do(proxyReq)
	if err != nil {
		logrus.Errorf("[%d] REQ ERR latency=%dms %s: %v", id, time.Since(start).Milliseconds(), url, err)
		remotedialer.DefaultErrorWriter(rw, req, 500, err)
		return
	}
	defer resp.Body.Close()
	logrus.Infof("[%d] REQ OK code=%d latency=%dms %s", id, resp.StatusCode, time.Since(start).Milliseconds(), url)
	rwHeader := rw.Header()
	for key, values := range resp.Header {
		rwHeader[textproto.CanonicalMIMEHeaderKey(key)] = values
	}
	rw.WriteHeader(resp.StatusCode)
	io.Copy(rw, resp.Body)
	logrus.Infof("[%d] REQ DONE latency=%dms %s", id, time.Since(start).Milliseconds(), url)
}

func getClusterClient(server *remotedialer.Server, clusterKey string, timeout time.Duration) *http.Client {
	l.Lock()
	defer l.Unlock()

	key := fmt.Sprintf("%s/%s", clusterKey, timeout)
	client := clients[key]
	if client != nil {
		return client
	}

	dialer := server.Dialer(clusterKey)
	client = &http.Client{
		Transport: &http.Transport{
			DialContext: dialer,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: timeout,
	}

	clients[key] = client
	return client
}

func Start(ctx context.Context, cfg *config.Config) error {
	handler := remotedialer.New(auth.Authorizer, remotedialer.DefaultErrorWriter)
	handler.ClientConnectAuthorizer = func(proto, address string) bool {
		if strings.HasSuffix(proto, "::tcp") {
			return true
		}
		if strings.HasSuffix(proto, "::unix") {
			return address == "/var/run/docker.sock"
		}
		if strings.HasSuffix(proto, "::npipe") {
			return address == "//./pipe/docker_engine"
		}
		return false
	}
	// TODO: support handler.AddPeer
	router := mux.NewRouter()
	router.Handle("/clusterdialer", handler)
	router.HandleFunc("/clusteragent/connect", func(rw http.ResponseWriter,
		req *http.Request) {
		clusterRegister(handler, rw, req, cfg.NeedClusterInfo)
	})
	router.PathPrefix("/").HandlerFunc(func(rw http.ResponseWriter,
		req *http.Request) {
		netportal(handler, rw, req, cfg.Timeout)
	})
	server := &http.Server{
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
		Addr:    cfg.Listen,
		Handler: router,
	}
	return server.ListenAndServe()
}
