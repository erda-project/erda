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

	credentialpb "github.com/erda-project/erda-proto-go/core/services/authentication/credentials/accesskey/pb"
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
	registerTimeout       = 5 * time.Second
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
					logrus.Debugf("session not found, try again [%s]", clusterKey)
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

	clientType := apistructs.ClusterDialerClientType(req.Header.Get(apistructs.ClusterDialerHeaderKeyClientType.String()))
	clusterKey := req.Header.Get(apistructs.ClusterDialerHeaderKeyClusterKey.String())
	clusterKey = clientType.MakeClientKey(clusterKey)
	if clusterKey == "" {
		remotedialer.DefaultErrorWriter(rw, req, 400, errors.New("missing header:X-Erda-Cluster-Key"))
		return
	}
	switch clientType {
	case apistructs.ClusterDialerClientTypeDefault, apistructs.ClusterDialerClientTypeCluster:
		if needClusterInfo {
			// Get cluster info from agent request
			info := req.Header.Get(apistructs.ClusterDialerHeaderKeyClusterInfo.String())
			if info == "" {
				remotedialer.DefaultErrorWriter(rw, req, 400, errors.New("missing header:X-Erda-Cluster-Info"))
				return
			}

			if req.Header.Get(apistructs.ClusterDialerHeaderKeyAuthorization.String()) == "" {
				remotedialer.DefaultErrorWriter(rw, req, 400, errors.New("missing header:Authorization"))
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
			// TODO: register action after authed better.
			go registerFunc(clusterKey, clusterInfo)
		}
	default:
		clientDataStr := req.Header.Get(apistructs.ClusterDialerHeaderKeyClientDetail.String())
		if clientDataStr != "" {
			var clientData apistructs.ClusterDialerClientDetail
			if err := json.Unmarshal([]byte(clientDataStr), &clientData); err != nil {
				logrus.Errorf("failed to unmarshal client data(skip clients update), clientType: %s, clusterKey: %s, err: %v",
					clientType, clusterKey, err)
				goto register
			}
			updateClientDetail(clientType, clusterKey, clientData)
		}
		logrus.Infof("client type [%s]", clientType)
	}

register:
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
	logrus.Infof("[%d] REQ cluster=%s setting-timeout=%s %s", id, clusterKey, timeout, url)
	proxyReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		logrus.Errorf("[%d] NEW REQ cluster=%s %s: %v", id, clusterKey, url, err)
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
		logrus.Errorf("[%d] REQ ERR cluster=%s latency=%dms %s: %v", id, clusterKey, time.Since(start).Milliseconds(), url, err)
		remotedialer.DefaultErrorWriter(rw, req, 500, err)
		return
	}
	defer resp.Body.Close()
	logrus.Infof("[%d] REQ OK cluster=%s code=%d latency=%dms %s", id, clusterKey, resp.StatusCode, time.Since(start).Milliseconds(), url)
	rwHeader := rw.Header()
	for key, values := range resp.Header {
		rwHeader[textproto.CanonicalMIMEHeaderKey(key)] = values
	}
	rw.WriteHeader(resp.StatusCode)
	io.Copy(rw, resp.Body)
	logrus.Infof("[%d] REQ DONE cluster=%s latency=%dms %s", id, clusterKey, time.Since(start).Milliseconds(), url)
}

func checkClusterIsExisted(server *remotedialer.Server, rw http.ResponseWriter, req *http.Request) {
	clusterKey := req.URL.Query().Get("clusterKey")
	clientType := apistructs.ClusterDialerClientType(req.URL.Query().Get("clientType"))
	clusterKey = clientType.MakeClientKey(clusterKey)
	isExisted := server.HasSession(clusterKey)
	rw.Write([]byte(strconv.FormatBool(isExisted)))
}

func getClusterClientData(server *remotedialer.Server, rw http.ResponseWriter, req *http.Request) {
	clusterKey := mux.Vars(req)["clusterKey"]
	clientType := apistructs.ClusterDialerClientType(mux.Vars(req)["clientType"])
	clusterKey = clientType.MakeClientKey(clusterKey)
	isExisted := server.HasSession(clusterKey)
	if !isExisted {
		remotedialer.DefaultErrorWriter(rw, req, 404, errors.New("cluster not found"))
		return
	}
	clientData, ok := getClientDetail(clientType, clusterKey)
	if !ok {
		remotedialer.DefaultErrorWriter(rw, req, 404, errors.New("client data not found"))
		return
	}
	clientByteData, err := clientData.Marshal()
	if err != nil {
		remotedialer.DefaultErrorWriter(rw, req, 500, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.Write(clientByteData)
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

func Start(ctx context.Context, credential credentialpb.AccessKeyServiceServer, cfg *config.Config) error {
	authorizer := auth.New(
		auth.WithCredentialClient(credential),
		auth.WithConfig(cfg),
	)

	handler := remotedialer.New(authorizer.Authorizer, remotedialer.DefaultErrorWriter)
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
	router.HandleFunc("/clusteragent/check", func(rw http.ResponseWriter,
		req *http.Request) {
		checkClusterIsExisted(handler, rw, req)
	})
	router.HandleFunc("/clusteragent/client-detail/{clientType}/{clusterKey}", func(rw http.ResponseWriter,
		req *http.Request) {
		getClusterClientData(handler, rw, req)
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
