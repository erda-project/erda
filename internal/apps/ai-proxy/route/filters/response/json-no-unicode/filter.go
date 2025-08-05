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

package json_no_unicode

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

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
	filter_define.RegisterFilterCreator("json-no-unicode", ResponseModifierCreator)
}

func (f *Filter) OnBodyChunk(resp *http.Response, chunk []byte) ([]byte, error) {
	// Only handle non-streaming responses with Content-Type as application/json
	if ctxhelper.GetIsStream(resp.Request.Context()) {
		return chunk, nil
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		return chunk, nil
	}

	// Try to parse JSON
	var m map[string]any
	if err := json.Unmarshal(chunk, &m); err != nil {
		// If not valid JSON, return original content directly
		return chunk, nil
	}

	// Re-encode with HTML escaping disabled
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(&m); err != nil {
		// Encoding failed, return original content
		return chunk, nil
	}

	// Remove newline character added by encoder.Encode
	result := buf.Bytes()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return result, nil
}
