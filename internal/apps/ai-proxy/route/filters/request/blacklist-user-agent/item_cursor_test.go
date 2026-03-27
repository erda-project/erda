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

func TestCursorItem_MatchPrompt(t *testing.T) {
	matcher, ok := any(cursorItem{}).(PromptMatcher)
	if !ok {
		t.Fatal("expected cursor item to implement PromptMatcher")
	}
	if !matcher.MatchPrompt(cursorSystemPromptHint + "\nYou can help with editing and running commands.") {
		t.Fatal("expected cursor prompt prefix to match")
	}
}

func TestCursorItem_MatchPromptAfterLeadingWhitespace(t *testing.T) {
	matcher := any(cursorItem{}).(PromptMatcher)
	if !matcher.MatchPrompt("\n \t" + cursorSystemPromptHint + "\nYou can help with editing and running commands.") {
		t.Fatal("expected cursor prompt prefix to match after trimming leading whitespace")
	}
}

func TestCursorItem_IgnorePromptWithoutPrefix(t *testing.T) {
	matcher := any(cursorItem{}).(PromptMatcher)
	if matcher.MatchPrompt("Tooling follows below.\n" + cursorSystemPromptHint) {
		t.Fatal("expected non-prefixed cursor prompt not to match")
	}
}

func TestCursorItem_MatchMessageGroupText(t *testing.T) {
	matcher, ok := any(cursorItem{}).(MessageGroupMatcher)
	if !ok {
		t.Fatal("expected cursor item to implement MessageGroupMatcher")
	}
	if !matcher.MatchMessageGroupText(cursorSystemPromptHint + "\nYou can help with editing and running commands.") {
		t.Fatal("expected cursor message-group text prefix to match")
	}
}

func TestCursorItem_IgnoreMessageGroupTextWithoutPrefix(t *testing.T) {
	matcher := any(cursorItem{}).(MessageGroupMatcher)
	if matcher.MatchMessageGroupText("Tooling follows below.\n" + cursorSystemPromptHint) {
		t.Fatal("expected non-prefixed cursor message-group text not to match")
	}
}
