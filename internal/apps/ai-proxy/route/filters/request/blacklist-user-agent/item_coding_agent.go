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

const (
	codingAgentItemName    = "coding-agent"
	codexSystemPromptHint  = "You are Codex"
	cursorSystemPromptHint = "You are Assistant, a coding agent based on"
)

type codingAgentItem struct{}

type codingAgentSubtype struct {
	name                  string
	matchHeader           func(key, value string) bool
	matchPrompt           func(prompt string) bool
	matchMessageGroupText func(text string) bool
}

var codingAgentSubtypes = []codingAgentSubtype{
	{
		name:                  "codex",
		matchHeader:           matchCodexHeader,
		matchPrompt:           isCodexSystemPrompt,
		matchMessageGroupText: isCodexSystemPrompt,
	},
	{
		name:                  "cursor",
		matchPrompt:           isCursorSystemPrompt,
		matchMessageGroupText: isCursorSystemPrompt,
	},
}

func init() {
	registerItem(codingAgentItem{})
}

func (codingAgentItem) Name() string {
	return codingAgentItemName
}

func (codingAgentItem) MatchHeader(key, value string) bool {
	return matchCodingAgentHeaderSubtype(key, value) != ""
}

func (codingAgentItem) MatchPrompt(prompt string) bool {
	return matchCodingAgentPromptSubtype(prompt) != ""
}

func (codingAgentItem) MatchMessageGroupText(text string) bool {
	return matchCodingAgentMessageGroupSubtype(text) != ""
}

func matchCodingAgentHeaderSubtype(key, value string) string {
	for _, subtype := range codingAgentSubtypes {
		if subtype.matchHeader != nil && subtype.matchHeader(key, value) {
			return subtype.name
		}
	}
	return ""
}

func matchCodingAgentPromptSubtype(prompt string) string {
	for _, subtype := range codingAgentSubtypes {
		if subtype.matchPrompt != nil && subtype.matchPrompt(prompt) {
			return subtype.name
		}
	}
	return ""
}

func matchCodingAgentMessageGroupSubtype(text string) string {
	for _, subtype := range codingAgentSubtypes {
		if subtype.matchMessageGroupText != nil && subtype.matchMessageGroupText(text) {
			return subtype.name
		}
	}
	return ""
}

func isCodexSystemPrompt(content string) bool {
	return matchPromptPrefix(content, codexSystemPromptHint)
}

func matchCodexHeader(key, value string) bool {
	return containsCodex(key) || containsCodex(value)
}

func containsCodex(input string) bool {
	return strings.Contains(strings.ToLower(input), "codex")
}

func isCursorSystemPrompt(content string) bool {
	return matchPromptPrefix(content, cursorSystemPromptHint)
}
