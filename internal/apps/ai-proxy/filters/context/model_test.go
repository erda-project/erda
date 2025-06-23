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
	"io"
	"net/http"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func Test_findModelID(t *testing.T) {
	type args struct {
		infor reverseproxy.HttpInfor
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "model id in header only",
			args: args{
				infor: reverseproxy.NewInfor(context.Background(), &http.Request{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey(vars.XAIProxyModelId): []string{"model-id-in-header"},
					},
				}),
			},
			want:    "model-id-in-header",
			wantErr: false,
		},
		{
			name: "model id in body only",
			args: args{
				infor: reverseproxy.NewInfor(context.Background(), &http.Request{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey("Content-Type"): []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(`{"model": "[ID:model-id-in-body]"}`)),
				}),
			},
			want:    "model-id-in-body",
			wantErr: false,
		},
		{
			name: "model id both in header and body, header is preferred",
			args: args{
				infor: reverseproxy.NewInfor(context.Background(), &http.Request{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey("Content-Type"):       []string{"application/json"},
						textproto.CanonicalMIMEHeaderKey(vars.XAIProxyModelId): []string{"model-id-in-header"},
					},
					Body: io.NopCloser(strings.NewReader(`{"model": "[ID:model-id-in-body]"}`)),
				}),
			},
			want:    "model-id-in-header",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findModelID(tt.args.infor)
			if (err != nil) != tt.wantErr {
				t.Errorf("findModelID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findModelID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMapOfAvailableNameWithModels(t *testing.T) {
	publisherValue, _ := structpb.NewValue("openai")
	modelIDValue, _ := structpb.NewValue("gpt-35-turbo")
	modelNameValue, _ := structpb.NewValue("gpt-3.5-turbo")

	models := []*modelpb.Model{
		{
			Name: "GPT-3.5 Turbo",
			Metadata: &metadatapb.Metadata{
				Public: map[string]*structpb.Value{
					"publisher":  publisherValue,
					"model_id":   modelIDValue,
					"model_name": modelNameValue,
				},
			},
		},
	}

	result := getMapOfAvailableNameWithModels(models)

	// Test that all expected keys exist
	expectedKeys := []string{
		"openai/GPT-3.5 Turbo", // publisher/model.name
		"openai/gpt-35-turbo",  // publisher/model_id
		"openai/gpt-3.5-turbo", // publisher/model_name
		"GPT-3.5 Turbo",        // model.name
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
