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
	"testing"

	"github.com/erda-project/erda/internal/pkg/ai-proxy/route"
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
		}
	}
}

func TestRoutes_FindRoute(t *testing.T) {
	var routes = route.Routes{
		{
			Path:    "/v1/completions",
			Method:  http.MethodPost,
			Filters: nil,
		},
	}
	findRoute, ok := routes.FindRoute("/v1/completions", "POST", make(http.Header))
	t.Log(findRoute, ok)
}
