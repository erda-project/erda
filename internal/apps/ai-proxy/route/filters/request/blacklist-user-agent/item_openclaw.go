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

// Source: https://github.com/search?q=repo:openclaw/openclaw+%22You+are+a+personal+assistant+running+inside+OpenClaw.%22&type=code
const openClawSystemPromptHint = "You are a personal assistant running inside OpenClaw."

type openClawItem struct{}

func init() {
	registerItem(openClawItem{})
}

func (openClawItem) Name() string {
	return "openclaw"
}

func (openClawItem) MatchPrompt(prompt string) bool {
	return isOpenClawSystemPrompt(prompt)
}

func (openClawItem) MatchMessageGroupText(text string) bool {
	return isOpenClawSystemPrompt(text)
}

func isOpenClawSystemPrompt(content string) bool {
	return matchPromptPrefix(content, openClawSystemPromptHint)
}
