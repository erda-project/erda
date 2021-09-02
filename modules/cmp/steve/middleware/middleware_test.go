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

package middleware

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

var (
	opts option
	log  *logrus.Logger
)

type option struct {
	port int
}

type Server struct {
	//auth   *Authenticator
	rw          http.ResponseWriter
	r           *http.Response
	chains      Chain
	audit       http.Handler
	router      *mux.Router
	httpHandler http.Handler
}

func handler() http.HandlerFunc {
	return func(http.ResponseWriter, *http.Request) {}
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	res := s.chains.Handler(s.httpHandler)
	res.ServeHTTP(rw, r)
	frame := []byte("123213213")
	_, err := rw.Write(frame)
	if err != nil {
		logrus.Error(err)
		return
	}
}
func Init() {
	log = logrus.New()
	p := rand.Intn(32767) + 1024
	opts = option{port: p}
}

func TestChain(t *testing.T) {
	Init()
	bundleOpts := []bundle.Option{
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*60),
			)),
		bundle.WithPipeline(),
		bundle.WithScheduler(),
		bundle.WithMonitor(),
		bundle.WithCoreServices(),
		bundle.WithOrchestrator(),
		bundle.WithDiceHub(),
		bundle.WithEventBox(),
		bundle.WithClusterManager(),
	}
	bdl := bundle.New(bundleOpts...)

	server := opts
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", server.port))
	if err != nil {
		t.Error(err)
		return
	}
	auth := NewAuthenticator(bdl)
	audit := NewAuditor(bdl)
	shell := NewShellHandler(context.Background())
	chain := Chain{
		audit.AuditMiddleWare,
		shell.HandleShell,
		auth.AuthMiddleware,
	}
	s := Server{
		router:      mux.NewRouter(),
		chains:      chain,
		httpHandler: handler(),
	}

	//s.auth = auth
	go func() {
		err := http.Serve(l, &s)
		if err != nil {
			t.Error(err)
		}
	}()

	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	buf := bytes.NewBuffer(make([]byte, 1024))
	hc.Post(fmt.Sprintf("127.0.0.1:%d/api/k8s/clusters/local/v1/nodes", server.port)).Header("Org-ID", "1").Header("User-ID", "2").Do().Body(buf)
	hc.Post(fmt.Sprintf("127.0.0.1:%d/api/k8s/clusters/local/kubectl-shell", server.port)).Header("Org-ID", "1").Header("User-ID", "2").Do().Body(buf)
}

func TestAudit(t *testing.T) {
	Init()
	bundleOpts := []bundle.Option{
		bundle.WithHTTPClient(
			httpclient.New(
				httpclient.WithTimeout(time.Second, time.Second*60),
			)),
		bundle.WithPipeline(),
		bundle.WithScheduler(),
		bundle.WithMonitor(),
		bundle.WithCoreServices(),
		bundle.WithOrchestrator(),
		bundle.WithDiceHub(),
		bundle.WithEventBox(),
		bundle.WithClusterManager(),
	}
	bdl := bundle.New(bundleOpts...)

	server := opts
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", server.port))
	if err != nil {
		t.Error(err)
		return
	}
	audit := NewAuditor(bdl)
	chain := Chain{
		audit.AuditMiddleWare,
	}
	s := Server{
		router:      mux.NewRouter(),
		chains:      chain,
		httpHandler: handler(),
	}

	//s.auth = auth
	go func() {
		err := http.Serve(l, &s)
		if err != nil {
			t.Error(err)
		}
	}()

	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	buf := bytes.NewBuffer(make([]byte, 1024))
	hc.Post(fmt.Sprintf("127.0.0.1:%d/api/k8s/clusters/local/kubectl-shell", server.port)).Header("Org-ID", "1").Header("User-ID", "2").Do().Body(buf)
}
