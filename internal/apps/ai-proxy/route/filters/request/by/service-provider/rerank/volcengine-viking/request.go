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
	"encoding/json"
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
)

type openAIRerankRequest struct {
	Model             string `json:"model"`
	Query             any    `json:"query"`
	Documents         []any  `json:"documents"`
	TopN              *int   `json:"top_n,omitempty"`
	ReturnDocuments   *bool  `json:"return_documents,omitempty"`
	RerankInstruction string `json:"rerank_instruction,omitempty"`
}

type vikingRerankRequest struct {
	RerankModel       string       `json:"rerank_model"`
	Datas             []vikingData `json:"datas"`
	TopK              *int         `json:"top_k,omitempty"`
	RerankInstruction string       `json:"rerank_instruction,omitempty"`
}

type vikingData struct {
	Query   any    `json:"query"`
	Content string `json:"content,omitempty"`
	Image   any    `json:"image,omitempty"`
	Title   string `json:"title,omitempty"`
}

const (
	maxVikingDatas = 200
)

func (f *Converter) OnProxyRequest(pr *httputil.ProxyRequest) error {
	var in openAIRerankRequest
	if err := json.NewDecoder(pr.Out.Body).Decode(&in); err != nil {
		return fmt.Errorf("failed to decode openai rerank request: %w", err)
	}

	if len(in.Documents) == 0 {
		return fmt.Errorf("documents is required for rerank")
	}
	if len(in.Documents) > maxVikingDatas {
		return fmt.Errorf("documents should not exceed %d items", maxVikingDatas)
	}

	globalQuery, _ := normalizeQuery(in.Query)
	datas := make([]vikingData, 0, len(in.Documents))
	hasContentOrImage := false
	for _, doc := range in.Documents {
		item, ok := toVikingData(doc)
		if !ok {
			continue
		}

		query, ok := normalizeQuery(item.Query)
		if !ok {
			query = globalQuery
		}
		if _, ok := normalizeQuery(query); !ok {
			return fmt.Errorf("query is required for rerank")
		}
		item.Query = query

		if strings.TrimSpace(item.Content) != "" || hasNonEmptyImage(item.Image) {
			hasContentOrImage = true
		}
		datas = append(datas, item)
	}
	if len(datas) == 0 {
		return fmt.Errorf("documents is required for rerank")
	}
	if !hasContentOrImage {
		return fmt.Errorf("documents should contain at least one content/image")
	}

	req := vikingRerankRequest{
		RerankModel:       in.Model,
		Datas:             datas,
		TopK:              in.TopN,
		RerankInstruction: strings.TrimSpace(in.RerankInstruction),
	}
	return body_util.SetBody(pr.Out, req)
}

func normalizeQuery(v any) (any, bool) {
	switch t := v.(type) {
	case string:
		s := strings.TrimSpace(t)
		return s, s != ""
	case nil:
		return nil, false
	case map[string]any:
		text := firstNonEmptyString(t["text"])
		if text != "" || hasNonEmptyImage(t["image"]) {
			return t, true
		}
		return nil, false
	default:
		return nil, false
	}
}

func toVikingData(v any) (vikingData, bool) {
	switch t := v.(type) {
	case string:
		if strings.TrimSpace(t) == "" {
			return vikingData{}, false
		}
		return vikingData{Content: t}, true
	case map[string]any:
		content := firstNonEmptyString(t["content"], t["text"], t["document"])
		return vikingData{
			Content: content,
			Image:   normalizeImage(t["image"]),
			Title:   firstNonEmptyString(t["title"]),
			Query:   t["query"],
		}, true
	default:
		return vikingData{}, false
	}
}

func firstNonEmptyString(values ...any) string {
	for _, v := range values {
		s, ok := v.(string)
		if ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func normalizeImage(v any) any {
	switch t := v.(type) {
	case string:
		if strings.TrimSpace(t) == "" {
			return nil
		}
		return t
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, s)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}

func hasNonEmptyImage(v any) bool {
	switch t := normalizeImage(v).(type) {
	case string:
		return strings.TrimSpace(t) != ""
	case []string:
		return len(t) > 0
	default:
		return false
	}
}
