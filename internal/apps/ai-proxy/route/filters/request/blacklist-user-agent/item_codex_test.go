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

func TestCodexItem_MatchPrompt(t *testing.T) {
	matcher, ok := any(codexItem{}).(PromptMatcher)
	if !ok {
		t.Fatal("expected codex item to implement PromptMatcher")
	}
	if !matcher.MatchPrompt(codexSystemPromptHint + "\nYou are running in a coding environment.") {
		t.Fatal("expected codex prompt prefix to match")
	}
}

func TestCodexItem_MatchPromptAfterLeadingWhitespace(t *testing.T) {
	matcher := any(codexItem{}).(PromptMatcher)
	if !matcher.MatchPrompt("\n \t" + codexSystemPromptHint + "\nYou are running in a coding environment.") {
		t.Fatal("expected codex prompt prefix to match after trimming leading whitespace")
	}
}

func TestCodexItem_IgnorePromptWithoutPrefix(t *testing.T) {
	matcher := any(codexItem{}).(PromptMatcher)
	if matcher.MatchPrompt("Tooling follows below.\n" + codexSystemPromptHint) {
		t.Fatal("expected non-prefixed codex prompt not to match")
	}
}

func TestCodexItem_MatchHeaderContainingCodexInKey(t *testing.T) {
	matcher, ok := any(codexItem{}).(HeaderMatcher)
	if !ok {
		t.Fatal("expected codex item to implement HeaderMatcher")
	}
	if !matcher.MatchHeader("X-Codex-Originator", "cli") {
		t.Fatal("expected codex header key to match")
	}
}

func TestCodexItem_MatchHeaderContainingCodexInValue(t *testing.T) {
	matcher := any(codexItem{}).(HeaderMatcher)
	if !matcher.MatchHeader("User-Agent", "codex_cli_rs/0.116.0") {
		t.Fatal("expected codex header value to match")
	}
}

func TestCodexItem_IgnoreHeaderWithoutCodex(t *testing.T) {
	matcher := any(codexItem{}).(HeaderMatcher)
	if matcher.MatchHeader("User-Agent", "openai-cli/1.2.3") {
		t.Fatal("expected header without codex not to match")
	}
}
