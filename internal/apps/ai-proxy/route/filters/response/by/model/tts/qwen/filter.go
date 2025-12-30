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

package qwen

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

func init() {
	filter_define.RegisterFilterCreator("qwen-tts-converter", Creator)
}

type QwenTTSConverter struct {
	filter_define.PassThroughResponseModifier
	buff bytes.Buffer
}

var Creator filter_define.ResponseModifierCreator = func(_ string, _ json.RawMessage) filter_define.ProxyResponseModifier {
	return &QwenTTSConverter{}
}

func (f *QwenTTSConverter) Enable(resp *http.Response) bool {
	model := ctxhelper.MustGetModel(resp.Request.Context())
	return model.Publisher == common_types.ModelPublisherQwen.String()
}
