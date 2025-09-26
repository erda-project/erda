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

package service_provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"google.golang.org/protobuf/types/known/structpb"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

func newProviderWithServiceType(t common_types.ServiceProviderType) *modelproviderpb.ModelProvider {
	return &modelproviderpb.ModelProvider{
		Metadata: &metadatapb.Metadata{
			Public: map[string]*structpb.Value{
				"service_provider_type": structpb.NewStringValue(t.String()),
			},
		},
	}
}

func TestBailianThinkingEncoder(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutModelProvider(ctx, newProviderWithServiceType(common_types.ServiceProviderTypeAliyunBailian))

	encoder := &BailianThinkingEncoder{}

	assert.True(t, encoder.CanEncode(ctx))

	result, err := encoder.Encode(ctx, types.CommonThinking{Mode: types.ModePtr(types.ModeOn)})
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{
		types.FieldEnableThinking: true,
		types.FieldThinkingBudget: 1024,
	}, result)
}

func TestBailianThinkingEncoder_CanEncodeVariants(t *testing.T) {
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutModelProvider(ctx, newProviderWithServiceType(common_types.ServiceProviderTypeAliyunBailian))

	encoder := &BailianThinkingEncoder{}

	assert.True(t, encoder.CanEncode(ctx))

	ctx = ctxhelper.InitCtxMapIfNeed(context.Background())
	ctxhelper.PutModelProvider(ctx, &modelproviderpb.ModelProvider{})
	assert.False(t, encoder.CanEncode(ctx))
}
