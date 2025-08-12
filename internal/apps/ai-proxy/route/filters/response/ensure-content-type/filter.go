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

package ensure_content_type

import (
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

type Filter struct {
	filter_define.PassThroughResponseModifier
}

var (
	_ filter_define.ProxyResponseModifier = (*Filter)(nil)
)

var ResponseModifierCreator filter_define.ResponseModifierCreator = func(name string, _ json.RawMessage) filter_define.ProxyResponseModifier {
	return &Filter{}
}

func init() {
	filter_define.RegisterFilterCreator("ensure-content-type", ResponseModifierCreator)
}

func (f *Filter) OnHeaders(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		if ctxhelper.MustGetIsStream(resp.Request.Context()) {
			resp.Header.Set("Content-Type", "text/event-stream")
		}
	}
	return nil
}
