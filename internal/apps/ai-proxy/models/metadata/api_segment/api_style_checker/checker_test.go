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

package api_style_checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	providerpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
)

func TestCheckIsOpenAICompatible(t *testing.T) {
	type args struct {
		provider *providerpb.ServiceProvider
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "without api segment",
			args: args{
				provider: &providerpb.ServiceProvider{Metadata: &metadatapb.Metadata{
					Public: map[string]*structpb.Value{},
				}},
			},
			want: false,
		},
		{
			name: "with api segment, but not openai-compatible",
			args: args{
				provider: &providerpb.ServiceProvider{Metadata: &metadatapb.Metadata{
					Public: map[string]*structpb.Value{
						"api": func() *structpb.Value {
							apiMap := map[string]any{
								"apiStyle": "not-openai-compatible",
							}
							v, _ := structpb.NewValue(apiMap)
							return v
						}(),
					},
				}},
			},
			want: false,
		},
		{
			name: "with api segment, and apiStyle is openai-compatible",
			args: args{
				provider: &providerpb.ServiceProvider{Metadata: &metadatapb.Metadata{
					Public: map[string]*structpb.Value{
						"api": func() *structpb.Value {
							apiMap := map[string]any{
								"apiStyle": string(api_style.APIStyleOpenAICompatible),
							}
							v, _ := structpb.NewValue(apiMap)
							return v
						}(),
					},
				}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, CheckIsOpenAICompatibleByProvider(tt.args.provider), "CheckIsOpenAICompatibleByProvider(%v)", tt.args.provider)
		})
	}
}
