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

package dashscope_director

import (
	"context"
	"fmt"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
	"github.com/erda-project/erda/pkg/reverseproxy"
)

func (f *DashScopeDirector) OnResponseChunk(ctx context.Context, infor reverseproxy.HttpInfor, w reverseproxy.Writer, chunk []byte) (signal reverseproxy.Signal, err error) {
	model := ctxhelper.MustGetModel(ctx)
	modelMeta := getModelMeta(model)

	// switch by model name
	switch modelMeta.Public.ModelName {
	case metadata.AliyunDashScopeModelNameQwenLong:
		return f.DefaultResponseFilter.OnResponseChunk(ctx, infor, w, chunk)
	default:
		return reverseproxy.Intercept, fmt.Errorf("unsupported model: %s", model.Name)
	}
}
