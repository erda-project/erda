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

type generalItem struct {
	rules GeneralRules
}

func (generalItem) Name() string {
	return "general"
}

func (g generalItem) MatchHeader(key, value string) bool {
	return g.containsConfiguredGeneralHeaderRule(key) || g.containsConfiguredGeneralHeaderRule(value)
}

func (g generalItem) MatchPrompt(prompt string) bool {
	return g.hasConfiguredGeneralPromptPrefix(prompt)
}

func (g generalItem) MatchMessageGroupText(text string) bool {
	return g.hasConfiguredGeneralPromptPrefix(text)
}

func (g generalItem) containsConfiguredGeneralHeaderRule(input string) bool {
	if input == "" {
		return false
	}
	normalizedInput := normalize(input)
	for _, rule := range g.rules.Headers {
		if rule != "" && strings.Contains(normalizedInput, rule) {
			return true
		}
	}
	return false
}

func (g generalItem) hasConfiguredGeneralPromptPrefix(input string) bool {
	for _, rule := range g.rules.Prompts {
		if rule != "" && matchPromptWithinWindowContains(input, rule, generalPromptMatchWindow) {
			return true
		}
	}
	return false
}
