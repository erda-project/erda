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

package volcengine_viking

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/common_types"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filter_define"
)

func init() {
	filter_define.RegisterFilterCreator("volcengine-viking-rerank-response-converter", ResponseModifierCreator)
}

type ResponseConverter struct {
	filter_define.PassThroughResponseModifier
}

var ResponseModifierCreator filter_define.ResponseModifierCreator = func(_ string, _ json.RawMessage) filter_define.ProxyResponseModifier {
	return &ResponseConverter{}
}

func (f *ResponseConverter) Enable(resp *http.Response) bool {
	return enabled(resp.Request.Context())
}

func enabled(ctx context.Context) bool {
	sp := ctxhelper.MustGetServiceProvider(ctx)
	return sp.Type == common_types.ServiceProviderTypeVolcengineViking.String()
}

func (f *ResponseConverter) OnBodyChunk(resp *http.Response, chunk []byte, _ int64) ([]byte, error) {
	var in vikingResponse
	if err := json.Unmarshal(chunk, &in); err != nil || len(in.Data.Scores) == 0 {
		return chunk, nil
	}

	results := make([]rerankResult, 0, len(in.Data.Scores))
	for i, score := range in.Data.Scores {
		results = append(results, rerankResult{
			Index:          i,
			RelevanceScore: score,
		})
	}
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].RelevanceScore > results[j].RelevanceScore
	})

	out := unifiedRerankResponse{
		Output: rerankOutput{Results: results},
		Usage: rerankUsage{
			TotalTokens: extractTotalTokens(in.Data.TokenUsage),
		},
	}
	if b, err := json.Marshal(out); err == nil {
		return b, nil
	}
	return chunk, nil
}

type vikingResponse struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Scores     []float64 `json:"scores"`
		TokenUsage any       `json:"token_usage"`
	} `json:"data"`
}

type unifiedRerankResponse struct {
	Output rerankOutput `json:"output"`
	Usage  rerankUsage  `json:"usage,omitempty"`
}

type rerankOutput struct {
	Results []rerankResult `json:"results"`
}

type rerankResult struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type rerankUsage struct {
	TotalTokens int64 `json:"total_tokens,omitempty"`
}

func extractTotalTokens(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case map[string]any:
		if total, ok := t["total_tokens"]; ok {
			return extractTotalTokens(total)
		}
		if total, ok := t["totalTokens"]; ok {
			return extractTotalTokens(total)
		}
	}
	return 0
}
