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

func TestOpenClawItem_MatchPrompt(t *testing.T) {
	matcher, ok := any(openClawItem{}).(PromptMatcher)
	if !ok {
		t.Fatal("expected openclaw item to implement PromptMatcher")
	}
	if !matcher.MatchPrompt(openClawSystemPromptHint + "\n## Tooling\nTool availability") {
		t.Fatal("expected openclaw prompt prefix to match")
	}
}

func TestOpenClawItem_IgnorePromptWithoutPrefix(t *testing.T) {
	matcher := any(openClawItem{}).(PromptMatcher)
	if matcher.MatchPrompt("Tooling follows below.\n" + openClawSystemPromptHint) {
		t.Fatal("expected non-prefixed openclaw prompt not to match")
	}
}

func TestOpenClawItem_MatchMessageGroupText(t *testing.T) {
	matcher, ok := any(openClawItem{}).(MessageGroupMatcher)
	if !ok {
		t.Fatal("expected openclaw item to implement MessageGroupMatcher")
	}
	if !matcher.MatchMessageGroupText(openClawSystemPromptHint + "\n## Tooling\nTool availability") {
		t.Fatal("expected openclaw message-group text prefix to match")
	}
}

func TestOpenClawItem_IgnoreMessageGroupTextWithoutPrefix(t *testing.T) {
	matcher := any(openClawItem{}).(MessageGroupMatcher)
	if matcher.MatchMessageGroupText("Tooling follows below.\n" + openClawSystemPromptHint) {
		t.Fatal("expected non-prefixed openclaw message-group text not to match")
	}
}
