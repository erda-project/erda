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
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"
)

func TestConverter_OnProxyRequest_GlobalQueryToEachData(t *testing.T) {
	body := `{
		"model":"doubao-seed-rerank",
		"query":"展示一张系统架构图",
		"documents":[
			{"title":"系统架构设计文档","image":"https://example.com/arch.png"},
			{"title":"会议纪要","content":"讨论了下个季度的 OKR。"}
		],
		"top_n":2,
		"rerank_instruction":"Whether the document answers the query"
	}`

	pr := makeProxyRequest(t, body)
	if err := (&Converter{}).OnProxyRequest(pr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out vikingRerankRequest
	if err := json.NewDecoder(pr.Out.Body).Decode(&out); err != nil {
		t.Fatalf("decode out body: %v", err)
	}
	if out.RerankModel != "doubao-seed-rerank" {
		t.Fatalf("unexpected model: %s", out.RerankModel)
	}
	if out.TopK == nil || *out.TopK != 2 {
		t.Fatalf("unexpected top_k: %#v", out.TopK)
	}
	if len(out.Datas) != 2 {
		t.Fatalf("unexpected datas length: %d", len(out.Datas))
	}
	if q, ok := out.Datas[0].Query.(string); !ok || q != "展示一张系统架构图" {
		t.Fatalf("unexpected data[0].query: %#v", out.Datas[0].Query)
	}
	if q, ok := out.Datas[1].Query.(string); !ok || q != "展示一张系统架构图" {
		t.Fatalf("unexpected data[1].query: %#v", out.Datas[1].Query)
	}
	if out.RerankInstruction == "" {
		t.Fatalf("rerank_instruction should be preserved")
	}
}

func TestConverter_OnProxyRequest_DocQueryObjectPreferred(t *testing.T) {
	body := `{
		"model":"doubao-seed-rerank",
		"query":"全局查询",
		"documents":[
			{
				"query":{"text":"局部查询","image":"https://example.com/q.png"},
				"title":"doc",
				"content":"c"
			}
		]
	}`

	pr := makeProxyRequest(t, body)
	if err := (&Converter{}).OnProxyRequest(pr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out vikingRerankRequest
	if err := json.NewDecoder(pr.Out.Body).Decode(&out); err != nil {
		t.Fatalf("decode out body: %v", err)
	}
	if len(out.Datas) != 1 {
		t.Fatalf("unexpected datas length: %d", len(out.Datas))
	}
	qMap, ok := out.Datas[0].Query.(map[string]any)
	if !ok {
		t.Fatalf("expected query object, got: %#v", out.Datas[0].Query)
	}
	if qMap["text"] != "局部查询" {
		t.Fatalf("unexpected query.text: %#v", qMap["text"])
	}
}

func TestConverter_OnProxyRequest_ImageOnlyDocumentAllowed(t *testing.T) {
	body := `{
		"model":"doubao-seed-rerank",
		"query":"展示一张系统架构图",
		"documents":[
			{"image":"https://example.com/a.png"}
		]
	}`
	pr := makeProxyRequest(t, body)
	if err := (&Converter{}).OnProxyRequest(pr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConverter_OnProxyRequest_ValidateFailures(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "missing query",
			body: `{"model":"doubao-seed-rerank","documents":[{"content":"x"}]}`,
			want: "query is required",
		},
		{
			name: "no content_or_image",
			body: `{"model":"doubao-seed-rerank","query":"q","documents":[{"title":"x"}]}`,
			want: "documents should contain at least one content/image",
		},
		{
			name: "empty documents",
			body: `{"model":"doubao-seed-rerank","query":"q","documents":[]}`,
			want: "documents is required",
		},
		{
			name: "invalid document type",
			body: `{"model":"doubao-seed-rerank","query":"q","documents":[123,{"content":"ok"}]}`,
			want: "documents[0] is invalid",
		},
		{
			name: "empty string document",
			body: `{"model":"doubao-seed-rerank","query":"q","documents":[" ",{"content":"ok"}]}`,
			want: "documents[0] is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := makeProxyRequest(t, tt.body)
			err := (&Converter{}).OnProxyRequest(pr)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error contains %q, got %v", tt.want, err)
			}
		})
	}
}

func TestConverter_OnProxyRequest_DocumentsLimit(t *testing.T) {
	items := make([]string, 0, maxVikingDatas+1)
	for i := 0; i < maxVikingDatas+1; i++ {
		items = append(items, `{"content":"x"}`)
	}
	body := `{"model":"doubao-seed-rerank","query":"q","documents":[` + strings.Join(items, ",") + `]}`
	pr := makeProxyRequest(t, body)
	err := (&Converter{}).OnProxyRequest(pr)
	if err == nil || !strings.Contains(err.Error(), "should not exceed") {
		t.Fatalf("expected documents limit error, got: %v", err)
	}
}

func makeProxyRequest(t *testing.T, body string) *httputil.ProxyRequest {
	t.Helper()
	in := httptest.NewRequest("POST", "http://example/v1/reranks", strings.NewReader(body))
	in.Header.Set("Content-Type", "application/json")
	out := in.Clone(in.Context())
	return &httputil.ProxyRequest{
		In:  in,
		Out: out,
	}
}
