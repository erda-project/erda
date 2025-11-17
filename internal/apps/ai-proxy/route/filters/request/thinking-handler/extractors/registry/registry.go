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

package registry

import (
	"fmt"
	"sort"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/extractors"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/extractors/impl"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// Registry manages all CommonThinking extractors
type Registry struct {
	extractors []extractors.CommonThinkingExtractor
}

// NewRegistry creates a new registry with default extractors
func NewRegistry() *Registry {
	extractorList := []extractors.CommonThinkingExtractor{
		&impl.AnthropicThinkingExtractor{},
		&impl.QwenThinkingExtractor{},
		&impl.OpenAIResponsesThinkingExtractor{},
		&impl.OpenAIChatThinkingExtractor{},
		&impl.GeminiThinkingExtractor{},
	}

	// sort by priority
	sort.Slice(extractorList, func(i, j int) bool {
		return extractorList[i].GetPriority() < extractorList[j].GetPriority()
	})

	return &Registry{
		extractors: extractorList,
	}
}

// ExtractAll extracts thinking fields using all applicable extractors
// return:
// - map of original fields
// - merged thinking config
// - error
func (r *Registry) ExtractAll(body map[string]any) (map[string]any, *types.CommonThinking, error) {
	var results []extractors.ExtractResult

	for _, extractor := range r.extractors {
		if !extractor.CanExtract(body) {
			continue
		}

		result := extractors.ExtractResult{
			ExtractorName:  extractor.GetName(),
			CommonThinking: &types.CommonThinking{},
			OriginalFields: extractor.RelatedFields(body),
		}

		// extract mode
		if mode, err := extractor.ExtractMode(body); err != nil {
			return nil, nil, fmt.Errorf("extractor %s failed to extract mode: %w", extractor.GetName(), err)
		} else if mode != nil {
			result.Mode = mode
		}

		// extract effort
		if effort, err := extractor.ExtractEffort(body); err != nil {
			return nil, nil, fmt.Errorf("extractor %s failed to extract effort: %w", extractor.GetName(), err)
		} else if effort != nil {
			result.Effort = effort
		}

		// extract budget tokens
		if budget, err := extractor.ExtractBudgetTokens(body); err != nil {
			return nil, nil, fmt.Errorf("extractor %s failed to extract budget tokens: %w", extractor.GetName(), err)
		} else if budget != nil {
			result.BudgetTokens = budget
		}

		// only add result if it extracted something
		if result.CommonThinking.Mode != nil || result.CommonThinking.Effort != nil || result.CommonThinking.BudgetTokens != nil {
			results = append(results, result)
		}
	}

	// get original fields
	originalFields := mergeOriginalFields(results)

	// merge results to get final thinking config
	ct, err := mergeAndValidateResults(results)

	return originalFields, ct, err
}

// mergeAndValidateResults merges extraction results using priority-based selection
// Since extractors are already sorted by priority, the first valid result wins
func mergeAndValidateResults(results []extractors.ExtractResult) (*types.CommonThinking, error) {
	if len(results) == 0 {
		return nil, nil
	}

	ct := &types.CommonThinking{}

	// merge mode by priority (first non-nil wins)
	// if multiple extractors provide a mode, we should check if conflicting modes exist
	for _, result := range results {
		if result.CommonThinking.Mode != nil && ct.Mode == nil {
			if ct.Mode == nil {
				ct.Mode = result.CommonThinking.Mode
			} else {
				if ct.Mode != result.CommonThinking.Mode {
					return nil, fmt.Errorf("conflicting thinking modes detected")
				}
			}
		}
	}

	// merge effort by priority (first non-nil wins)
	for _, result := range results {
		if result.CommonThinking.Effort != nil && ct.Effort == nil {
			ct.Effort = result.CommonThinking.Effort
			break
		}
	}

	// merge budget tokens by priority (first non-nil wins)
	for _, result := range results {
		if result.CommonThinking.BudgetTokens != nil && ct.BudgetTokens == nil {
			ct.BudgetTokens = result.CommonThinking.BudgetTokens
			break
		}
	}

	return ct, nil
}

// mergeOriginalFields merges the OriginalFields from all ExtractResults
// If the same key appears in multiple results, the later one will overwrite the former
func mergeOriginalFields(results []extractors.ExtractResult) map[string]any {
	merged := make(map[string]any)
	for _, result := range results {
		for key, value := range result.OriginalFields {
			merged[key] = value
		}
	}
	return merged
}
