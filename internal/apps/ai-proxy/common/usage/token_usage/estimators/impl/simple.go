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
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/tiktoken-go/tokenizer"
	"github.com/tiktoken-go/tokenizer/codec"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	usagepb "github.com/erda-project/erda-proto-go/apps/aiproxy/usage/token/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/usage/token_usage/estimators/types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

type SimpleEstimator struct{}

var _ types.UsageEstimator = (*SimpleEstimator)(nil)

func (SimpleEstimator) Estimate(resp *http.Response, createReq *usagepb.TokenUsageCreateRequest) bool {
	var inputTokens uint64
	var outputTokens uint64

	ctx := resp.Request.Context()

	// metadata
	var meta metadata.Metadata
	if createReq.Metadata != nil {
		cputil.MustObjJSONTransfer(createReq.Metadata, &meta)
	}
	if meta.Public == nil {
		meta.Public = make(map[string]any)
	}
	meta.Public["estimator"] = "simple-estimator"

	req := ctxhelper.MustGetReverseProxyRequestInSnapshot(ctx)
	if req.Body != nil {
		reqBodyBytes, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(reqBodyBytes))
		var inputTokenizerName string
		inputTokens, inputTokenizerName = estimateTokenCount(ctxhelper.MustGetModel(ctx).Name, reqBodyBytes, true)
		meta.Public["input_tokenizer"] = inputTokenizerName
	}
	respBodyStr := ctxhelper.MustGetReverseProxyWholeHandledResponseBodyStr(ctx)
	if len(respBodyStr) > 0 {
		var outputTokenizerName string
		outputTokens, outputTokenizerName = estimateTokenCount(ctxhelper.MustGetModel(ctx).Name, []byte(respBodyStr), false)
		meta.Public["output_tokenizer"] = outputTokenizerName
	}

	createReq.InputTokens = inputTokens
	createReq.OutputTokens = outputTokens
	createReq.TotalTokens = inputTokens + outputTokens
	cputil.MustObjJSONTransfer(meta, &createReq.Metadata)

	return true
}

func estimateTokenCount(model string, data []byte, isInput bool) (uint64, string) {
	if tokens, name, ok := countTokensFromStructuredJSON(model, data, isInput); ok {
		return tokens, name
	}
	if tokens, name, ok := countTokensFromSSEJSON(model, data, isInput); ok {
		return tokens, name
	}
	return countTokensFallback(model, data)
}

func countTokensFromStructuredJSON(model string, data []byte, isInput bool) (uint64, string, bool) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
		return 0, "", false
	}
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return 0, "", false
	}
	parts := collectTextFragments(payload, isInput)
	if len(parts) == 0 {
		return 0, "", false
	}
	joined := strings.Join(parts, "\n")
	tokens, name := countTokensFallback(model, []byte(joined))
	return tokens, name, true
}

type sseItemText struct {
	builder *strings.Builder
	done    bool
	text    string
	order   int
}

func countTokensFromSSEJSON(model string, data []byte, isInput bool) (uint64, string, bool) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	const maxScanTokenSize = 1024 * 1024 // 1MB
	scanner.Buffer(make([]byte, 0, 64*1024), maxScanTokenSize)
	items := make(map[string]*sseItemText)
	var (
		parts       []string
		order       int
		hasSSELines bool
	)
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		if !bytes.HasPrefix(line, []byte("data:")) {
			continue
		}
		hasSSELines = true
		payload := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
		if len(payload) == 0 || bytes.Equal(payload, []byte("[DONE]")) {
			continue
		}
		decoder := json.NewDecoder(bytes.NewReader(payload))
		decoder.UseNumber()
		var obj any
		if err := decoder.Decode(&obj); err != nil {
			continue
		}
		m, _ := obj.(map[string]any)
		eventType, _ := m["type"].(string)
		order++
		switch {
		case strings.HasSuffix(eventType, ".delta"):
			if deltaStr, ok := m["delta"].(string); ok {
				itemID, _ := m["item_id"].(string)
				if itemID != "" {
					it := ensureItem(items, itemID, order)
					if it.done {
						continue
					}
					if it.builder == nil {
						it.builder = &strings.Builder{}
					}
					it.builder.WriteString(deltaStr)
				} else {
					parts = append(parts, deltaStr)
				}
				continue
			}
		case strings.HasSuffix(eventType, ".done"):
			if textStr, ok := m["text"].(string); ok && textStr != "" {
				itemID, _ := m["item_id"].(string)
				if itemID != "" {
					it := ensureItem(items, itemID, order)
					it.done = true
					it.text = textStr
				} else {
					parts = append(parts, textStr)
				}
				continue
			}
		case strings.HasSuffix(eventType, ".completed"):
			// completed event usually repeats full content, skip to avoid double counting
			continue
		}
		parts = append(parts, collectTextFragments(obj, isInput)...)
	}
	if len(items) > 0 {
		var ordered []*sseItemText
		for _, it := range items {
			ordered = append(ordered, it)
		}
		sort.Slice(ordered, func(i, j int) bool {
			return ordered[i].order < ordered[j].order
		})
		for _, it := range ordered {
			switch {
			case it.done && it.text != "":
				parts = append(parts, it.text)
			case it.builder != nil:
				parts = append(parts, it.builder.String())
			}
		}
	}
	if !hasSSELines || len(parts) == 0 {
		return 0, "", false
	}
	joined := strings.Join(parts, "\n")
	tokens, name := countTokensFallback(model, []byte(joined))
	return tokens, name, true
}

