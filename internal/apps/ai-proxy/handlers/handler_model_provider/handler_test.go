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

package handler_model_provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	"github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func TestDesensitizeProvider(t *testing.T) {
	type want struct {
		apiKey      string
		secretIsNil bool
	}

	tests := []struct {
		name         string
		setupCtx     func(context.Context)
		buildPayload func() *pb.ModelProvider
		want         want
	}{
		{
			name:     "non-admin hides sensitive fields",
			setupCtx: nil,
			buildPayload: func() *pb.ModelProvider {
				return &pb.ModelProvider{
					ApiKey: "should-hide",
					Metadata: &metadatapb.Metadata{
						Secret: map[string]*structpb.Value{
							"token": structpb.NewStringValue("value"),
						},
					},
				}
			},
			want: want{
				apiKey:      "",
				secretIsNil: true,
			},
		},
		{
			name: "admin keeps sensitive fields",
			setupCtx: func(ctx context.Context) {
				ctxhelper.PutIsAdmin(ctx, true)
			},
			buildPayload: func() *pb.ModelProvider {
				return &pb.ModelProvider{
					ApiKey: "keep-me",
					Metadata: &metadatapb.Metadata{
						Secret: map[string]*structpb.Value{
							"token": structpb.NewStringValue("value"),
						},
					},
				}
			},
			want: want{
				apiKey:      "keep-me",
				secretIsNil: false,
			},
		},
		{
			name: "client owned provider keeps sensitive fields",
			setupCtx: func(ctx context.Context) {
				ctxhelper.PutClient(ctx, &clientpb.Client{Id: "client-1"})
			},
			buildPayload: func() *pb.ModelProvider {
				return &pb.ModelProvider{
					ClientId: "client-1",
					ApiKey:   "keep-me",
					Metadata: &metadatapb.Metadata{
						Secret: map[string]*structpb.Value{
							"token": structpb.NewStringValue("value"),
						},
					},
				}
			},
			want: want{
				apiKey:      "keep-me",
				secretIsNil: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
			if tt.setupCtx != nil {
				tt.setupCtx(ctx)
			}

			provider := tt.buildPayload()
			desensitizeProvider(ctx, provider)

			assert.Equal(t, tt.want.apiKey, provider.ApiKey)
			if provider.Metadata != nil {
				if tt.want.secretIsNil {
					assert.Nil(t, provider.Metadata.Secret)
				} else {
					assert.NotNil(t, provider.Metadata.Secret)
				}
			}
		})
	}
}
