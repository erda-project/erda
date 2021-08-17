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

package steve

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rancher/apiserver/pkg/urlbuilder"
	"github.com/rancher/steve/pkg/server/router"
)

// RoutesWrapper wraps a router for steve server with different prefix
func RoutesWrapper(prefix string) router.RouterFunc {
	return func(h router.Handlers) http.Handler {
		m := mux.NewRouter()
		m.UseEncodedPath()
		m.StrictSlash(true)
		m.Use(urlbuilder.RedirectRewrite)

		s := m.PathPrefix(prefix).Subrouter()
		s.Path("/").Handler(h.APIRoot).HeadersRegexp("Accepts", ".*json.*")
		s.Path("/{name:v1}").Handler(h.APIRoot)

		s.Path("/v1/{type}").Handler(h.K8sResource)
		s.Path("/v1/{type}/{nameorns}").Queries("link", "{link}").Handler(h.K8sResource)
		s.Path("/v1/{type}/{nameorns}").Queries("action", "{action}").Handler(h.K8sResource)
		s.Path("/v1/{type}/{nameorns}").Handler(h.K8sResource)
		s.Path("/v1/{type}/{namespace}/{name}").Queries("action", "{action}").Handler(h.K8sResource)
		s.Path("/v1/{type}/{namespace}/{name}").Queries("link", "{link}").Handler(h.K8sResource)
		s.Path("/v1/{type}/{namespace}/{name}").Handler(h.K8sResource)
		s.Path("/v1/{type}/{namespace}/{name}/{link}").Handler(h.K8sResource)
		s.Path("/api").Handler(h.K8sProxy)
		s.PathPrefix("/api/").Handler(h.K8sProxy)
		s.PathPrefix("/apis").Handler(h.K8sProxy)
		s.PathPrefix("/openapi").Handler(h.K8sProxy)
		s.PathPrefix("/version").Handler(h.K8sProxy)
		s.NotFoundHandler = h.Next
		return m
	}
}
