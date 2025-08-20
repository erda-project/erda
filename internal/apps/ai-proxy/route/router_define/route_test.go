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

package router_define

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

func TestExpandRoute(t *testing.T) {
	tests := []struct {
		name           string
		route          *Route
		expectedRoutes int
		expectedError  string
	}{
		{
			name: "single path and method",
			route: &Route{
				Path:   "/api/v1/test",
				Method: "GET",
			},
			expectedRoutes: 1,
		},
		{
			name: "multiple paths single method",
			route: &Route{
				Path:   "/api/v1/test,/api/v2/test",
				Method: "GET",
			},
			expectedRoutes: 2,
		},
		{
			name: "single path multiple methods",
			route: &Route{
				Path:   "/api/v1/test",
				Method: "GET,POST",
			},
			expectedRoutes: 2,
		},
		{
			name: "multiple paths and methods",
			route: &Route{
				Path:   "/api/v1/test,/api/v2/test",
				Method: "GET,POST",
			},
			expectedRoutes: 4,
		},
		{
			name: "empty path segments ignored",
			route: &Route{
				Path:   "/api/v1/test,,/api/v2/test",
				Method: "GET",
			},
			expectedRoutes: 2,
		},
		{
			name: "empty method segments ignored",
			route: &Route{
				Path:   "/api/v1/test",
				Method: "GET,,POST",
			},
			expectedRoutes: 2,
		},
		{
			name: "wildcard method not allowed",
			route: &Route{
				Path:   "/api/v1/test",
				Method: "*",
			},
			expectedError: "method cannot be *",
		},
		{
			name: "wildcard in multiple methods not allowed",
			route: &Route{
				Path:   "/api/v1/test",
				Method: "GET,*,POST",
			},
			expectedError: "method cannot be *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expandedRoutes, err := ExpandRoute(tt.route)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, expandedRoutes)
				return
			}

			require.NoError(t, err)
			assert.Len(t, expandedRoutes, tt.expectedRoutes)

			// verify each expanded route has single path and method
			for _, route := range expandedRoutes {
				assert.NotContains(t, route.Path, ",", "route path should not contain comma")
				assert.NotContains(t, route.Method, ",", "route method should not contain comma")
			}
		})
	}
}

func TestExpandRoute_preservesRouteProperties(t *testing.T) {
	originalRoute := &Route{
		Path:   "/api/v1/test,/api/v2/test",
		Method: "GET,POST",
		RequestFilters: []filter_define.FilterConfig{
			{Name: "auth"},
		},
		ResponseFilters: []filter_define.FilterConfig{
			{Name: "cors"},
		},
	}

	expandedRoutes, err := ExpandRoute(originalRoute)
	require.NoError(t, err)
	assert.Len(t, expandedRoutes, 4) // 2 paths × 2 methods = 4 routes

	// verify all routes preserve the original filters
	for _, route := range expandedRoutes {
		assert.Len(t, route.RequestFilters, 1)
		assert.Equal(t, "auth", route.RequestFilters[0].Name)
		assert.Len(t, route.ResponseFilters, 1)
		assert.Equal(t, "cors", route.ResponseFilters[0].Name)
	}
}

func TestExpandRoute_pathMethodCombinations(t *testing.T) {
	expectedCombinations := []struct {
		path   string
		method string
	}{
		{"/api/v1/test", "GET"},
		{"/api/v1/test", "POST"},
		{"/api/v2/test", "GET"},
		{"/api/v2/test", "POST"},
	}

	route := &Route{
		Path:   "/api/v1/test,/api/v2/test",
		Method: "GET,POST",
	}

	expandedRoutes, err := ExpandRoute(route)
	require.NoError(t, err)
	assert.Len(t, expandedRoutes, 4)

	// verify we have all expected combinations
	actualCombinations := make(map[string]bool)
	for _, route := range expandedRoutes {
		key := route.Path + ":" + route.Method
		actualCombinations[key] = true
	}

	for _, expected := range expectedCombinations {
		key := expected.path + ":" + expected.method
		assert.True(t, actualCombinations[key], "missing combination: %s", key)
	}
}
