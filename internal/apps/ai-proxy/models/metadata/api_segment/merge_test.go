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

package api_segment

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
)

func TestMergeAPIStyleConfig(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		pathMatcher string
		apiSegments []*API
		expected    *api_style.APIStyleConfig
	}{
		{
			name:        "single provider config",
			method:      "POST",
			pathMatcher: "/v1/chat/completions",
			apiSegments: []*API{
				{
					APIStyleConfigs: map[string]api_style.APIStyleConfig{
						"POST:/v1/chat/completions": {
							Host:   "api.openai.com",
							Scheme: "https",
							Headers: map[string][]string{
								"Authorization": {"set", "Bearer ${@provider.apiKey}"},
							},
						},
					},
				},
			},
			expected: &api_style.APIStyleConfig{
				Host:   "api.openai.com",
				Scheme: "https",
				Headers: map[string][]string{
					"Authorization": {"set", "Bearer ${@provider.apiKey}"},
				},
			},
		},
		{
			name:        "provider and model config merge",
			method:      "POST",
			pathMatcher: "/v1/chat/completions",
			apiSegments: []*API{
				// Provider level (lower priority)
				{
					APIStyleConfigs: map[string]api_style.APIStyleConfig{
						"POST:/v1/chat/completions": {
							Host:   "api.openai.com",
							Scheme: "https",
							Headers: map[string][]string{
								"Authorization": {"set", "Bearer ${@provider.apiKey}"},
								"User-Agent":    {"set", "erda-ai-proxy"},
							},
						},
					},
				},
				// Model level (higher priority)
				{
					APIStyleConfigs: map[string]api_style.APIStyleConfig{
						"POST:/v1/chat/completions": {
							Host: "api.openai.com/v2", // Override host
							Body: &api_style.BodyTransform{
								Rename: map[string]string{
									"max_tokens": "max_completion_tokens",
								},
							},
							Headers: map[string][]string{
								"X-Model": {"set", "gpt-5"}, // Add new header
							},
						},
					},
				},
			},
			expected: &api_style.APIStyleConfig{
				Host:   "api.openai.com/v2", // Overridden by model
				Scheme: "https",             // From provider
				Headers: map[string][]string{
					"Authorization": {"set", "Bearer ${@provider.apiKey}"}, // From provider
					"User-Agent":    {"set", "erda-ai-proxy"},              // From provider
					"X-Model":       {"set", "gpt-5"},                      // From model
				},
				Body: &api_style.BodyTransform{
					Rename: map[string]string{
						"max_tokens": "max_completion_tokens",
					},
				},
			},
		},
		{
			name:        "body transformation replace",
			method:      "POST",
			pathMatcher: "/v1/chat/completions",
			apiSegments: []*API{
				// Provider level
				{
					APIStyleConfigs: map[string]api_style.APIStyleConfig{
						"POST:/v1/chat/completions": {
							Body: &api_style.BodyTransform{
								Drop: []string{
									"top_p",
								},
								Default: map[string]any{
									"temperature": 1.0,
								},
							},
						},
					},
				},
				// Model level
				{
					APIStyleConfigs: map[string]api_style.APIStyleConfig{
						"POST:/v1/chat/completions": {
							Body: &api_style.BodyTransform{
								Rename: map[string]string{
									"max_tokens": "max_completion_tokens",
								},
								Drop: []string{
									"frequency_penalty",
								},
								Clamp: map[string]api_style.NumericClamp{
									"max_completion_tokens": {
										Min: func() *float64 { v := 1.0; return &v }(),
										Max: func() *float64 { v := 8192.0; return &v }(),
									},
								},
							},
						},
					},
				},
			},
			expected: &api_style.APIStyleConfig{
				Body: &api_style.BodyTransform{
					Rename: map[string]string{
						"max_tokens": "max_completion_tokens",
					},
					Drop: []string{
						"frequency_penalty", // From model
					},
					Clamp: map[string]api_style.NumericClamp{
						"max_completion_tokens": {
							Min: func() *float64 { v := 1.0; return &v }(),
							Max: func() *float64 { v := 8192.0; return &v }(),
						},
					},
				},
			},
		},
		{
			name:        "no matching config",
			method:      "GET",
			pathMatcher: "/v1/models",
			apiSegments: []*API{
				{
					APIStyleConfigs: map[string]api_style.APIStyleConfig{
						"POST:/v1/chat/completions": {
							Host: "api.openai.com",
						},
					},
				},
			},
			expected: nil,
		},
		{
			name:        "empty api segments",
			method:      "POST",
			pathMatcher: "/v1/chat/completions",
			apiSegments: []*API{},
			expected:    nil,
		},
		{
			name:        "GPT-5 specific case",
			method:      "POST",
			pathMatcher: "/v1/chat/completions",
			apiSegments: []*API{
				// Provider level - OpenAI base config
				{
					APIStyleConfigs: map[string]api_style.APIStyleConfig{
						"POST:/v1/chat/completions": {
							Host:   "api.openai.com",
							Scheme: "https",
							Headers: map[string][]string{
								"Authorization": {"set", "Bearer ${@provider.apiKey}"},
								"Content-Type":  {"set", "application/json"},
							},
						},
					},
				},
				// Model level - GPT-5 specific transformations
				{
					APIStyleConfigs: map[string]api_style.APIStyleConfig{
						"*:*": {
							Body: &api_style.BodyTransform{
								Rename: map[string]string{
									"max_tokens": "max_completion_tokens",
								},
							},
						},
					},
				},
			},
			expected: &api_style.APIStyleConfig{
				Host:   "api.openai.com",
				Scheme: "https",
				Headers: map[string][]string{
					"Authorization": {"set", "Bearer ${@provider.apiKey}"},
					"Content-Type":  {"set", "application/json"},
				},
				Body: &api_style.BodyTransform{
					Rename: map[string]string{
						"max_tokens": "max_completion_tokens",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeAPIStyleConfig(tt.method, tt.pathMatcher, tt.apiSegments...)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.Host, result.Host)
			assert.Equal(t, tt.expected.Scheme, result.Scheme)
			assert.Equal(t, tt.expected.Method, result.Method)
			assert.Equal(t, tt.expected.Headers, result.Headers)

			// Compare Body transformation
			if tt.expected.Body == nil {
				assert.Nil(t, result.Body)
			} else {
				assert.NotNil(t, result.Body)
				if tt.expected.Body == nil {
					assert.Nil(t, result.Body)
				} else {
					assert.NotNil(t, result.Body)
					assert.Equal(t, tt.expected.Body.Rename, result.Body.Rename)
					assert.Equal(t, tt.expected.Body.Drop, result.Body.Drop)
					assert.Equal(t, tt.expected.Body.Default, result.Body.Default)
					assert.Equal(t, tt.expected.Body.Clamp, result.Body.Clamp)
				}
			}
		})
	}
}
