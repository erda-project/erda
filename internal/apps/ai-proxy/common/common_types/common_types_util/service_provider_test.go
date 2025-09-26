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

package common_types_util

import (
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
)

func TestGetServiceProviderType(t *testing.T) {
	tests := []struct {
		name     string
		provider *modelproviderpb.ModelProvider
		want     string
	}{
		{
			name: "has service provider type",
			provider: &modelproviderpb.ModelProvider{
				Metadata: &metadatapb.Metadata{
					Public: map[string]*structpb.Value{
						metaKeyServiceProviderType: structpb.NewStringValue(common_types.ServiceProviderTypeVolcengineArk.String()),
					},
				},
			},
			want: common_types.ServiceProviderTypeVolcengineArk.String(),
		},
		{
			name:     "nil provider",
			provider: nil,
			want:     "",
		},
		{
			name:     "nil metadata",
			provider: &modelproviderpb.ModelProvider{},
			want:     "",
		},
		{
			name: "nil public map",
			provider: &modelproviderpb.ModelProvider{
				Metadata: &metadatapb.Metadata{},
			},
			want: "",
		},
		{
			name: "key missing",
			provider: &modelproviderpb.ModelProvider{
				Metadata: &metadatapb.Metadata{
					Public: map[string]*structpb.Value{
						"other": structpb.NewStringValue("value"),
					},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetServiceProviderType(tt.provider); got != tt.want {
				t.Fatalf("GetServiceProviderType() = %q, want %q", got, tt.want)
			}
		})
	}
}