func ensureItem(items map[string]*sseItemText, id string, order int) *sseItemText {
	if it, ok := items[id]; ok {
		return it
	}
	it := &sseItemText{order: order}
	items[id] = it
	return it
}

func collectTextFragments(v any, isInput bool) []string {
	switch val := v.(type) {
	case string:
		return []string{val}
	case []any:
		var result []string
		for _, item := range val {
			result = append(result, collectTextFragments(item, isInput)...)
		}
		return result
	case map[string]any:
		var result []string
		for key, value := range val {
			lowerKey := strings.ToLower(key)
			if shouldCaptureKey(lowerKey, isInput) {
				result = append(result, collectTextFragments(value, isInput)...)
			} else if shouldTraverseKey(lowerKey, isInput) {
				result = append(result, collectTextFragments(value, isInput)...)
			}
		}
		return result
	default:
		return nil
	}
}

func shouldCaptureKey(key string, isInput bool) bool {
	if isInput {
		return inputCaptureKeys[key]
	}
	return outputCaptureKeys[key]
}

func shouldTraverseKey(key string, isInput bool) bool {
	if isInput {
		return inputTraverseKeys[key]
	}
	return outputTraverseKeys[key]
}

var (
	inputCaptureKeys = map[string]bool{
		"prompt":      true,
		"input":       true,
		"inputs":      true,
		"content":     true,
		"text":        true,
		"query":       true,
		"instruction": true,
		"arguments":   true,
		"code":        true,
		"user_prompt": true,
	}
	inputTraverseKeys = map[string]bool{
		"messages": true,
		"contents": true,
		"input":    true,
		"inputs":   true,
		"request":  true,
		"data":     true,
		"payload":  true,
	}
	outputCaptureKeys = map[string]bool{
		"content":    true,
		"text":       true,
		"output":     true,
		"outputs":    true,
		"response":   true,
		"completion": true,
		"result":     true,
		"delta":      true,
	}
	outputTraverseKeys = map[string]bool{
		"choices":  true,
		"message":  true,
		"messages": true,
		"contents": true,
		"response": true,
		"data":     true,
		"payload":  true,
	}
)

func countTokensFallback(model string, data []byte) (uint64, string) {
	if tokens, name := tryCountByOpenaiTokenizer(model, data); name != "" {
		return tokens, "openai: " + name
	}
	return approximateTokenCount(data), "buffer-words"
}

// tryCountByOpenaiTokenizer try to get token count by openai tokenizer.
// return:
// - token count
// - tokenizer name
func tryCountByOpenaiTokenizer(model string, data []byte) (uint64, string) {
	modelCodec, _ := tokenizer.ForModel(tokenizer.Model(model))
	if modelCodec == nil {
		modelCodec = codec.NewO200kBase()
	}
	n, err := modelCodec.Count(string(data))
	if err != nil {
		return 0, ""
	}
	return uint64(n), modelCodec.GetName()
}

func approximateTokenCount(data []byte) uint64 {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Split(bufio.ScanWords)
	var count uint64
	for scanner.Scan() {
		count++
	}
	if count == 0 {
		count = uint64(len(data) / 4)
	}
	if count == 0 {
		count = 1
	}
	return count
}
