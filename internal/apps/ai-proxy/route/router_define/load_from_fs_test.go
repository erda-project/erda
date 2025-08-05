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
	"encoding/json"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

func TestValidateFiltersExistence(t *testing.T) {
	// Save original FilterFactory state
	originalRequestFilters := filter_define.FilterFactory.RequestFilters
	originalResponseFilters := filter_define.FilterFactory.ResponseFilters
	defer func() {
		filter_define.FilterFactory.RequestFilters = originalRequestFilters
		filter_define.FilterFactory.ResponseFilters = originalResponseFilters
	}()

	// Reset FilterFactory to test state
	filter_define.FilterFactory.RequestFilters = make(map[string]filter_define.RequestRewriterCreator)
	filter_define.FilterFactory.ResponseFilters = make(map[string]filter_define.ResponseModifierCreator)

	// Register some test filters
	filter_define.FilterFactory.RequestFilters["test-request-filter"] = func(name string, config json.RawMessage) filter_define.ProxyRequestRewriter {
		return nil
	}
	filter_define.FilterFactory.ResponseFilters["test-response-filter"] = func(name string, config json.RawMessage) filter_define.ProxyResponseModifier {
		return nil
	}

	testCases := []struct {
		name          string
		routes        []*Route
		expectError   bool
		errorContains string
	}{
		{
			name: "All filters exist",
			routes: []*Route{
				{
					Path:   "/test",
					Method: "GET",
					RequestFilters: []filter_define.FilterConfig{
						{Name: "test-request-filter"},
					},
					ResponseFilters: []filter_define.FilterConfig{
						{Name: "test-response-filter"},
					},
				},
			},
			expectError: false,
		},
		{
			name: "Request filter does not exist",
			routes: []*Route{
				{
					Path:   "/test",
					Method: "GET",
					RequestFilters: []filter_define.FilterConfig{
						{Name: "nonexistent-request-filter"},
					},
				},
			},
			expectError:   true,
			errorContains: "request filter 'nonexistent-request-filter'",
		},
		{
			name: "Response filter does not exist",
			routes: []*Route{
				{
					Path:   "/test",
					Method: "GET",
					ResponseFilters: []filter_define.FilterConfig{
						{Name: "nonexistent-response-filter"},
					},
				},
			},
			expectError:   true,
			errorContains: "response filter 'nonexistent-response-filter'",
		},
		{
			name: "Multiple filters do not exist",
			routes: []*Route{
				{
					Path:   "/test1",
					Method: "GET",
					RequestFilters: []filter_define.FilterConfig{
						{Name: "missing-req-filter-1"},
					},
				},
				{
					Path:   "/test2",
					Method: "POST",
					ResponseFilters: []filter_define.FilterConfig{
						{Name: "missing-resp-filter-1"},
					},
				},
			},
			expectError:   true,
			errorContains: "missing-req-filter-1",
		},
		{
			name: "Duplicate missing filter is reported only once",
			routes: []*Route{
				{
					Path:   "/test1",
					Method: "GET",
					RequestFilters: []filter_define.FilterConfig{
						{Name: "duplicate-missing-filter"},
					},
				},
				{
					Path:   "/test2",
					Method: "POST",
					RequestFilters: []filter_define.FilterConfig{
						{Name: "duplicate-missing-filter"},
					},
				},
			},
			expectError:   true,
			errorContains: "request filter 'duplicate-missing-filter'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFiltersExistence(tc.routes)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Error message does not contain expected content. Expected: %s, actual error: %s", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got error: %v", err)
				}
			}
		})
	}
}

func TestLoadRoutesFromEmbeddedDirExtended(t *testing.T) {
	// Create a mock filesystem with test YAML files
	testFS := fstest.MapFS{
		"openai-compatible/test1.yaml": &fstest.MapFile{
			Data: []byte(`routes:
  - path: "/v1/chat/completions"
    method: "POST"
    request_filters:
      - name: "initialize"
    response_filters:
      - name: "set-ai-proxy-header"`),
		},
		"proxy/test2.yaml": &fstest.MapFile{
			Data: []byte(`routes:
  - path: "/v1/proxy/test"
    method: "GET"
    request_filters:
      - name: "initialize"
    response_filters:
      - name: "set-ai-proxy-header"`),
		},
		"test3.yaml": &fstest.MapFile{
			Data: []byte(`routes:
  - path: "/v1/models"
    method: "GET"
    request_filters:
      - name: "initialize"
    response_filters:
      - name: "set-ai-proxy-header"`),
		},
	}

	// Manually call loadRoutesFromFS to load all yaml files
	yamlFile := &YamlFile{Routes: []*Route{}}
	err := loadRoutesFromFS(testFS, yamlFile)
	if err != nil {
		t.Fatalf("Failed to load routes: %v", err)
	}

	if yamlFile == nil {
		t.Fatal("YamlFile is nil")
	}

	if len(yamlFile.Routes) == 0 {
		t.Fatal("No routes loaded")
	}

	// Verify route count (should include routes from subdirectories and root)
	// test1.yaml(1) + test2.yaml(1) + test3.yaml(1) = 3
	expectedRoutes := 3
	if len(yamlFile.Routes) != expectedRoutes {
		t.Errorf("Expected %d routes, got %d", expectedRoutes, len(yamlFile.Routes))
	}

	t.Logf("Successfully loaded %d routes", len(yamlFile.Routes))

	// Create route mapping for verification
	routeMap := make(map[string]*Route)
	for _, route := range yamlFile.Routes {
		key := route.Method + " " + route.Path
		routeMap[key] = route
		t.Logf("Route: %s %s", route.Method, route.Path)
	}

	// Verify that routes from different directories are loaded
	t.Run("Routes from subdirectories loaded", func(t *testing.T) {
		// Route from openai-compatible subdirectory
		if routeMap["POST /v1/chat/completions"] == nil {
			t.Error("Route from openai-compatible subdirectory not found")
		}

		// Route from proxy subdirectory
		if routeMap["GET /v1/proxy/test"] == nil {
			t.Error("Route from proxy subdirectory not found")
		}

		// Route from root directory
		if routeMap["GET /v1/models"] == nil {
			t.Error("Route from root directory not found")
		}
	})
}
