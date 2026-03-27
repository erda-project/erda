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

type generalItem struct{}

func init() {
	registerItem(generalItem{})
}

func (generalItem) Name() string {
	return "general"
}

func (generalItem) Enabled() bool {
	return len(getConfig().General.ItemTypes) > 0
}

func (generalItem) MatchHeader(key, value string) bool {
	return containsConfiguredGeneralItemType(key) || containsConfiguredGeneralItemType(value)
}

func (generalItem) MatchPrompt(prompt string) bool {
	return containsConfiguredGeneralItemType(prompt)
}

func (generalItem) MatchMessageGroupText(text string) bool {
	return containsConfiguredGeneralItemType(text)
}

func containsConfiguredGeneralItemType(input string) bool {
	if input == "" {
		return false
	}
	normalizedInput := normalize(input)
	for _, itemType := range getConfig().General.ItemTypes {
		if itemType != "" && strings.Contains(normalizedInput, itemType) {
			return true
		}
	}
	return false
}
