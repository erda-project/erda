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

import "testing"

func TestCodingAgentItem_MatchCodexPrompt(t *testing.T) {
	matcher, ok := any(codingAgentItem{}).(PromptMatcher)
	if !ok {
		t.Fatal("expected coding-agent item to implement PromptMatcher")
	}
	if !matcher.MatchPrompt(codexSystemPromptHint + "\nYou are running in a coding environment.") {
		t.Fatal("expected codex prompt prefix to match coding-agent")
	}
}

func TestCodingAgentItem_MatchCodexPromptAfterLeadingWhitespace(t *testing.T) {
	matcher := any(codingAgentItem{}).(PromptMatcher)
	if !matcher.MatchPrompt("\n \t" + codexSystemPromptHint + "\nYou are running in a coding environment.") {
		t.Fatal("expected codex prompt prefix to match coding-agent after trimming leading whitespace")
	}
}

func TestCodingAgentItem_IgnoreCodexPromptWithoutPrefix(t *testing.T) {
	matcher := any(codingAgentItem{}).(PromptMatcher)
	if matcher.MatchPrompt("Tooling follows below.\n" + codexSystemPromptHint) {
		t.Fatal("expected non-prefixed codex prompt not to match coding-agent")
	}
}

func TestCodingAgentItem_MatchCodexHeaderContainingCodexInKey(t *testing.T) {
	matcher, ok := any(codingAgentItem{}).(HeaderMatcher)
	if !ok {
		t.Fatal("expected coding-agent item to implement HeaderMatcher")
	}
	if !matcher.MatchHeader("X-Codex-Originator", "cli") {
		t.Fatal("expected codex header key to match coding-agent")
	}
}

func TestCodingAgentItem_MatchCodexHeaderContainingCodexInValue(t *testing.T) {
	matcher := any(codingAgentItem{}).(HeaderMatcher)
	if !matcher.MatchHeader("User-Agent", "codex_cli_rs/0.116.0") {
		t.Fatal("expected codex header value to match coding-agent")
	}
}

func TestCodingAgentItem_IgnoreHeaderWithoutCodex(t *testing.T) {
	matcher := any(codingAgentItem{}).(HeaderMatcher)
	if matcher.MatchHeader("User-Agent", "openai-cli/1.2.3") {
		t.Fatal("expected header without codex not to match coding-agent")
	}
}

func TestCodingAgentItem_MatchCursorPrompt(t *testing.T) {
	matcher := any(codingAgentItem{}).(PromptMatcher)
	if !matcher.MatchPrompt(cursorSystemPromptHint + "\nYou can help with editing and running commands.") {
		t.Fatal("expected cursor prompt prefix to match coding-agent")
	}
}

func TestCodingAgentItem_MatchCursorMessageGroupText(t *testing.T) {
	matcher, ok := any(codingAgentItem{}).(MessageGroupMatcher)
	if !ok {
		t.Fatal("expected coding-agent item to implement MessageGroupMatcher")
	}
	if !matcher.MatchMessageGroupText(cursorSystemPromptHint + "\nYou can help with editing and running commands.") {
		t.Fatal("expected cursor message-group text prefix to match coding-agent")
	}
}

func TestCodingAgentItem_IgnoreCursorMessageGroupTextWithoutPrefix(t *testing.T) {
	matcher := any(codingAgentItem{}).(MessageGroupMatcher)
	if matcher.MatchMessageGroupText("Tooling follows below.\n" + cursorSystemPromptHint) {
		t.Fatal("expected non-prefixed cursor message-group text not to match coding-agent")
	}
}
