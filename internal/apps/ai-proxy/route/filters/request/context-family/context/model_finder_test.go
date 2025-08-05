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

package context

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define/path_matcher"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestHeaderFinder_Find(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		expectedResult *ModelIdentifier
		expectedError  bool
	}{
		{
			name: "find model by model id header",
			headers: map[string]string{
				vars.XAIProxyModelId: "model-uuid-123",
			},
			expectedResult: &ModelIdentifier{ID: "model-uuid-123"},
		},
		{
			name: "find model by model name header",
			headers: map[string]string{
				vars.XAIProxyModel: "gpt-4",
			},
			expectedResult: &ModelIdentifier{Name: "gpt-4"},
		},
		{
			name: "find model by model name header with UUID format",
			headers: map[string]string{
				vars.XAIProxyModel: "gpt-4 [ID:uuid-456]",
			},
			expectedResult: &ModelIdentifier{ID: "uuid-456", Name: "gpt-4 [ID:uuid-456]"},
		},
		{
			name: "find model by model name header (alternative)",
			headers: map[string]string{
				vars.XAIProxyModelName: "claude-3",
			},
			expectedResult: &ModelIdentifier{Name: "claude-3"},
		},
		{
			name: "priority test - model id takes precedence",
			headers: map[string]string{
				vars.XAIProxyModelId:   "uuid-priority",
				vars.XAIProxyModel:     "gpt-4",
				vars.XAIProxyModelName: "claude-3",
			},
			expectedResult: &ModelIdentifier{ID: "uuid-priority"},
		},
		{
			name: "no model headers",
			headers: map[string]string{
				"Authorization": "Bearer token",
			},
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			finder := &HeaderFinder{}
			result, err := finder.Find(req, nil)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestPathFinder_Find(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		expectedResult *ModelIdentifier
		expectedError  bool
	}{
		{
			name: "find model from path parameter",
			setupContext: func() context.Context {
				ctx := context.Background()
				// Create path matcher and set it in context
				pm := path_matcher.NewPathMatcher("/test/{model}")
				pm.SetValue("model", "gpt-4")
				return context.WithValue(ctx, vars.CtxKeyPathMatcher{}, pm)
			},
			expectedResult: &ModelIdentifier{Name: "gpt-4"},
		},
		{
			name: "find model with UUID format from path",
			setupContext: func() context.Context {
				ctx := context.Background()
				pm := path_matcher.NewPathMatcher("/test/{model}")
				pm.SetValue("model", "gpt-4 [ID:uuid-789]")
				return context.WithValue(ctx, vars.CtxKeyPathMatcher{}, pm)
			},
			expectedResult: &ModelIdentifier{ID: "uuid-789", Name: "gpt-4 [ID:uuid-789]"},
		},
		{
			name: "no model in path parameters",
			setupContext: func() context.Context {
				ctx := context.Background()
				pm := path_matcher.NewPathMatcher("/test/{other}")
				pm.SetValue("other", "value")
				return context.WithValue(ctx, vars.CtxKeyPathMatcher{}, pm)
			},
			expectedResult: nil,
		},
		{
			name: "no path matcher in context",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			req := httptest.NewRequest("POST", "/test", nil)
			req = req.WithContext(ctx)

			finder := &PathFinder{}
			result, err := finder.Find(req, nil)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestJSONBodyFinder_FindModelName(t *testing.T) {
	tests := []struct {
		name          string
		contentType   string
		body          string
		fieldKey      string
		expectedModel string
		expectedError bool
	}{
		{
			name:          "find model from JSON body",
			contentType:   "application/json",
			body:          `{"model": "gpt-4", "messages": []}`,
			fieldKey:      "model",
			expectedModel: "gpt-4",
		},
		{
			name:          "model field not found in JSON",
			contentType:   "application/json",
			body:          `{"messages": []}`,
			fieldKey:      "model",
			expectedModel: "",
		},
		{
			name:          "non-JSON content type",
			contentType:   "text/plain",
			body:          `{"model": "gpt-4"}`,
			fieldKey:      "model",
			expectedModel: "",
		},
		{
			name:          "invalid JSON",
			contentType:   "application/json",
			body:          `{"model": "gpt-4"`,
			fieldKey:      "model",
			expectedError: true,
		},
		{
			name:          "model field is not string",
			contentType:   "application/json",
			body:          `{"model": 123}`,
			fieldKey:      "model",
			expectedModel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			finder := &JSONBodyFinder{}
			result, err := finder.FindModelName(req, tt.fieldKey)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedModel, result)
			}
		})
	}
}

func TestFormBodyFinder_FindModelName(t *testing.T) {
	tests := []struct {
		name          string
		contentType   string
		body          string
		fieldKey      string
		expectedModel string
		expectedError bool
	}{
		{
			name:          "find model from form body",
			contentType:   "application/x-www-form-urlencoded",
			body:          "model=gpt-4&temperature=0.7",
			fieldKey:      "model",
			expectedModel: "gpt-4",
		},
		{
			name:          "model field not found in form",
			contentType:   "application/x-www-form-urlencoded",
			body:          "temperature=0.7",
			fieldKey:      "model",
			expectedModel: "",
		},
		{
			name:          "non-form content type",
			contentType:   "application/json",
			body:          "model=gpt-4",
			fieldKey:      "model",
			expectedModel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			finder := &FormBodyFinder{}
			result, err := finder.FindModelName(req, tt.fieldKey)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedModel, result)
			}
		})
	}
}

