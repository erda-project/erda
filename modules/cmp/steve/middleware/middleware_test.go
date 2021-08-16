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
