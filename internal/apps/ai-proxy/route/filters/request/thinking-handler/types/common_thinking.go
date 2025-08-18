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

package types

// CommonThinking represents the unified thinking configuration
type CommonThinking struct {
	// mode controls thinking behavior: "on", "off", "auto"
	// derived from fields: thinking.type, enable_thinking
	Mode *CommonThinkingMode `json:"mode,omitempty"`

	// effort controls reasoning effort level: "minimal", "low", "medium", "high"
	// derived from fields: reasoning.effort, reasoning_effort
	Effort *CommonThinkingEffort `json:"effort,omitempty"`

	// budgetTokens controls thinking token budget
	// derived from fields: thinking.budget_tokens, thinking_budget
	BudgetTokens *int `json:"budget_tokens,omitempty"`
}

func (ct *CommonThinking) MustGetMode() CommonThinkingMode {
	if ct.Mode != nil {
		return *ct.Mode
	}
	// if mode not specified, judge by effort and budget_tokens
	ct.Mode = ModePtr(ModeOff)
	if ct.Effort != nil {
		ct.Mode = ModePtr(ModeOn)
	}
	if ct.BudgetTokens != nil && *ct.BudgetTokens > 0 {
		ct.Mode = ModePtr(ModeOn)
	}
	return *ct.Mode
}

func MapBudgetTokensToEffort(budgetTokens int) CommonThinkingEffort {
	if budgetTokens < 1024 {
		return EffortMinimal
	} else if budgetTokens < 2048 {
		return EffortLow
	} else if budgetTokens < 4096 {
		return EffortMedium
	} else {
		return EffortHigh
	}
}

func MapEffortToBudgetTokens(effort CommonThinkingEffort) int {
	switch effort {
	case EffortMinimal:
		return 1024
	case EffortLow:
		return 2048
	case EffortMedium:
		return 4096
	case EffortHigh:
		return 8192
	default:
		return 1024
	}
}

type CommonThinkingMode string

// Thinking modes
const (
	ModeOn   CommonThinkingMode = "on"
	ModeOff  CommonThinkingMode = "off"
	ModeAuto CommonThinkingMode = "auto"
)

type CommonThinkingEffort string

// Effort levels
const (
	EffortMinimal CommonThinkingEffort = "minimal"
	EffortLow     CommonThinkingEffort = "low"
	EffortMedium  CommonThinkingEffort = "medium"
	EffortHigh    CommonThinkingEffort = "high"
)

// ModePtr returns a pointer to CommonThinkingMode
func ModePtr(mode CommonThinkingMode) *CommonThinkingMode {
	return &mode
}

// EffortPtr returns a pointer to CommonThinkingEffort
func EffortPtr(effort CommonThinkingEffort) *CommonThinkingEffort {
	return &effort
}

// IsValidEffort checks if a string is a valid effort level
func IsValidEffort(effort string) bool {
	switch CommonThinkingEffort(effort) {
	case EffortMinimal, EffortLow, EffortMedium, EffortHigh:
		return true
	}
	return false
}
