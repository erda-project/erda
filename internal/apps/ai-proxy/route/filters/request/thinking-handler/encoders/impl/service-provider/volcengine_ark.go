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
	"fmt"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types/common_types_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/encoders"
	mp "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/encoders/impl/model-publisher"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// VolcengineThinkingEncoder handles Volcengine-specific thinking payloads. It primarily
// exists to decouple provider-specific logic from publisher-specific implementations.
type VolcengineThinkingEncoder struct {
	delegate mp.BytedanceThinkingEncoder
}

func (e *VolcengineThinkingEncoder) CanEncode(ctx context.Context) bool {
	provider := ctxhelper.MustGetServiceProvider(ctx)
	return strings.EqualFold(common_types_util.GetServiceProviderType(provider), common_types.ServiceProviderTypeVolcengineArk.String())
}

func (e *VolcengineThinkingEncoder) Encode(ctx context.Context, ct types.CommonThinking) (map[string]any, error) {
	return e.delegate.Encode(ctx, ct)
}

func (e *VolcengineThinkingEncoder) GetPriority() int {
	return 0 // highest priority to override publisher-based defaults when provider is Volcengine
}

func (e *VolcengineThinkingEncoder) GetName() string {
	return fmt.Sprintf("service_provider: %s", common_types.ServiceProviderTypeAliyunBailian.String())
}

var _ encoders.CommonThinkingEncoder = (*VolcengineThinkingEncoder)(nil)
