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

package thinking_handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/router_define/path_matcher"
	extractorsregistry "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/extractors/registry"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
)

func TestDetectAndAggregateThinking(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected *types.CommonThinking
	}{
		{
			name: "doubao style - enabled",
			input: map[string]any{
				"thinking": map[string]any{
					"type": "enabled",
				},
			},
			expected: &types.CommonThinking{
				Mode: modePtr(types.ModeOn),
			},
		},
		{
			name: "qwen style",
			input: map[string]any{
				"enable_thinking": true,
				"thinking_budget": 4096,
			},
			expected: &types.CommonThinking{
				Mode:         modePtr(types.ModeOn),
				BudgetTokens: intPtr(4096),
			},
		},
		{
			name: "openai responses style",
			input: map[string]any{
				"reasoning": map[string]any{
					"effort": "high",
				},
			},
			expected: &types.CommonThinking{
				Mode:   modePtr(types.ModeOn), // OpenAI extractor infers mode=on when effort exists
				Effort: effortPtr(types.EffortHigh),
			},
		},
		{
			name: "openai chat style",
			input: map[string]any{
				"reasoning_effort": "medium",
			},
			expected: &types.CommonThinking{
				Mode:   modePtr(types.ModeOn), // OpenAI extractor infers mode=on when effort exists
				Effort: effortPtr(types.EffortMedium),
			},
		},
		{
			name: "mixed with priority - thinking.type wins",
			input: map[string]any{
				"thinking": map[string]any{
					"type": "disabled",
				},
				"enable_thinking": true,
			},
			expected: &types.CommonThinking{
				Mode: modePtr(types.ModeOff),
			},
		},
		{
			name: "anthropic full",
			input: map[string]any{
				"thinking": map[string]any{
					"type":          "enabled",
					"budget_tokens": 8192,
				},
			},
			expected: &types.CommonThinking{
				Mode:         modePtr(types.ModeOn),
				BudgetTokens: intPtr(8192),
			},
		},
		{
			name:     "no thinking fields",
			input:    map[string]any{"messages": []any{}},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the new registry-based approach
			extractorsRegistry := extractorsregistry.NewRegistry()
			extractResults, err := extractorsRegistry.ExtractAll(tt.input)
			require.NoError(t, err)

			result, err := extractorsregistry.mergeAndValidateResults(extractResults)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEncodeThinking(t *testing.T) {
	tests := []struct {
		name           string
		ct             *types.CommonThinking
		targetProvider string
		targetAPI      string
		modelName      string
		inputBody      map[string]any
		expectedBody   map[string]any
		expectError    bool
	}{
		{
			name: "doubao to openai chat",
			ct: &types.CommonThinking{
				Mode: modePtr(types.ModeOff),
			},
			targetProvider: types.ProviderOpenAI,
			targetAPI:      types.APIChat,
			modelName:      "o1-preview",
			inputBody:      map[string]any{},
			expectedBody: map[string]any{
				"reasoning_effort": "minimal",
			},
		},
		{
			name: "qwen to anthropic",
			ct: &types.CommonThinking{
				Mode:         modePtr(types.ModeOn),
				BudgetTokens: intPtr(4096),
			},
			targetProvider: types.ProviderAnthropic,
			targetAPI:      types.APIChat,
			modelName:      "claude-3",
			inputBody:      map[string]any{},
			expectedBody: map[string]any{
				"thinking": map[string]any{
					"type":          "enabled",
					"budget_tokens": 4096,
				},
			},
		},
		{
			name: "openai responses to qwen",
			ct: &types.CommonThinking{
				Effort: effortPtr(types.EffortHigh),
			},
			targetProvider: types.ProviderQwen,
			targetAPI:      types.APIChat,
			modelName:      "qwen-max",
			inputBody:      map[string]any{},
			expectedBody: map[string]any{
				"enable_thinking": true,
			},
		},
		{
			name: "anthropic budget too small",
			ct: &types.CommonThinking{
				Mode:         modePtr(types.ModeOn),
				BudgetTokens: intPtr(512),
			},
			targetProvider: types.ProviderAnthropic,
			targetAPI:      types.APIChat,
			modelName:      "claude-3",
			inputBody:      map[string]any{},
			expectError:    true,
		},
		{
			name: "non-reasoning model to openai",
			ct: &types.CommonThinking{
				Effort: effortPtr(types.EffortHigh),
			},
			targetProvider: types.ProviderOpenAI,
			targetAPI:      types.APIChat,
			modelName:      "gpt-4",
			inputBody:      map[string]any{},
			expectedBody:   map[string]any{}, // No fields added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := make(map[string]any)
			for k, v := range tt.inputBody {
				body[k] = v
			}

			// create a mock context with model and path information
			ctx := ctxhelper.InitCtxMapIfNeed(context.Background())

			// create mock protobuf metadata for testing OpenAI reasoning models  
			publicDataMap := map[string]any{
				"publisher": tt.targetProvider,
				"reasoning": "true",
			}
			if tt.modelName != "o1-preview" && tt.modelName != "o1" && tt.modelName != "o1-mini" {
				// For non-reasoning models, remove reasoning capability
				publicDataMap["reasoning"] = "false"
			}

			publicData, _ := structpb.NewStruct(publicDataMap)
			modelMetadata := &metadatapb.Metadata{
				Public: publicData.Fields,
			}

			// create mock protobuf model
			mockModel := &modelpb.Model{
				Name:      tt.modelName,
				Metadata:  modelMetadata,
				Publisher: tt.targetProvider,
			}

			// create mock path matcher
			pathMatcherImpl := &path_matcher.PathMatcher{
				Pattern: "/v1/chat/completions", // default to chat completions
			}
			if tt.targetAPI == types.APIResponses {
				pathMatcherImpl.Pattern = "/v1/responses"
			}

			// set context values
			ctxhelper.PutModel(ctx, mockModel)
			ctxhelper.PutPathMatcher(ctx, pathMatcherImpl)

			err := EncodeThinking(ctx, tt.ct, body)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}

func TestCleanUnrelatedThinkingFields(t *testing.T) {
	tests := []struct {
		name           string
		inputBody      map[string]any
		targetProvider string
		targetAPI      string
		expectedBody   map[string]any
	}{
		{
			name: "clean for openai chat",
			inputBody: map[string]any{
				"thinking":         map[string]any{"type": "enabled"},
				"enable_thinking":  true,
				"thinking_budget":  4096,
				"reasoning_effort": "medium",
				"reasoning": map[string]any{
					"effort": "high",
					"other":  "field",
				},
			},
			targetProvider: types.ProviderOpenAI,
			targetAPI:      types.APIChat,
			expectedBody: map[string]any{
				"reasoning_effort": "medium",
				"reasoning": map[string]any{
					"other": "field",
				},
			},
		},
		{
			name: "clean for anthropic",
			inputBody: map[string]any{
				"thinking": map[string]any{
					"type":          "enabled",
					"budget_tokens": 4096,
				},
				"reasoning":        map[string]any{"effort": "high"},
				"reasoning_effort": "medium",
				"enable_thinking":  true,
				"thinking_budget":  2048,
			},
			targetProvider: types.ProviderAnthropic,
			targetAPI:      types.APIChat,
			expectedBody: map[string]any{
				"thinking": map[string]any{
					"type":          "enabled",
					"budget_tokens": 4096,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := make(map[string]any)
			for k, v := range tt.inputBody {
				body[k] = v
			}

			CleanUnrelatedThinkingFields(body, tt.targetProvider, tt.targetAPI)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}

func TestThinkingHandlerIntegration(t *testing.T) {
	tests := []struct {
		name         string
		inputBody    map[string]any
		provider     string
		apiPath      string
		modelName    string
		expectedBody map[string]any
		expectAudit  bool
	}{
		{
			name: "doubao to openai integration",
			inputBody: map[string]any{
				"thinking": map[string]any{
					"type": "disabled",
				},
				"messages": []any{
					map[string]any{"role": "user", "content": "hello"},
				},
			},
			provider:  "openai",
			apiPath:   "/v1/chat/completions",
			modelName: "o1-preview",
			expectedBody: map[string]any{
				"reasoning_effort": "minimal",
				"messages": []any{
					map[string]any{"role": "user", "content": "hello"},
				},
			},
			expectAudit: true,
		},
		{
			name: "no thinking fields",
			inputBody: map[string]any{
				"messages": []any{
					map[string]any{"role": "user", "content": "hello"},
				},
			},
			provider:  "openai",
			apiPath:   "/v1/chat/completions",
			modelName: "gpt-4",
			expectedBody: map[string]any{
				"messages": []any{
					map[string]any{"role": "user", "content": "hello"},
				},
			},
			expectAudit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create filter
			filter := &ThinkingHandler{}

			// create test request
			bodyBytes, _ := json.Marshal(tt.inputBody)
			req := &http.Request{
				Method: "POST",
				URL:    &url.URL{Path: tt.apiPath},
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: io.NopCloser(bytes.NewReader(bodyBytes)),
			}

			// create context with model metadata using helper function
			model := createTestModel(tt.provider, tt.apiPath, tt.modelName)

			ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
			ctxhelper.PutModel(ctx, model)

			// create and add path matcher
			pathMatcherImpl := &path_matcher.PathMatcher{
				Pattern: tt.apiPath,
			}
			ctxhelper.PutPathMatcher(ctx, pathMatcherImpl)

			req = req.WithContext(ctx)

			// create proxy request
			pr := &httputil.ProxyRequest{
				In:  req,
				Out: req.Clone(ctx),
			}

			// execute filter
			err := filter.OnProxyRequest(pr)
			require.NoError(t, err)

			// read and parse response body
			bodyBytes, err = io.ReadAll(pr.Out.Body)
			require.NoError(t, err)

			var actualBody map[string]any
			err = json.Unmarshal(bodyBytes, &actualBody)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedBody, actualBody)

			// check audit header
			auditHeader := pr.Out.Header.Get(vars.XAIProxyRequestBodyTransform)
			if tt.expectAudit {
				assert.NotEmpty(t, auditHeader)

				var auditRecords []AuditRecord
				err = json.Unmarshal([]byte(auditHeader), &auditRecords)
				require.NoError(t, err)

				assert.Len(t, auditRecords, 1)
				assert.Equal(t, "thinking_handler", auditRecords[0].Op)
				assert.NotNil(t, auditRecords[0].From)
				assert.NotNil(t, auditRecords[0].To)
			} else {
				assert.Empty(t, auditHeader)
			}
		})
	}
}

func TestIsOpenAIReasoningModel(t *testing.T) {
	tests := []struct {
		name      string
		modelName string
		expected  bool
	}{
		{
			"o1-preview reasoning model",
			"o1-preview",
			true,
		},
		{
			"o1 reasoning model",
			"o1",
			true,
		},
		{
			"o1-mini reasoning model",
			"o1-mini",
			true,
		},
		{
			"gpt-4 non-reasoning model",
			"gpt-4",
			false,
		},
		{
			"claude non-reasoning model",
			"claude-3",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := types.IsOpenAIReasoningModel(tt.modelName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions for tests
func createTestModel(provider, apiPath, modelName string) *modelpb.Model {
	publicDataMap := map[string]any{
		"provider": map[string]any{
			"name": provider,
		},
		"api_segment": map[string]any{
			"path": apiPath,
		},
		"publisher": provider,
	}

	// Add reasoning capability for OpenAI reasoning models
	if provider == "openai" && isReasoningModelName(modelName) {
		publicDataMap["reasoning"] = "true"
		publicDataMap["enable_thinking"] = "true"
	} else {
		publicDataMap["reasoning"] = "false"
	}

	publicData, _ := structpb.NewStruct(publicDataMap)
	modelMetadata := &metadatapb.Metadata{
		Public: publicData.Fields,
	}

	return &modelpb.Model{
		Name:      modelName,
		Metadata:  modelMetadata,
		Publisher: provider,
	}
}

// Helper function to check if a model name indicates reasoning capability
func isReasoningModelName(modelName string) bool {
	reasoningModels := []string{"o1", "o1-preview", "o1-mini", "o3", "o3-mini"}
	for _, model := range reasoningModels {
		if modelName == model || (len(modelName) > len(model) &&
			modelName[:len(model)] == model && modelName[len(model)] == '-') {
			return true
		}
	}
	return false
}

// Helper functions for test cases
func modePtr(mode types.CommonThinkingMode) *types.CommonThinkingMode {
	return &mode
}

func effortPtr(effort types.CommonThinkingEffort) *types.CommonThinkingEffort {
	return &effort
}

func intPtr(i int) *int {
	return &i
}
