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

package aliyun_bailian

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	serviceproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/service_provider/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func TestEnabled(t *testing.T) {
	tests := []struct {
		name           string
		spType         string
		modelPublisher string
		expected       bool
	}{
		{
			name:           "AliyunBailian + Qwen should be enabled",
			spType:         common_types.ServiceProviderTypeAliyunBailian.String(),
			modelPublisher: common_types.ModelPublisherQwen.String(),
			expected:       true,
		},
		{
			name:           "AliyunBailian + Bytedance should be disabled",
			spType:         common_types.ServiceProviderTypeAliyunBailian.String(),
			modelPublisher: common_types.ModelPublisherBytedance.String(),
			expected:       false,
		},
		{
			name:           "VolcengineArk + Qwen should be disabled",
			spType:         common_types.ServiceProviderTypeVolcengineArk.String(),
			modelPublisher: common_types.ModelPublisherQwen.String(),
			expected:       false,
		},
		{
			name:           "VolcengineArk + Bytedance should be disabled",
			spType:         common_types.ServiceProviderTypeVolcengineArk.String(),
			modelPublisher: common_types.ModelPublisherBytedance.String(),
			expected:       false,
		},
		{
			name:           "Other provider should be disabled",
			spType:         "other-provider",
			modelPublisher: common_types.ModelPublisherQwen.String(),
			expected:       false,
		},
		{
			name:           "Other publisher should be disabled",
			spType:         common_types.ServiceProviderTypeAliyunBailian.String(),
			modelPublisher: "other-publisher",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = ctxhelper.InitCtxMapIfNeed(ctx)
			ctxhelper.PutServiceProvider(ctx, &serviceproviderpb.ServiceProvider{
				Type: tt.spType,
			})
			ctxhelper.PutModel(ctx, &modelpb.Model{
				Publisher: tt.modelPublisher,
			})

			result := Enabled(ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}
