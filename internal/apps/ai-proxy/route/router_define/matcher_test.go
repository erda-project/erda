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
	"strings"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define/path_matcher"
)

// TestRoute is a test implementation of RouteMatcher
type TestRoute struct {
	path        string
	method      string
	name        string
	pathMatcher *path_matcher.PathMatcher
}

// NewTestRoute creates a new test route
func NewTestRoute(path, method, name string) *TestRoute {
	return &TestRoute{
		path:        path,
		method:      method,
		name:        name,
		pathMatcher: path_matcher.NewPathMatcher(path),
	}
}

// MethodMatcher implements RouteMatcher interface
func (tr *TestRoute) MethodMatcher(method string) bool {
	return strings.EqualFold(tr.method, method) || tr.method == "*"
}

// PathMatcher implements RouteMatcher interface
func (tr *TestRoute) PathMatcher() PathMatcher {
	return tr.pathMatcher
}

// GetPath implements RouteMatcher interface
func (tr *TestRoute) GetPath() string {
	return tr.path
}

// GetMethod implements RouteMatcher interface
func (tr *TestRoute) GetMethod() string {
	return tr.method
}

// GetName returns the test route name
func (tr *TestRoute) GetName() string {
	return tr.name
}

func TestCalculateRoutePriority(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		method   string
		expected int
	}{
		{
			name:     "root path should have low priority",
			path:     "/",
			method:   "GET",
			expected: -45, // -50(root) + 0(no valid segments) + 0(length) + 5(method)
		},
		{
			name:     "exact path should have high priority",
			path:     "/api/v1/models",
			method:   "GET",
			expected: 41, // 3*10(exact) + 3*2(length) + 5(method)
		},
		{
			name:     "path with one parameter",
			path:     "/api/files/{file_id}",
			method:   "GET",
			expected: 32, // 2*10(exact) + 1*1(param) + 3*2(length) + 5(method)
		},
		{
			name:     "path with multiple parameters",
			path:     "/api/{version}/files/{file_id}",
			method:   "GET",
			expected: 35, // 1*10(exact) + 3*1(param) + 4*2(length) + 5(method)
		},
		{
			name:     "wildcard path should have lower priority",
			path:     "/api/*",
			method:   "GET",
			expected: 9, // 1*10(exact) + 2*2(length) + 5(method) - 20(wildcard)
		},
		{
			name:     "any method should have lower priority",
			path:     "/api/files",
			method:   "*",
			expected: 24, // 2*10(exact) + 2*2(length) + 0(any method)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := NewTestRoute(tt.path, tt.method, tt.name)
			priority := CalculateRoutePriority(route)
			if priority != tt.expected {
				t.Errorf("CalculateRoutePriority() = %v, want %v", priority, tt.expected)
			}
		})
	}
}

