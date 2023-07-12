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

package route_test

import (
	"net/http"
	"sort"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/internal/pkg/ai-proxy/route"
	"github.com/erda-project/erda/pkg/strutil"
)

func TestRoute_MatchPath(t *testing.T) {
	var cases = []struct {
		Route *route.Route
		Paths map[string]bool
	}{
		{
			Route: &route.Route{Path: "/v1/models"},
			Paths: map[string]bool{
				"/v1/models":                        true,
				"/v1/models/some-model":             false,
				"/v1/models/some-model/some-detail": false,
			},
		},
		{
			Route: &route.Route{Path: "/v1/models/{model}"},
			Paths: map[string]bool{
				"/v1/models":                        false,
				"/v1/models/some-model":             true,
				"/v1/models/some-model/some-detail": false,
			},
		},
		{
			Route: &route.Route{Path: "/openai/deployments/{deploymentId}/completions"},
			Paths: map[string]bool{
				"/openai/deployments//completions":                         false,
				"/openai/deployments/my-deployment-id/completions":         true,
				"/openai/deployments/my-deployment-id/details/completions": false,
			},
		},
	}

	for _, c := range cases {
		for p, match := range c.Paths {
			_ = c.Route.Validate()
			if ok := c.Route.Match(p, http.MethodGet, make(http.Header)); ok != match {
				t.Fatalf("match error, path: %s, path regex: %s, expect match: %v, got match: %v", p, c.Route.PathRegexExpr(), match, ok)
			}
			t.Logf("c.Route.PathRegexExpr: %s", c.Route.PathRegexExpr())
		}
	}
}

func TestRoutes_FindRoute(t *testing.T) {
	var routes = route.Routes{
		{
			Path:    "/v1/completions",
			Method:  http.MethodPost,
			Filters: nil,
			Router: &route.Router{
				To:         "openai",
				InstanceId: "default",
			},
		},
		{
			Path:   "/v1/completions",
			Method: http.MethodPost,
			HeaderMatcher: map[string]string{
				vars.XAIProxyProvider:         "azure",
				vars.XAIProxyProviderInstance: "terminus2",
			},
			Filters: nil,
			Router: &route.Router{
				To:         "azure",
				InstanceId: "terminus2",
			},
		},
	}
	sort.Sort(routes)
	t.Logf("routes: %s", strutil.TryGetYamlStr(routes))
	for _, rout := range routes {
		if err := rout.Validate(); err != nil {
			t.Log(err)
		}
	}

	t.Run("find openai default", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v1/completions", nil)
		if err != nil {
			t.Fatalf("failed to http.NewReqeust, err: %v", err)
		}
		request.Header.Set(vars.XAIProxyProvider, "openai")
		request.Header.Set(vars.XAIProxyProviderInstance, "default")
		findRoute := routes.FindRoute(request)
		t.Logf("findRoute: %s\nfindRoute.IsNotFoundRoute: %v", strutil.TryGetYamlStr(findRoute), findRoute.IsNotFoundRoute())
		if findRoute.IsNotFoundRoute() {
			t.Error("the route is not NotFoundRoute")
		}
		if findRoute.Router.To != "openai" {
			t.Fatal("find route error")
		}
	})
	t.Run("find azure terminus2", func(t *testing.T) {
		request, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v1/completions", nil)
		if err != nil {
			t.Fatalf("failed to http.NewReqeust, err: %v", err)
		}
		request.Header.Set(vars.XAIProxyProvider, "azure")
		request.Header.Set(vars.XAIProxyProviderInstance, "terminus2")
		findRoute := routes.FindRoute(request)
		t.Logf("findRoute: %s\nfindRoute.IsNotFoundRoute: %v", strutil.TryGetYamlStr(findRoute), findRoute.IsNotFoundRoute())
		if findRoute.IsNotFoundRoute() {
			t.Error("the route is not NotFoundRoute")
		}
		if findRoute.Router.To != "azure" {
			t.Fatal("find route error")
		}
	})
}

func TestRoute_Validate(t *testing.T) {
	if err := (&route.Route{
		Path:          "/",
		PathMatcher:   "",
		Method:        "",
		MethodMatcher: "",
		HeaderMatcher: nil,
		Router: &route.Router{
			To:         route.ToNotFound,
			InstanceId: "",
			Scheme:     "",
			Host:       "",
			Rewrite:    "",
		},
		Filters: nil,
	}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestRoute_RewritePath(t *testing.T) {
	var routes = route.Routes{
		{
			Path:   "/v1/models/{model}",
			Method: http.MethodGet,
			Router: &route.Router{
				To:         "azure",
				InstanceId: "default",
				Scheme:     "https",
				Host:       "default.azure.com",
				Rewrite:    "/openai/models/${ path.model }",
			},
		},
	}
	for _, r := range routes {
		if err := r.Validate(); err != nil {
			t.Log(err)
		}
	}
	request, err := http.NewRequest(http.MethodPost, "http://localhost:8080/v1/models/my-model-2023", nil)
	if err != nil {
		t.Fatalf("failed to http.NewReqeust, err: %v", err)
	}
	findRoute := routes.FindRoute(request)
	t.Log(findRoute, findRoute.IsNotFoundRoute())
	if findRoute.IsNotFoundRoute() {
		t.Fatal("the route is not NotFoundRoute")
	}
	newPath := findRoute.RewritePath(make(map[string]string))
	t.Logf("newPath: %s", newPath)
}
