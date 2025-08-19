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

package custom_http_director

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
)

func TestJSONBodyTransformer(t *testing.T) {
	tests := []struct {
		name            string
		requestBody     map[string]any
		bodyTransform   *api_style.BodyTransform
		expectedBody    map[string]any
		expectedChanges []BodyTransformChange
	}{
		{
			name: "GPT-5 max_tokens transformation",
			requestBody: map[string]any{
				"model": "gpt-5",
				"messages": []any{
					map[string]any{"role": "user", "content": "Hello"},
				},
				"max_tokens":  150,
				"top_p":       0.9,
				"temperature": 0.8,
			},
			bodyTransform: &api_style.BodyTransform{
				Rename: map[string]string{
					"max_tokens": "max_completion_tokens",
				},
				Drop: []string{"top_p"},
				Default: map[string]any{
					"temperature": 1.0, // Should not override existing 0.8
				},
				Clamp: map[string]api_style.NumericClamp{
					"max_completion_tokens": {
						Min: func() *float64 { v := 1.0; return &v }(),
						Max: func() *float64 { v := 8192.0; return &v }(),
					},
				},
			},
			expectedBody: map[string]any{
				"model": "gpt-5",
				"messages": []any{
					map[string]any{"role": "user", "content": "Hello"},
				},
				"max_completion_tokens": int64(150), // Renamed from max_tokens
				"temperature":           0.8,        // Original value kept (not defaulted)
				// top_p should be dropped
			},
			expectedChanges: []BodyTransformChange{
				{Op: "rename", Key: "max_tokens", From: 150, To: "max_completion_tokens"},
				{Op: "drop", Key: "top_p", From: 0.9},
			},
		},
		{
			name: "clamping test",
			requestBody: map[string]any{
				"max_tokens": 50000, // Should be clamped to 8192
			},
			bodyTransform: &api_style.BodyTransform{
				Rename: map[string]string{
					"max_tokens": "max_completion_tokens",
				},
				Clamp: map[string]api_style.NumericClamp{
					"max_completion_tokens": {
						Min: func() *float64 { v := 1.0; return &v }(),
						Max: func() *float64 { v := 8192.0; return &v }(),
					},
				},
			},
			expectedBody: map[string]any{
				"max_completion_tokens": int64(8192), // Clamped from 50000
			},
			expectedChanges: []BodyTransformChange{
				{Op: "rename", Key: "max_tokens", From: 50000, To: "max_completion_tokens"},
				{Op: "clamp", Key: "max_completion_tokens", From: 50000.0, To: 8192.0},
			},
		},
		{
			name: "force set values",
			requestBody: map[string]any{
				"temperature": 0.8,
			},
			bodyTransform: &api_style.BodyTransform{
				Force: map[string]any{
					"temperature": 1.0,
					"stream":      false,
				},
			},
			expectedBody: map[string]any{
				"temperature": 1.0,
				"stream":      false,
			},
			expectedChanges: []BodyTransformChange{
				{Op: "force", Key: "temperature", From: 0.8, Value: 1.0},
				{Op: "force", Key: "stream", From: nil, Value: false},
			},
		},
		{
			name: "default values only for missing keys",
			requestBody: map[string]any{
				"temperature": 0.8, // Should not be overridden
			},
			bodyTransform: &api_style.BodyTransform{
				Default: map[string]any{
					"temperature": 1.0,   // Should not override existing
					"stream":      false, // Should be added
				},
			},
			expectedBody: map[string]any{
				"temperature": 0.8,
				"stream":      false,
			},
			expectedChanges: []BodyTransformChange{
				{Op: "default", Key: "stream", Value: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request body
			bodyBytes, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			// Create HTTP request with context containing model
			req, err := http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Add model to context (required by transformBody)
			ctx := ctxhelper.InitCtxMapIfNeed(req.Context())
			ctxhelper.PutModel(ctx, &modelpb.Model{Name: "test-model"})
			req = req.WithContext(ctx)

			// Create transformer
			transformer := &JSONBodyTransformer{}

			// Apply transformation
			err = transformer.Transform(req, tt.bodyTransform)
			assert.NoError(t, err)

			// Read and verify the transformed body
			bodyBytes, err = io.ReadAll(req.Body)
			assert.NoError(t, err)

			var actualBody map[string]any
			err = json.Unmarshal(bodyBytes, &actualBody)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody, actualBody)

			// Verify changes were stored in context
			if len(tt.expectedChanges) > 0 {
				transformResult, ok := ctxhelper.GetRequestBodyTransformChanges(ctx)
				assert.True(t, ok)
				assert.NotNil(t, transformResult)

				actualChanges := *transformResult.(*[]BodyTransformChange)
				assert.Len(t, actualChanges, len(tt.expectedChanges))

				for i, expected := range tt.expectedChanges {
					assert.Equal(t, expected.Op, actualChanges[i].Op)
					assert.Equal(t, expected.Key, actualChanges[i].Key)
					if expected.From != nil {
						assert.Equal(t, expected.From, actualChanges[i].From)
					}
					if expected.To != nil {
						assert.Equal(t, expected.To, actualChanges[i].To)
					}
					if expected.Value != nil {
						assert.Equal(t, expected.Value, actualChanges[i].Value)
					}
				}
			}
		})
	}
}

func TestBodyDirectorIntegration(t *testing.T) {
	t.Run("no transformation when no body config", func(t *testing.T) {
		requestBody := map[string]any{
			"max_tokens": 150,
		}

		bodyBytes, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		req, err := http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		apiStyleConfig := api_style.APIStyleConfig{
			// No Body configuration
		}

		err = bodyDirector(req, apiStyleConfig)
		assert.NoError(t, err)

		// Read body and verify it's unchanged
		bodyBytes, err = io.ReadAll(req.Body)
		assert.NoError(t, err)

		var actualBody map[string]any
		err = json.Unmarshal(bodyBytes, &actualBody)
		assert.NoError(t, err)

		expectedBody := map[string]any{
			"max_tokens": 150.0, // JSON unmarshaling converts numbers to float64
		}
		assert.Equal(t, expectedBody, actualBody)
	})

	t.Run("non-JSON content type ignored", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader([]byte("plain text")))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		apiStyleConfig := api_style.APIStyleConfig{
			Body: &api_style.BodyTransform{
				Rename: map[string]string{
					"max_tokens": "max_completion_tokens",
				},
			},
		}

		err = bodyDirector(req, apiStyleConfig)
		assert.NoError(t, err)

		// Body should remain unchanged
		bodyBytes, err := io.ReadAll(req.Body)
		assert.NoError(t, err)
		assert.Equal(t, "plain text", string(bodyBytes))
	})
}

func TestJSONBodyTransformerCanHandle(t *testing.T) {
	transformer := &JSONBodyTransformer{}

	tests := []struct {
		contentType string
		expected    bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"application/vnd.api+json", false}, // Only exact prefix match
		{"text/plain", false},
		{"application/xml", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			assert.Equal(t, tt.expected, transformer.CanHandle(tt.contentType))
		})
	}
}

func TestGetBodyTransformerByContentType(t *testing.T) {
	tests := []struct {
		contentType         string
		expectedTransformer bool
	}{
		{"application/json", true},
		{"application/json; charset=utf-8", true},
		{"text/plain", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			transformer := getBodyTransformerByContentType(tt.contentType)
			if tt.expectedTransformer {
				assert.NotNil(t, transformer)
				assert.IsType(t, &JSONBodyTransformer{}, transformer)
			} else {
				assert.Nil(t, transformer)
			}
		})
	}
}
