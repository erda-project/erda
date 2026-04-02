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

func TestGeneralItem_MatchHeaderByConfiguredRules(t *testing.T) {
	t.Cleanup(func() { SetGeneralRules("", "") })
	SetGeneralRules("claude code ;;; opencode", "")

	matcher, ok := any(generalItem{}).(HeaderMatcher)
	if !ok {
		t.Fatal("expected general item to implement HeaderMatcher")
	}
	if !matcher.MatchHeader("User-Agent", "claude code/1.0") {
		t.Fatal("expected general item to match configured header value")
	}
}

func TestGeneralItem_MatchPromptByConfiguredWindowContains(t *testing.T) {
	t.Cleanup(func() { SetGeneralRules("", "") })
	SetGeneralRules("", "terminal-based coding assistant ;;; you are opencode")

	matcher, ok := any(generalItem{}).(PromptMatcher)
	if !ok {
		t.Fatal("expected general item to implement PromptMatcher")
	}
	if !matcher.MatchPrompt("You are GPT-5.2 running in the Assistant client, a terminal-based coding assistant.") {
		t.Fatal("expected general item to match configured prompt substring within first 200 chars")
	}
}

func TestGeneralItem_IgnorePromptOutsideConfiguredWindow(t *testing.T) {
	t.Cleanup(func() { SetGeneralRules("", "") })
	SetGeneralRules("", "target-window-marker")

	matcher := any(generalItem{}).(PromptMatcher)
	longPrefix := ""
	for i := 0; i < 201; i++ {
		longPrefix += "a"
	}
	if matcher.MatchPrompt(longPrefix + "target-window-marker") {
		t.Fatal("expected general item not to match prompt substring beyond first 200 chars")
	}
}

func TestGeneralItem_MatchMessageGroupTextByConfiguredWindowContains(t *testing.T) {
	t.Cleanup(func() { SetGeneralRules("", "") })
	SetGeneralRules("", "assistant client")

	matcher, ok := any(generalItem{}).(MessageGroupMatcher)
	if !ok {
		t.Fatal("expected general item to implement MessageGroupMatcher")
	}
	if !matcher.MatchMessageGroupText("You are GPT-5.1 running in the Assistant client, a terminal-based coding assistant.") {
		t.Fatal("expected general item to match configured message-group substring within first 200 chars")
	}
}

func TestGeneralItem_IgnoreSignalsWhenNoConfiguredRules(t *testing.T) {
	t.Cleanup(func() { SetGeneralRules("", "") })
	SetGeneralRules("", "")

	headerMatcher := any(generalItem{}).(HeaderMatcher)
	if headerMatcher.MatchHeader("User-Agent", "claude code/1.0") {
		t.Fatal("expected general item not to match header when config is empty")
	}
	promptMatcher := any(generalItem{}).(PromptMatcher)
	if promptMatcher.MatchPrompt("You are Claude Code.") {
		t.Fatal("expected general item not to match prompt when config is empty")
	}
	messageMatcher := any(generalItem{}).(MessageGroupMatcher)
	if messageMatcher.MatchMessageGroupText("You are OpenCode, running in this environment.") {
		t.Fatal("expected general item not to match message-group text when config is empty")
	}
}
