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

package impl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"

	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/usage/token_usage/extractors/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/vars"
	"github.com/erda-project/erda/pkg/strutil"
)

type OpenAIExtractor struct{}

var _ types.UsageExtractor = (*OpenAIExtractor)(nil)

func (OpenAIExtractor) TryExtract(resp *http.Response, createReq *usagepb.TokenUsageCreateRequest) bool {
	ctx := resp.Request.Context()
	respBodyStr := ctxhelper.MustGetReverseProxyWholeHandledResponseBodyStr(ctx)
	if len(respBodyStr) == 0 {
		return false
	}
	usageMap := extractOpenAIUsage([]byte(respBodyStr), ctxhelper.MustGetIsStream(ctx))
	if usageMap == nil {
		return false
	}

	if v, ok := getUintFromMaps(usageMap, "input_tokens", "prompt_tokens"); ok {
		createReq.InputTokens = v
	}
	if v, ok := getUintFromMaps(usageMap, "output_tokens", "completion_tokens"); ok {
		createReq.OutputTokens = v
	}
	if v, ok := getUintFromMaps(usageMap, "total_tokens"); ok {
		createReq.TotalTokens = v
	}
	if raw, err := json.Marshal(map[string]any{"usage": usageMap}); err == nil {
		createReq.UsageDetails = string(raw)
	}
	return true
}

func extractOpenAIUsage(body []byte, isStream bool) map[string]any {
	if len(body) == 0 {
		return nil
	}
	if isStream {
		return parseOpenAIStreamUsage(body)
	}
	return parseOpenAIJSONUsage(body)
}

func parseOpenAIJSONUsage(body []byte) map[string]any {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	usageMap := findUsageMap(payload)
	if usageMap == nil {
		return nil
	}
	return usageMap
}

func parseOpenAIStreamUsage(body []byte) map[string]any {
	if len(body) == 0 {
		return nil
	}
	const tailLimit = 2000
	if usage := collectUsageFromStream(body, tailLimit); usage != nil {
		return usage
	}
	return collectUsageFromStream(body, len(body))
}

func collectUsageFromStream(body []byte, limit int) map[string]any {
	if limit < len(body) {
		body = body[len(body)-limit:]
	}
	scanner := bufio.NewScanner(bytes.NewReader(body))
	const maxScanTokenSize = 1024 * 1024 // 1MB
	scanner.Buffer(make([]byte, 0, 64*1024), maxScanTokenSize)
	var lines [][]byte
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		lines = append(lines, line)
	}
	strutil.ReverseSlice(lines)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		line = vars.TrimChunkDataPrefix(line)
		if usageMap := parseOpenAIJSONUsage(line); usageMap != nil {
			return usageMap
		}
	}
	return nil
}

func findUsageMap(v any) map[string]any {
	switch val := v.(type) {
	case map[string]any:
		if usageVal, ok := val["usage"]; ok {
			if usageMap, ok := usageVal.(map[string]any); ok {
				return usageMap
			}
		}
		for _, nested := range val {
			if usage := findUsageMap(nested); usage != nil {
				return usage
			}
		}
	case []any:
		for _, item := range val {
			if usage := findUsageMap(item); usage != nil {
				return usage
			}
		}
	}
	return nil
}

func getUintFromMaps(m map[string]any, keys ...string) (uint64, bool) {
	if m == nil {
		return 0, false
	}
	for _, key := range keys {
		if val, ok := getUintFromValue(m[key]); ok {
			return val, true
		}
	}
	return 0, false
}

func getUintFromValue(v any) (uint64, bool) {
	switch val := v.(type) {
	case float64:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case int:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case int64:
		if val < 0 {
			return 0, false
		}
		return uint64(val), true
	case uint64:
		return val, true
	case json.Number:
		i, err := val.Int64()
		if err != nil || i < 0 {
			return 0, false
		}
		return uint64(i), true
	default:
		return 0, false
	}
}
