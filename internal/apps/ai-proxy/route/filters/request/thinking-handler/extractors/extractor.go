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

package extractors

import (
	"encoding/json"

	"github.com/erda-project/erda/internal/apps/ai-proxy/route/filters/request/thinking-handler/types"
)

// CommonThinkingExtractor defines the interface for extracting CommonThinking fields
type CommonThinkingExtractor interface {
	// ExtractMode extracts thinking mode from request body
	ExtractMode(body map[string]any) (*types.CommonThinkingMode, error)

	// ExtractEffort extracts effort level from request body
	ExtractEffort(body map[string]any) (*types.CommonThinkingEffort, error)

	// ExtractBudgetTokens extracts budget tokens from request body
	ExtractBudgetTokens(body map[string]any) (*int, error)

	// GetPriority returns the priority of this extractor (lower number = higher priority)
	GetPriority() int

	// GetName returns the name of this extractor for logging and debugging
	GetName() string

	// CanExtract checks if this extractor can process the given body
	CanExtract(body map[string]any) bool

	// RelatedFields extracts the original fields that this extractor is concerned with
	RelatedFields(body map[string]any) map[string]any
}

// ExtractResult represents the result of extraction from one extractor
type ExtractResult struct {
	ExtractorName string
	*types.CommonThinking
	// OriginalFields contains the original fields that this extractor processed
	OriginalFields map[string]any
}

// ExtractIntValue safely extracts integer value from interface{}
func ExtractIntValue(val any) *int {
	switch v := val.(type) {
	case int:
		return &v
	case float64:
		intVal := int(v)
		return &intVal
	case json.Number:
		if intVal, err := v.Int64(); err == nil {
			result := int(intVal)
			return &result
		}
	}
	return nil
}
