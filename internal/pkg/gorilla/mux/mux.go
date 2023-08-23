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
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

var (
	_ Mux = (*Server)(nil)
)

type Mux interface {
	LitenAndServe(addr string)
	Handle(path, method string, h http.Handler, middles ...Middle)
	HandlePrefix(prefix, method string, h http.Handler, middles ...Middle)
	HandleMatch(match func(r *http.Request) bool, h http.Handler, middles ...Middle)
}

type Server struct {
	Router *mux.Router
}

func New() *Server {
	return &Server{
		Router: mux.NewRouter(),
	}
}

func (s *Server) LitenAndServe(addr string) {
	go func() {
		if err := http.ListenAndServe(addr, s.Router); err != nil {
			panic(fmt.Sprintf("failed to LitenAndServe %s: %v", addr, err))
		}
	}()
}

func (s *Server) Handle(path, method string, h http.Handler, middles ...Middle) {
	h = Wraps(h, middles...)
	if method == "*" {
		s.Router.Path(path).Handler(h)
	} else {
		s.Router.Path(path).Methods(strings.ToUpper(method)).Handler(h)
	}
}

func (s *Server) HandlePrefix(prefix, method string, h http.Handler, middles ...Middle) {
	h = Wraps(h, middles...)
	if method == "*" {
		s.Router.PathPrefix(prefix).Handler(h)
	} else {
		s.Router.PathPrefix(prefix).Methods(strings.ToUpper(method)).Handler(h)
	}
}

func (s *Server) HandleMatch(match func(r *http.Request) bool, h http.Handler, middles ...Middle) {
	h = Wraps(h, middles...)
	s.Router.MatcherFunc(func(req *http.Request, _ *mux.RouteMatch) bool { return match(req) }).Handler(h)
}

func Wraps(h http.Handler, middles ...Middle) http.Handler {
	for _, m := range middles {
		h = m(h)
	}
	return h
}
