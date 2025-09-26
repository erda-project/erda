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

package model_publisher

import (
	"context"
	"fmt"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/encoders"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// BytedanceThinkingEncoder handles Bytedance/Doubao thinking encoding
type BytedanceThinkingEncoder struct{}

func (e *BytedanceThinkingEncoder) CanEncode(ctx context.Context) bool {
	model := ctxhelper.MustGetModel(ctx)
	return strings.EqualFold(model.Publisher, string(common_types.ModelPublisherBytedance))
}

func (e *BytedanceThinkingEncoder) Encode(ctx context.Context, ct types.CommonThinking) (map[string]any, error) {
	thinkingObj := make(map[string]any)
	appendBodyMap := map[string]any{types.FieldThinking: thinkingObj}

	switch ct.MustGetMode() {
	case types.ModeOn:
		thinkingObj[types.FieldType] = "enabled"
	case types.ModeOff:
		thinkingObj[types.FieldType] = "disabled"
	case types.ModeAuto:
		thinkingObj[types.FieldType] = "auto"
	}

	return appendBodyMap, nil
}

func (e *BytedanceThinkingEncoder) GetPriority() int {
	return 5 // lower priority
}

func (e *BytedanceThinkingEncoder) GetName() string {
	return fmt.Sprintf("model_publisher: %s", common_types.ModelPublisherBytedance)
}

var _ encoders.CommonThinkingEncoder = (*BytedanceThinkingEncoder)(nil)
