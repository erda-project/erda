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

package mux

import (
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
)

var (
	_ Mux = (*provider)(nil)
)

var (
	name = "gorilla-mux"
	spec = servicehub.Spec{
		Services:    []string{name},
		Summary:     "http mux by gorilla",
		Description: "http mux by gorilla",
		ConfigFunc:  func() any { return new(Config) },
		Types:       []reflect.Type{reflect.TypeOf((Mux)(nil))},
		Creator:     func() servicehub.Provider { return new(provider) },
	}
)

func init() {
	servicehub.Register(name, &spec)
}

type provider struct {
	Config *Config
	L      logs.Logger

	Router *mux.Router
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.Router = mux.NewRouter()
	return nil
}

// Start .
func (p *provider) Start() error {
	go func() {
		p.L.Infof("LitenAndServe %s", p.Config.Addr)
		if err := (&http.Server{
			Addr:              p.Config.Addr,
			Handler:           p.Router,
			ReadTimeout:       time.Second * 60,
			ReadHeaderTimeout: time.Second * 60,
		}).ListenAndServe(); err != nil {
			p.L.Fatalf("failed to LitenAndServe %s: %v", err)
		}
	}()
	return nil
}

// Close .
func (p *provider) Close() error {
	return nil
}

func (p *provider) Handle(path, method string, h http.Handler, middles ...Middle) {
	p.L.Infof("handle %s %s", path, method)
	h = Wraps(h, middles...)
	if method == "*" {
		p.Router.Path(path).Handler(h)
	} else {
		p.Router.Path(path).Methods(strings.ToUpper(method)).Handler(h)
	}
}

func (p *provider) HandlePrefix(prefix, method string, h http.Handler, middles ...Middle) {
	p.L.Infof("handle prefix %s %s", prefix, method)
	h = Wraps(h, middles...)
	if method == "*" {
		p.Router.PathPrefix(prefix).Handler(h)
	} else {
		p.Router.PathPrefix(prefix).Methods(strings.ToUpper(method)).Handler(h)
	}
}

func (p *provider) HandleMatch(match func(r *http.Request) bool, h http.Handler, middles ...Middle) {
	p.L.Infof("handle match %T", match)
	h = Wraps(h, middles...)
	p.Router.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool { return match(req) }).Handler(h)
}

func (p *provider) HandleNotFound(h http.Handler, middles ...Middle) {
	p.L.Infof("handle not found")
	p.Router.NotFoundHandler = Wraps(h, middles...)
}

func (p *provider) HandleMethodNotAllowed(h http.Handler, middles ...Middle) {
	p.L.Infof("handle method not allowed")
	p.Router.MethodNotAllowedHandler = Wraps(h, middles...)
}

type Config struct {
	Addr string `json:"addr" yaml:"addr"`
}

func Wraps(h http.Handler, middles ...Middle) http.Handler {
	for _, m := range middles {
		h = m(h)
	}
	return h
}
