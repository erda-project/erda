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
	"testing"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/audit/audithelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/message"
)

func TestCursorItem_MatchMessageGroup(t *testing.T) {
	ctx := newDetectContextForTest()
	ctxhelper.PutMessageGroup(ctx, message.Group{
		RequestedMessages: message.Messages{
			openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: cursorSystemPromptHint + "\nYou can help with editing and running commands.",
			},
		},
	})

	matched, source := cursorItem{}.Match(ctx)
	if !matched || source != "message_group" {
		t.Fatalf("expected cursor message-group match, got matched=%v source=%q", matched, source)
	}
}

func TestCursorItem_MatchRawInstructionsByPrefix(t *testing.T) {
	ctx := newDetectContextForTest()
	body, err := json.Marshal(map[string]any{
		"instructions": cursorSystemPromptHint + "\nYou can help with editing and running commands.",
	})
	if err != nil {
		t.Fatalf("failed to marshal raw instructions body: %v", err)
	}
	ctxhelper.PutReverseProxyRequestBodyBytes(ctx, body)

	matched, source := cursorItem{}.Match(ctx)
	if !matched || source != "request_body.instructions" {
		t.Fatalf("expected cursor raw instructions match, got matched=%v source=%q", matched, source)
	}
}

func TestCursorItem_MatchAfterLeadingWhitespace(t *testing.T) {
	ctx := newDetectContextForTest()
	body, err := json.Marshal(map[string]any{
		"instructions": "\n \t" + cursorSystemPromptHint + "\nYou can help with editing and running commands.",
	})
	if err != nil {
		t.Fatalf("failed to marshal raw instructions body: %v", err)
	}
	ctxhelper.PutReverseProxyRequestBodyBytes(ctx, body)

	matched, source := cursorItem{}.Match(ctx)
	if !matched || source != "request_body.instructions" {
		t.Fatalf("expected cursor raw instructions match after trimming leading whitespace, got matched=%v source=%q", matched, source)
	}
}

func TestCursorItem_MatchAuditPromptByPrefix(t *testing.T) {
	ctx := newDetectContextForTest()
	audithelper.Note(ctx, "prompt", cursorSystemPromptHint+"\nYou can help with editing and running commands.")

	matched, source := cursorItem{}.Match(ctx)
	if !matched || source != "audit.prompt" {
		t.Fatalf("expected cursor audit prompt match, got matched=%v source=%q", matched, source)
	}
}

func TestCursorItem_IgnoreUserMessageContainingPrompt(t *testing.T) {
	ctx := newDetectContextForTest()
	putRawChatRequestBodyForItemTest(t, ctx, []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: cursorSystemPromptHint,
		},
	})

	matched, source := cursorItem{}.Match(ctx)
	if matched || source != "" {
		t.Fatalf("expected user message not to match cursor, got matched=%v source=%q", matched, source)
	}
}

func TestCursorItem_IgnoreSystemMessageContainingPromptButNotPrefixed(t *testing.T) {
	ctx := newDetectContextForTest()
	putRawChatRequestBodyForItemTest(t, ctx, []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Tooling follows below.\n" + cursorSystemPromptHint,
		},
	})

	matched, source := cursorItem{}.Match(ctx)
	if matched || source != "" {
		t.Fatalf("expected non-prefixed system message not to match cursor, got matched=%v source=%q", matched, source)
	}
}
