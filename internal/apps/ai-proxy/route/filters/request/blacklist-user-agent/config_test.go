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

func TestNormalizeBlacklist_NormalizesCommaSeparatedValues(t *testing.T) {
	got := normalizeBlacklist(splitBlacklist(" openclaw, Coding-Agent ,, "))
	if len(got) != 2 || got[0] != "openclaw" || got[1] != "coding-agent" {
		t.Fatalf("unexpected normalized blacklist: %#v", got)
	}
}

func TestNormalizeGeneralRules_NormalizesHeadersAndPrompts(t *testing.T) {
	headers := normalizeGeneralRules(splitGeneralRules(" Claude Code ;;; X-Code;Agent ;;; ;; "))
	if len(headers) != 2 || headers[0] != "claude code" || headers[1] != "x-code;agent" {
		t.Fatalf("unexpected general header rules: %#v", headers)
	}

	prompts := normalizeGeneralRules(splitGeneralRules(" You are OpenCode ;;; You are Claude; Code ;;; ;; "))
	if len(prompts) != 2 || prompts[0] != "you are opencode" || prompts[1] != "you are claude; code" {
		t.Fatalf("unexpected general prompt rules: %#v", prompts)
	}
}
