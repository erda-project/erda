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

import "strings"

const codexSystemPromptHint = "You are Codex"

type codexItem struct{}

func init() {
	registerItem(codexItem{})
}

func (codexItem) Name() string {
	return "codex"
}

func (codexItem) MatchHeader(key, value string) bool {
	return containsCodex(key) || containsCodex(value)
}

func (codexItem) MatchPrompt(prompt string) bool {
	return isCodexSystemPrompt(prompt)
}

func isCodexSystemPrompt(content string) bool {
	return matchPromptPrefix(content, codexSystemPromptHint)
}

func containsCodex(input string) bool {
	return strings.Contains(strings.ToLower(input), "codex")
}
