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

package volcengine_ark

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func (f *VolcengineMultimodalEmbeddingResponseConverter) OnBodyChunk(resp *http.Response, chunk []byte, index int64) ([]byte, error) {
	var raw map[string]any
	if err := json.Unmarshal(chunk, &raw); err != nil {
		return chunk, nil
	}
	if !needsConvert(raw) {
		return chunk, nil
	}
	converted := convertToCanonical(resp.Request.Context(), raw)
	out, err := json.Marshal(converted)
	if err != nil {
		return chunk, nil
	}
	return out, nil
}

func needsConvert(raw map[string]any) bool {
	data, ok := raw["data"].(map[string]any)
	if !ok {
		return false
	}
	_, hasEmbedding := data["embedding"]
	return hasEmbedding
}

func convertToCanonical(ctx context.Context, raw map[string]any) map[string]any {
	data := raw["data"].(map[string]any)
	item := map[string]any{
		"index":     0,
		"embedding": data["embedding"],
	}
	if typ := inferSingleInputType(ctx); typ != "" {
		item["type"] = typ
	}
	if v, ok := data["multi_embedding"]; ok {
		item["multi_embedding"] = v
	}
	if v, ok := data["sparse_embedding"]; ok {
		item["sparse_embedding"] = v
	}
	raw["data"] = []any{item}

	if id, ok := raw["id"].(string); ok && strings.TrimSpace(id) != "" {
		raw["request_id"] = id
		delete(raw, "id")
	}
	delete(raw, "object")
	raw["usage"] = normalizeUsage(raw["usage"])

	return raw
}

func inferSingleInputType(ctx context.Context) string {
	reqOut, ok := ctxhelper.GetReverseProxyRequestOutSnapshot(ctx)
	if !ok || reqOut == nil || reqOut.Body == nil {
		return ""
	}
	b, err := io.ReadAll(reqOut.Body)
	if err != nil {
		return ""
	}
	reqOut.Body = io.NopCloser(bytes.NewReader(b))

	var req map[string]any
	if err = json.Unmarshal(b, &req); err != nil {
		return ""
	}
	input, ok := req["input"].([]any)
	if !ok || len(input) == 0 {
		return ""
	}
	seen := map[string]bool{}
	for _, item := range input {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		t, _ := m["type"].(string)
		n := normalizeInputType(t)
		if n == "" {
			continue
		}
		seen[n] = true
	}
	if len(seen) != 1 {
		return ""
	}
	for k := range seen {
		return k
	}
	return ""
}

func normalizeInputType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "text":
		return "text"
	case "image", "image_url":
		return "image"
	case "video", "video_url":
		return "video"
	default:
		return ""
	}
}

func normalizeUsage(raw any) map[string]any {
	usage := map[string]any{
		"total_tokens":  0,
		"input_tokens":  0,
		"output_tokens": 0,
		"input_tokens_details": map[string]any{
			"text_tokens":  0,
			"image_tokens": 0,
			"video_tokens": 0,
		},
	}

	m, ok := raw.(map[string]any)
	if !ok {
		return usage
	}
	total := toInt64(m["total_tokens"])
	input := toInt64(m["prompt_tokens"])
	output := int64(0)
	if total >= input {
		output = total - input
	}
	usage["total_tokens"] = total
	usage["input_tokens"] = input
	usage["output_tokens"] = output

	details, ok := m["prompt_tokens_details"].(map[string]any)
	if !ok {
		return usage
	}
	normalizedDetails := usage["input_tokens_details"].(map[string]any)
	if v, exists := details["text_tokens"]; exists {
		normalizedDetails["text_tokens"] = toInt64(v)
	}
	if v, exists := details["image_tokens"]; exists {
		normalizedDetails["image_tokens"] = toInt64(v)
	}
	if v, exists := details["video_tokens"]; exists {
		normalizedDetails["video_tokens"] = toInt64(v)
	}
	return usage
}

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case float64:
		return int64(x)
	default:
		return 0
	}
}
