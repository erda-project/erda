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
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda/internal/apps/ai-proxy/cache/cachehelpers"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
)

// createProxyRequest creates a ProxyRequest for testing
func createProxyRequest(req *http.Request) *httputil.ProxyRequest {
	// Set default values
	if req.Method == "" {
		req.Method = "POST"
	}
	if req.URL == nil {
		req.URL = &url.URL{Path: "/test"}
	}

	// Ensure Out request has correct Body
	clonedReq := req.Clone(req.Context())
	if req.Body != nil {
		// Read original body
		bodyBytes, _ := io.ReadAll(req.Body)
		// Reset body for both requests
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		clonedReq.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}
	return &httputil.ProxyRequest{
		In:  req,
		Out: clonedReq,
	}
}

func Test_getMapOfAvailableNameWithModels(t *testing.T) {
	publisherValue, _ := structpb.NewValue("openai")
	modelIDValue, _ := structpb.NewValue("gpt-35-turbo")
	modelNameValue, _ := structpb.NewValue("gpt-3.5-turbo")

	models := []*cachehelpers.ModelWithProvider{
		{
			Model: &modelpb.Model{
				Name: "GPT-3.5 Turbo",
				Metadata: &metadatapb.Metadata{
					Public: map[string]*structpb.Value{
						"publisher":  publisherValue,
						"model_id":   modelIDValue,
						"model_name": modelNameValue,
					},
				},
			},
			Provider: &providerpb.ModelProvider{
				Type: "Azure",
			},
		},
	}

	result := getMapOfAvailableNameWithModels(models)

	// Test that all expected keys exist based on current implementation logic
	// The function generates keys based on: ${publisher}/${model.name}, ${model.name}, and ${provider.type}/${model.name}
	expectedKeys := []string{
		"openai/GPT-3.5 Turbo", // publisher/model.name
		"GPT-3.5 Turbo",        // model.name
		"Azure/GPT-3.5 Turbo",  // provider.type/model.name (same as publisher in this case)
	}

	for _, key := range expectedKeys {
		if _, exists := result[key]; !exists {
			t.Errorf("Expected key %s not found in result", key)
		}
	}

	// Test that empty keys should not exist
	if _, exists := result[""]; exists {
		t.Error("Empty key should not exist in result")
	}
	if _, exists := result["/"]; exists {
		t.Error("Slash-only key should not exist in result")
	}
}

func Test_modelGetter(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	older := now.Add(-time.Hour)

	tests := []struct {
		name     string
		models   []*modelpb.Model
		expected *modelpb.Model
	}{
		{
			name:     "empty models",
			models:   []*modelpb.Model{},
			expected: nil,
		},
		{
			name: "single model",
			models: []*modelpb.Model{
				{
					Id:        "model1",
					UpdatedAt: timestamppb.New(now),
				},
			},
			expected: &modelpb.Model{
				Id:        "model1",
				UpdatedAt: timestamppb.New(now),
			},
		},
		{
			name: "multiple models - should return latest",
			models: []*modelpb.Model{
				{
					Id:        "model1",
					UpdatedAt: timestamppb.New(older),
				},
				{
					Id:        "model2",
					UpdatedAt: timestamppb.New(now),
				},
			},
			expected: &modelpb.Model{
				Id:        "model2",
				UpdatedAt: timestamppb.New(now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := modelGetter(ctx, tt.models)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected model, got nil")
					return
				}
				if result.Id != tt.expected.Id {
					t.Errorf("Expected model ID %s, got %s", tt.expected.Id, result.Id)
				}
			}
		})
	}
}
