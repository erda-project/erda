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

package blacklist_user_agent

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
)

func TestCodexItem_MatchMessageGroup(t *testing.T) {
	ctx := newDetectContextForTest()
	ctxhelper.PutMessageGroup(ctx, message.Group{
		RequestedMessages: message.Messages{
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: codexSystemPromptHint + "\nYou are running in a coding environment.",
			},
		},
	})

	matched, source := codexItem{}.Match(ctx)
	if !matched || source != "message_group" {
		t.Fatalf("expected codex message-group match, got matched=%v source=%q", matched, source)
	}
}

func TestCodexItem_MatchRawInstructionsByPrefix(t *testing.T) {
	ctx := newDetectContextForTest()
	body, err := json.Marshal(map[string]any{
		"instructions": codexSystemPromptHint + "\nYou are running in a coding environment.",
	})
	if err != nil {
		t.Fatalf("failed to marshal raw instructions body: %v", err)
	}
	ctxhelper.PutReverseProxyRequestBodyBytes(ctx, body)

	matched, source := codexItem{}.Match(ctx)
	if !matched || source != "request_body.instructions" {
		t.Fatalf("expected codex raw instructions match, got matched=%v source=%q", matched, source)
	}
}

func TestCodexItem_MatchAfterLeadingWhitespace(t *testing.T) {
	ctx := newDetectContextForTest()
	body, err := json.Marshal(map[string]any{
		"instructions": "\n \t" + codexSystemPromptHint + "\nYou are running in a coding environment.",
	})
	if err != nil {
		t.Fatalf("failed to marshal raw instructions body: %v", err)
	}
	ctxhelper.PutReverseProxyRequestBodyBytes(ctx, body)

	matched, source := codexItem{}.Match(ctx)
	if !matched || source != "request_body.instructions" {
		t.Fatalf("expected codex raw instructions match after trimming leading whitespace, got matched=%v source=%q", matched, source)
	}
}

func TestCodexItem_MatchAuditPromptByPrefix(t *testing.T) {
	ctx := newDetectContextForTest()
	audithelper.Note(ctx, "prompt", codexSystemPromptHint+"\nYou are running in a coding environment.")

	matched, source := codexItem{}.Match(ctx)
	if !matched || source != "audit.prompt" {
		t.Fatalf("expected codex audit prompt match, got matched=%v source=%q", matched, source)
	}
}

func TestCodexItem_IgnoreUserMessageContainingPrompt(t *testing.T) {
	ctx := newDetectContextForTest()
	putRawChatRequestBodyForItemTest(t, ctx, []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: codexSystemPromptHint,
		},
	})

	matched, source := codexItem{}.Match(ctx)
	if matched || source != "" {
		t.Fatalf("expected user message not to match codex, got matched=%v source=%q", matched, source)
	}
}

func TestCodexItem_IgnoreSystemMessageContainingPromptButNotPrefixed(t *testing.T) {
	ctx := newDetectContextForTest()
	putRawChatRequestBodyForItemTest(t, ctx, []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Tooling follows below.\n" + codexSystemPromptHint,
		},
	})

	matched, source := codexItem{}.Match(ctx)
	if matched || source != "" {
		t.Fatalf("expected non-prefixed system message not to match codex, got matched=%v source=%q", matched, source)
	}
}

func TestCodexItem_MatchRequestHeaderContainingCodex(t *testing.T) {
	ctx := newDetectContextForTest()
	req := httptest.NewRequest("POST", "http://example.com/v1/chat/completions", nil)
	req.Header.Set("Originator", "codex_cli_rs")
	req.Header.Set("User-Agent", "codex_cli_rs/0.116.0 (Mac OS 26.4.0; arm64) iTerm.app/3.6.9beta1")
	ctxhelper.PutReverseProxyRequestInSnapshot(ctx, req)

	matched, source := codexItem{}.Match(ctx)
	if !matched || source != "request_header" {
		t.Fatalf("expected codex request-header match, got matched=%v source=%q", matched, source)
	}
}

func TestCodexItem_IgnoreRequestHeaderWithoutCodex(t *testing.T) {
	ctx := newDetectContextForTest()
	req := httptest.NewRequest("POST", "http://example.com/v1/chat/completions", nil)
	req.Header.Set("User-Agent", "openai-cli/1.2.3")
	ctxhelper.PutReverseProxyRequestInSnapshot(ctx, req)

	matched, source := codexItem{}.Match(ctx)
	if matched || source != "" {
		t.Fatalf("expected request header without codex not to match, got matched=%v source=%q", matched, source)
	}
}
