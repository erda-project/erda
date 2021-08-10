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

// TODO: refactor this module when pkg/httpserver ready
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/eventbox/conf"
	"github.com/erda-project/erda/modules/eventbox/server/types"
)

type Server struct {
	eps    []types.Endpoint
	router *mux.Router
	srv    *http.Server
}

func New() (*Server, error) {
	return &Server{
		eps:    []types.Endpoint{},
		router: mux.NewRouter(),
	}, nil
}

func (s *Server) AddEndPoints(eps []types.Endpoint) {
	for _, ep := range eps {
		logrus.Infof("Server register endpoint: %s", ep.Path)
	}
	s.eps = append(s.eps, eps...)
}

func (s *Server) Start() error {
	s.initEndpoints()
	srv := &http.Server{
		Addr:         conf.ListenAddr(),
		Handler:      s.router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	s.srv = srv
	logrus.Infof("start listen addr: %s", conf.ListenAddr())
	err := srv.ListenAndServe()
	if err != nil {
		if err != http.ErrServerClosed {
			logrus.Errorf("failed to listen and serve: %v", err)
			return err
		}
		return nil
	}
	return nil
}

func (s *Server) Stop() error {
	s.srv.Shutdown(context.Background())
	return nil
}

func (s *Server) Router() *mux.Router {
	return s.router
}

func (s *Server) initEndpoints() {
	endpoints := s.eps

	for _, ep := range endpoints {
		s.router.PathPrefix("/api/dice/eventbox").Path(ep.Path).Methods(ep.Method).HandlerFunc(s.internal(ep.Handler))
	}
}

func (s *Server) internal(handler func(context.Context, *http.Request, map[string]string) (types.Responser, error)) http.HandlerFunc {
	pctx := context.Background()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(pctx)
		defer cancel()
		defer r.Body.Close()

		start := time.Now()
		logrus.Infof("start %s %s", r.Method, r.URL.String())
		defer func() {
			logrus.Infof("end %s %s (took %v)", r.Method, r.URL.String(), time.Since(start))
		}()

		response, err := handler(ctx, r, mux.Vars(r))
		if err != nil {
			logrus.Errorf("failed to handle request: %s (%v)", r.URL.String(), err)

			if response != nil {
				w.WriteHeader(response.GetStatus())
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			io.WriteString(w, err.Error())
			return
		}
		if response == nil || response.GetContent() == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if response.Raw() {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(response.GetStatus())
			w.Write([]byte(fmt.Sprintf("%v", response.GetContent())))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(response.GetStatus())
		if err := json.NewEncoder(w).Encode(response.GetContent()); err != nil {
			logrus.Errorf("failed to send response: %s (%v)", r.URL.String(), err)
			return
		}
	}
}
