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
