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
	"encoding/json"
	"net/http/httputil"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

func init() {
	filter_define.RegisterFilterCreator("aliyun-bailian-tts-converter", Creator)
}

type BailianTTSConverter struct{}

var Creator filter_define.RequestRewriterCreator = func(name string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &BailianTTSConverter{}
}

func (f *BailianTTSConverter) Enable(pr *httputil.ProxyRequest) bool {
	return Enabled(pr.In.Context())
}

func Enabled(ctx context.Context) bool {
	sp := ctxhelper.MustGetServiceProvider(ctx)
	model := ctxhelper.MustGetModel(ctx)
	if sp.Type != common_types.ServiceProviderTypeAliyunBailian.String() {
		return false
	}
	if model.Publisher != common_types.ModelPublisherQwen.String() {
		return false
	}
	return true
}
