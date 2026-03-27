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

const cursorSystemPromptHint = "You are Assistant, a coding agent based on"

type cursorItem struct{}

func init() {
	registerItem(cursorItem{})
}

func (cursorItem) Name() string {
	return "cursor"
}

func (cursorItem) MatchPrompt(prompt string) bool {
	return isCursorSystemPrompt(prompt)
}

func (cursorItem) MatchMessageGroupText(text string) bool {
	return isCursorSystemPrompt(text)
}

func isCursorSystemPrompt(content string) bool {
	return matchPromptPrefix(content, cursorSystemPromptHint)
}
