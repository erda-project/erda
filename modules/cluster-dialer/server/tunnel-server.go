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
	l       sync.Mutex
	clients = map[string]*http.Client{}
	counter int64
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
	if needClusterInfo {
		clusterKey := req.Header.Get("Authorization")
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

		bdl := bundle.New(bundle.WithClusterManager())
		c, err := bdl.GetCluster(clusterKey)
		if err != nil {
			logrus.Debugf("failed to get cluster from cluster-manager: %s, err: %v", clusterKey, err)
			remotedialer.DefaultErrorWriter(rw, req, 500, err)
			return
		}
		if c.ManageConfig != nil && c.ManageConfig.Type != apistructs.ManageProxy {
			err = fmt.Errorf("cluster %s is not proxy type", clusterKey)
			logrus.Debug(err)
			remotedialer.DefaultErrorWriter(rw, req, 500, err)
			return
		}
		if err = bdl.PatchCluster(&apistructs.ClusterPatchRequest{
			Name: clusterKey,
			ManageConfig: &apistructs.ManageConfig{
				Type:    apistructs.ManageProxy,
				Address: clusterInfo.Address,
				CaData:  clusterInfo.CACert,
				Token:   clusterInfo.Token,
			},
		}, map[string][]string{httputil.InternalHeader: {"cluster-dialer"}}); err != nil {
			remotedialer.DefaultErrorWriter(rw, req, 500, err)
			return
		}
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
