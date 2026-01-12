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
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	aliyun_bailian "github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/by/service-provider/tts/aliyun-bailian"
)

func init() {
	filter_define.RegisterFilterCreator("aliyun-bailian-tts-response-converter", Creator)
}

type BailianTTSConverter struct {
	filter_define.PassThroughResponseModifier

	audioURL string
}

var Creator filter_define.ResponseModifierCreator = func(_ string, _ json.RawMessage) filter_define.ProxyResponseModifier {
	return &BailianTTSConverter{}
}

func (f *BailianTTSConverter) Enable(resp *http.Response) bool {
	return aliyun_bailian.Enabled(resp.Request.Context())
}
