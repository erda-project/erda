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

package force_stream_usage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	httperrorutil "github.com/erda-project/erda/pkg/http/httputil"
)

const Name = "force-stream-usage"

var _ filter_define.ProxyRequestRewriter = (*ForceStreamUsage)(nil)

func init() {
	filter_define.RegisterFilterCreator(Name, Creator)
}

type ForceStreamUsage struct{}

var Creator filter_define.RequestRewriterCreator = func(_ string, _ json.RawMessage) filter_define.ProxyRequestRewriter {
	return &ForceStreamUsage{}
}

func (f *ForceStreamUsage) OnProxyRequest(pr *httputil.ProxyRequest) error {
	if !ctxhelper.MustGetIsStream(pr.Out.Context()) {
		return nil
	}

	if !strings.HasPrefix(pr.Out.Header.Get(httperrorutil.HeaderKeyContentType), string(httperrorutil.ApplicationJson)) {
		return nil
	}

	// only set stream_options for /v1/chat/completions
	if !strings.HasPrefix(pr.Out.URL.Path, vars.RequestPathPrefixV1ChatCompletions) {
		return nil
	}

	bodyCopy, err := body_util.SmartCloneBody(&pr.Out.Body, body_util.MaxSample)
	if err != nil {
		return fmt.Errorf("failed to clone request body: %w", err)
	}

	var jsonBody map[string]any
	decoder := json.NewDecoder(bodyCopy)
	if err := decoder.Decode(&jsonBody); err != nil {
		if errors.Is(err, io.EOF) {
			jsonBody = make(map[string]any)
		} else {
			return fmt.Errorf("failed to decode request body: %w", err)
		}
	}
	if jsonBody == nil {
		jsonBody = make(map[string]any)
	}

	streamOptions := ensureStringAnyMap(jsonBody["stream_options"])
	streamOptions["include_usage"] = true
	jsonBody["stream_options"] = streamOptions

	buf, err := json.Marshal(jsonBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	if err := body_util.SetBody(pr.Out, buf); err != nil {
		return fmt.Errorf("failed to set request body: %w", err)
	}

	return nil
}

func ensureStringAnyMap(v any) map[string]any {
	if v == nil {
		return make(map[string]any)
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return make(map[string]any)
}
