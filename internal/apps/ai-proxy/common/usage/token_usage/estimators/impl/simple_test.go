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
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestTryCountByOpenaiTokenizer(t *testing.T) {
	data := []byte("hello world from erda")
	tokens, name := tryCountByOpenaiTokenizer("gpt-4o-mini", data)

	if tokens == 0 {
		t.Fatalf("expected positive token count, got %d", tokens)
	}
	if name == "" {
		t.Fatalf("expected tokenizer name, got empty string")
	}
}

func TestEstimateTokenCountUsesOpenAITokenizerWhenAvailable(t *testing.T) {
	data := []byte("simple content for counting")
	tokens, tokenizerName := estimateTokenCount("unknown-model", data, true)

	if tokens == 0 {
		t.Fatalf("expected positive token count, got %d", tokens)
	}
	if !strings.HasPrefix(tokenizerName, "openai:") {
		t.Fatalf("expected tokenizer name to have openai prefix, got %q", tokenizerName)
	}
}

func TestEstimateTokenCountFromJSONExtractsRelevantFields(t *testing.T) {
	prompt := "hello world"
	body := []byte(fmt.Sprintf(`{"prompt":%q,"metadata":"ignored"}`, prompt))

	expectedTokens, _ := countTokensFallback("unknown-model", []byte(prompt))
	tokens, _ := estimateTokenCount("unknown-model", body, true)

	if tokens != expectedTokens {
		t.Fatalf("expected tokens %d, got %d", expectedTokens, tokens)
	}
}

func TestCountTokensFromStructuredJSON_InputMessages(t *testing.T) {
	body := []byte(`{"messages":[{"content":"hello world"},{"content":"foo bar"}],"metadata":{"ignored":true}}`)
	expectedJoined := "hello world\nfoo bar"
	expectedTokens, expectedName := countTokensFallback("unknown-model", []byte(expectedJoined))

	tokens, name, ok := countTokensFromStructuredJSON("unknown-model", body, true)
	if !ok {
		t.Fatalf("expected structured JSON to be handled")
	}
	if tokens != expectedTokens {
		t.Fatalf("expected tokens %d, got %d", expectedTokens, tokens)
	}
	if name != expectedName {
		t.Fatalf("expected tokenizer name %q, got %q", expectedName, name)
	}
}

func TestCountTokensFromStructuredJSON_OutputChoices(t *testing.T) {
	body := []byte(`{"choices":[{"message":{"content":"answer one"}},{"message":{"content":"answer two"}}],"other":"ignored"}`)
	expectedJoined := "answer one\nanswer two"
	expectedTokens, expectedName := countTokensFallback("unknown-model", []byte(expectedJoined))

	tokens, name, ok := countTokensFromStructuredJSON("unknown-model", body, false)
	if !ok {
		t.Fatalf("expected structured JSON output to be handled")
	}
	if tokens != expectedTokens {
		t.Fatalf("expected tokens %d, got %d", expectedTokens, tokens)
	}
	if name != expectedName {
		t.Fatalf("expected tokenizer name %q, got %q", expectedName, name)
	}
}

func TestEstimateTokenCountFallbackForPlainText(t *testing.T) {
	data := []byte("plain text without json markers")
	expectedTokens, expectedName := countTokensFallback("unknown-model", data)

	tokens, name := estimateTokenCount("unknown-model", data, true)
	if tokens != expectedTokens {
		t.Fatalf("expected tokens %d, got %d", expectedTokens, tokens)
	}
	if name != expectedName {
		t.Fatalf("expected tokenizer name %q, got %q", expectedName, name)
	}
}

const sampleDoubaoResponseText = `杭州，这座被誉为“人间天堂”的千年古城，位于中国东南沿海的浙江省北部，是长三角南翼的中心城市，更是一座自然与人文交织、传统与现代共生的魅力之城。
若问何时来最好？或许是烟花三月，或许是桂子飘香，又或许，就是此刻。`

const sampleDoubaoSummaryText = `
用户让我介绍杭州，我需要先确定介绍的结构和内容。首先，杭州是浙江省的省会，有“人间天堂”的美誉，所以开头可以先点出这个称号，然后从多个方面展开。
最后总结杭州的特点，古今交融，自然与人文结合，适合旅游和生活。需要注意语言要流畅，信息准确，重点突出。`

func TestEstimateTokenCountForDoubaoStream(t *testing.T) {
	textJSON, err := json.Marshal(sampleDoubaoResponseText)
	if err != nil {
		t.Fatalf("failed to marshal sample text: %v", err)
	}
	summaryJSON, err := json.Marshal(sampleDoubaoSummaryText)
	if err != nil {
		t.Fatalf("failed to marshal summary text: %v", err)
	}
	body := fmt.Sprintf(`event: response.output_text.delta
data: {"type":"response.output_text.delta","content_index":0,"delta":"就是","item_id":"msg_02176041337904600000000000000000000ffffac15bca5f1eb87","output_index":1,"sequence_number":1466}

event: response.output_text.delta
data: {"type":"response.output_text.delta","content_index":0,"delta":"此刻","item_id":"msg_02176041337904600000000000000000000ffffac15bca5f1eb87","output_index":1,"sequence_number":1467}

event: response.output_text.delta
data: {"type":"response.output_text.delta","content_index":0,"delta":"。","item_id":"msg_02176041337904600000000000000000000ffffac15bca5f1eb87","output_index":1,"sequence_number":1468}

event: response.output_text.done
data: {"type":"response.output_text.done","content_index":0,"item_id":"msg_02176041337904600000000000000000000ffffac15bca5f1eb87","output_index":1,"text":%s,"sequence_number":1469}

event: response.content_part.done
data: {"type":"response.content_part.done","content_index":0,"item_id":"msg_02176041337904600000000000000000000ffffac15bca5f1eb87","output_index":1,"part":{"type":"output_text","text":%s},"sequence_number":1470}

event: response.output_item.done
data: {"type":"response.output_item.done","output_index":1,"item":{"type":"message","role":"assistant","content":[{"type":"output_text","text":%s}],"status":"completed","id":"msg_02176041337904600000000000000000000ffffac15bca5f1eb87"},"sequence_number":1471}

event: response.completed
data: {"type":"response.completed","response":{"created_at":1760413368,"id":"resp_0217604133681540b5206b6c0f3bb2682a16c13f4ce95f35a6513","output":[{"id":"rs_02176041336846700000000000000000000ffffac15bca53329c0","type":"reasoning","summary":[{"type":"summary_text","text":%s}],"status":"completed"},{"type":"message","role":"assistant","content":[{"type":"output_text","text":%s}],"status":"completed","id":"msg_02176041337904600000000000000000000ffffac15bca5f1eb87"}],"usage":{"input_tokens":91,"output_tokens":1461,"total_tokens":1552}},"sequence_number":1472}

data: [DONE]
`, string(textJSON), string(textJSON), string(textJSON), string(summaryJSON), string(textJSON))

	expectedTokens, expectedName := countTokensFallback("unknown-model", []byte(sampleDoubaoResponseText))
	tokens, name := estimateTokenCount("unknown-model", []byte(body), false)

	if tokens != expectedTokens {
		t.Fatalf("expected tokens %d, got %d", expectedTokens, tokens)
	}
	if name != expectedName {
		t.Fatalf("expected tokenizer name %q, got %q", expectedName, name)
	}
}