func TestMultipartFormBodyFinder_FindModelName(t *testing.T) {
	tests := []struct {
		name          string
		setupRequest  func() *http.Request
		fieldKey      string
		expectedModel string
		expectedError bool
	}{
		{
			name: "find model from multipart form",
			setupRequest: func() *http.Request {
				body := strings.NewReader(`--boundary123
Content-Disposition: form-data; name="model"

gpt-4
--boundary123
Content-Disposition: form-data; name="temperature"

0.7
--boundary123--`)
				req := httptest.NewRequest("POST", "/test", body)
				req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary123")
				return req
			},
			fieldKey:      "model",
			expectedModel: "gpt-4",
		},
		{
			name: "model field not found in multipart form",
			setupRequest: func() *http.Request {
				body := strings.NewReader(`--boundary123
Content-Disposition: form-data; name="temperature"

0.7
--boundary123--`)
				req := httptest.NewRequest("POST", "/test", body)
				req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary123")
				return req
			},
			fieldKey:      "model",
			expectedModel: "",
		},
		{
			name: "non-multipart content type",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", strings.NewReader("test"))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			fieldKey:      "model",
			expectedModel: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()

			finder := &MultipartFormBodyFinder{}
			result, err := finder.FindModelName(req, tt.fieldKey)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedModel, result)
			}
		})
	}
}

func TestBodyFinder_Find(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedResult *ModelIdentifier
		expectedError  bool
	}{
		{
			name: "find model from JSON body",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"model": "gpt-4"}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedResult: &ModelIdentifier{Name: "gpt-4"},
		},
		{
			name: "find model from form body",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", strings.NewReader("model=claude-3"))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				return req
			},
			expectedResult: &ModelIdentifier{Name: "claude-3"},
		},
		{
			name: "no model in body",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"temperature": 0.7}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedResult: nil,
		},
		{
			name: "model with UUID format",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"model": "gpt-4 [ID:uuid-123]"}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedResult: &ModelIdentifier{ID: "uuid-123", Name: "gpt-4 [ID:uuid-123]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()

			finder := &BodyFinder{}
			result, err := finder.Find(req, nil)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestFindModelIdentifier(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		requestCtx     interface{}
		expectedResult *ModelIdentifier
		expectedError  bool
	}{
		{
			name: "find model from header (highest priority)",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"model": "body-model"}`))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set(vars.XAIProxyModelId, "header-uuid")
				return req
			},
			expectedResult: &ModelIdentifier{ID: "header-uuid"},
		},
		{
			name: "find model from path when no header",
			setupRequest: func() *http.Request {
				ctx := context.Background()
				pm := path_matcher.NewPathMatcher("/test/{model}")
				pm.SetValue("model", "path-model")
				ctx = context.WithValue(ctx, vars.CtxKeyPathMatcher{}, pm)

				req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"model": "body-model"}`))
				req.Header.Set("Content-Type", "application/json")
				req = req.WithContext(ctx)
				return req
			},
			expectedResult: &ModelIdentifier{Name: "path-model"},
		},
		{
			name: "find model from body when no header or path",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"model": "body-model"}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedResult: &ModelIdentifier{Name: "body-model"},
		},
		{
			name: "no model found",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"temperature": 0.7}`))
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()

			result, err := findModelIdentifier(req, tt.requestCtx)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestGetCustomBodyModelFieldByPathAndMethod(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		path          string
		expectedField string
	}{
		{
			name:          "anthropic path returns model field",
			method:        "POST",
			path:          "/proxy/v1/anthropic/messages",
			expectedField: "model",
		},
		{
			name:          "other path returns empty",
			method:        "POST",
			path:          "/proxy/v1/openai/chat/completions",
			expectedField: "",
		},
		{
			name:          "non-POST method returns empty",
			method:        "GET",
			path:          "/proxy/v1/anthropic/messages",
			expectedField: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCustomBodyModelFieldByPathAndMethod(tt.method, tt.path)
			assert.Equal(t, tt.expectedField, result)
		})
	}
}

func TestGetStandardFinderByContentType(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		expectedType string
	}{
		{
			name:         "form content type returns FormBodyFinder",
			contentType:  "application/x-www-form-urlencoded",
			expectedType: "*context.FormBodyFinder",
		},
		{
			name:         "multipart content type returns MultipartFormBodyFinder",
			contentType:  "multipart/form-data; boundary=123",
			expectedType: "*context.MultipartFormBodyFinder",
		},
		{
			name:         "json content type returns JSONBodyFinder",
			contentType:  "application/json",
			expectedType: "*context.JSONBodyFinder",
		},
		{
			name:         "unknown content type returns JSONBodyFinder",
			contentType:  "text/plain",
			expectedType: "*context.JSONBodyFinder",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStandardFinderByContentType(tt.contentType)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedType, getTypeName(result))
		})
	}
}

// Helper function to get type name for testing
func getTypeName(obj interface{}) string {
	switch obj.(type) {
	case *FormBodyFinder:
		return "*context.FormBodyFinder"
	case *MultipartFormBodyFinder:
		return "*context.MultipartFormBodyFinder"
	case *JSONBodyFinder:
		return "*context.JSONBodyFinder"
	default:
		return "unknown"
	}
}
