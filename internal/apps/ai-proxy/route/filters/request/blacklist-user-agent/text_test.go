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
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestMatchPromptPrefix(t *testing.T) {
	if !matchPromptPrefix("\n \tYou are Codex\nTooling follows below.", "You are Codex") {
		t.Fatal("expected leading whitespace to be trimmed before prefix matching")
	}
	if matchPromptPrefix("Tooling follows below.\nYou are Codex", "You are Codex") {
		t.Fatal("expected non-prefixed content not to match")
	}
}

func TestMatchPromptWithinWindowContains(t *testing.T) {
	if !matchPromptWithinWindowContains("\n \tYou are GPT-5.2 running in the Assistant client, a terminal-based coding assistant.", "terminal-based coding assistant", generalPromptMatchWindow) {
		t.Fatal("expected substring within first 200 chars to match after trimming leading whitespace")
	}
	longPrefix := ""
	for i := 0; i < generalPromptMatchWindow+1; i++ {
		longPrefix += "a"
	}
	if matchPromptWithinWindowContains(longPrefix+"terminal-based coding assistant", "terminal-based coding assistant", generalPromptMatchWindow) {
		t.Fatal("expected substring beyond first 200 chars not to match")
	}
}

func TestChatMessageText(t *testing.T) {
	if got := chatMessageText(openai.ChatCompletionMessage{Content: "plain text"}); got != "plain text" {
		t.Fatalf("expected plain content to be returned when multi-content is empty, got %q", got)
	}

	got := chatMessageText(openai.ChatCompletionMessage{
		Content: "ignored",
		MultiContent: []openai.ChatMessagePart{
			{Type: openai.ChatMessagePartTypeText, Text: "line 1"},
			{Type: openai.ChatMessagePartTypeImageURL},
			{Type: openai.ChatMessagePartTypeText, Text: "line 2"},
			{Type: openai.ChatMessagePartTypeText},
		},
	})
	if got != "line 1\nline 2" {
		t.Fatalf("expected multi-content text parts to be joined with newlines, got %q", got)
	}
}