func TestRouter_FindBestMatch(t *testing.T) {
	tests := []struct {
		name          string
		routes        []struct{ path, method, name string }
		requestMethod string
		requestPath   string
		expectedRoute string
		shouldMatch   bool
	}{
		{
			name: "exact match should win over parameter match",
			routes: []struct{ path, method, name string }{
				{"/api/files", "GET", "exact"},
				{"/api/{resource}", "GET", "param"},
				{"/", "*", "root"},
			},
			requestMethod: "GET",
			requestPath:   "/api/files",
			expectedRoute: "exact",
			shouldMatch:   true,
		},
		{
			name: "more specific parameter match should win",
			routes: []struct{ path, method, name string }{
				{"/api/files/{file_id}", "GET", "specific"},
				{"/api/{resource}/{id}", "GET", "general"},
				{"/{path}", "*", "catch_all"},
			},
			requestMethod: "GET",
			requestPath:   "/api/files/123",
			expectedRoute: "specific",
			shouldMatch:   true,
		},
		{
			name: "method-specific route should win over any method",
			routes: []struct{ path, method, name string }{
				{"/api/files", "*", "any_method"},
				{"/api/files", "GET", "get_method"},
			},
			requestMethod: "GET",
			requestPath:   "/api/files",
			expectedRoute: "get_method",
			shouldMatch:   true,
		},
		{
			name: "longer path should win over shorter path",
			routes: []struct{ path, method, name string }{
				{"/api", "GET", "short"},
				{"/api/v1/files", "GET", "long"},
			},
			requestMethod: "GET",
			requestPath:   "/api/v1/files",
			expectedRoute: "long",
			shouldMatch:   true,
		},
		{
			name: "no matching route",
			routes: []struct{ path, method, name string }{
				{"/api/files", "POST", "post_only"},
				{"/different/path", "GET", "different"},
			},
			requestMethod: "GET",
			requestPath:   "/api/files",
			expectedRoute: "",
			shouldMatch:   false,
		},
		{
			name: "root path should be last resort",
			routes: []struct{ path, method, name string }{
				{"/", "*", "root"},
				{"/api/{resource}", "GET", "api_resource"},
			},
			requestMethod: "GET",
			requestPath:   "/api/files",
			expectedRoute: "api_resource",
			shouldMatch:   true,
		},
		{
			name: "complex scenario with mixed routes",
			routes: []struct{ path, method, name string }{
				{"/", "*", "root"},
				{"/{path}", "*", "catch_all"},
				{"/api/*", "GET", "api_wildcard"},
				{"/api/{version}", "GET", "api_version"},
				{"/api/v1/{resource}", "GET", "api_v1_resource"},
				{"/api/v1/files", "GET", "api_v1_files"},
				{"/api/v1/files/{id}", "GET", "api_v1_files_id"},
			},
			requestMethod: "GET",
			requestPath:   "/api/v1/files/123",
			expectedRoute: "api_v1_files_id",
			shouldMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()

			// Add routes
			for _, route := range tt.routes {
				router.AddRoute(NewTestRoute(route.path, route.method, route.name))
			}

			// Find match
			matched := router.FindBestMatch(tt.requestMethod, tt.requestPath)

			if !tt.shouldMatch {
				if matched != nil {
					t.Errorf("Expected no match, but got route: %s", matched.(*TestRoute).GetName())
				}
				return
			}

			if matched == nil {
				t.Errorf("Expected to match route %s, but got no match", tt.expectedRoute)
				return
			}

			matchedName := matched.(*TestRoute).GetName()
			if matchedName != tt.expectedRoute {
				t.Errorf("Expected route %s, but got %s", tt.expectedRoute, matchedName)
			}
		})
	}
}

func TestRouter_AddRoute(t *testing.T) {
	router := NewRouter()

	route1 := NewTestRoute("/api/files", "GET", "route1")
	route2 := NewTestRoute("/api/users", "POST", "route2")

	router.AddRoute(route1)
	router.AddRoute(route2)

	routes := router.GetRoutes()
	if len(routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(routes))
	}
}

func TestRouter_Clear(t *testing.T) {
	router := NewRouter()

	router.AddRoute(NewTestRoute("/test", "GET", "test"))
	router.Clear()

	routes := router.GetRoutes()
	if len(routes) != 0 {
		t.Errorf("Expected 0 routes after clear, got %d", len(routes))
	}
}

func BenchmarkRouter_FindBestMatch(b *testing.B) {
	router := NewRouter()

	// Add some test routes
	routes := []struct{ path, method, name string }{
		{"/", "*", "root"},
		{"/{path}", "*", "catch_all"},
		{"/api/v1/models", "GET", "models"},
		{"/api/v1/files", "GET", "files"},
		{"/api/v1/files/{id}", "GET", "file_by_id"},
		{"/api/v1/files/{id}/metadata", "GET", "file_metadata"},
		{"/api/{version}/{resource}", "GET", "version_resource"},
		{"/api/{version}/{resource}/{id}", "GET", "version_resource_id"},
	}

	for _, route := range routes {
		router.AddRoute(NewTestRoute(route.path, route.method, route.name))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.FindBestMatch("GET", "/api/v1/files/123")
	}
}
