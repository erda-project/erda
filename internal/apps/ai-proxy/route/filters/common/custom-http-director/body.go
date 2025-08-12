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

package custom_http_director

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
	"github.com/erda-project/erda/internal/apps/ai-proxy/route/body_util"
)

// BodyTransformChange records a single body transformation operation
type BodyTransformChange struct {
	Op    string `json:"op"`              // rename, drop, default, clamp, force
	Key   string `json:"key,omitempty"`   // parameter name
	From  any    `json:"from,omitempty"`  // original value (for rename, clamp)
	To    any    `json:"to,omitempty"`    // new value (for rename, clamp)
	Value any    `json:"value,omitempty"` // set value (for default, force)
}

// BodyTransformer defines the interface for transforming request bodies based on content type
type BodyTransformer interface {
	CanHandle(contentType string) bool
	Transform(r *http.Request, config *api_style.BodyTransform) error
}

// getBodyTransformerByContentType selects appropriate body rewriter based on Content-Type
func getBodyTransformerByContentType(contentType string) BodyTransformer {
	transformers := []BodyTransformer{
		&JSONBodyTransformer{},
	}
	for _, rewriter := range transformers {
		if rewriter.CanHandle(contentType) {
			return rewriter
		}
	}
	return nil
}

// JSONBodyTransformer handles JSON body transformations
type JSONBodyTransformer struct{}

func (jbt *JSONBodyTransformer) CanHandle(contentType string) bool {
	return strings.HasPrefix(contentType, "application/json")
}

func (jbt *JSONBodyTransformer) Transform(r *http.Request, config *api_style.BodyTransform) error {
	if config == nil {
		return nil
	}
	return jbt.transformBody(r, config)
}

func (jbt *JSONBodyTransformer) transformBody(r *http.Request, config *api_style.BodyTransform) error {
	if config == nil {
		return nil
	}

	var jsonBody map[string]any
	if err := json.NewDecoder(r.Body).Decode(&jsonBody); err != nil {
		return fmt.Errorf("failed to decode request body: %v", err)
	}

	// Apply transformations
	var changes []BodyTransformChange
	if err := jbt.applyTransformations(jsonBody, config, &changes); err != nil {
		return fmt.Errorf("failed to apply transformations: %v", err)
	}

	// reset body
	b, err := json.Marshal(jsonBody)
	if err != nil {
		return fmt.Errorf("failed to marshal transformed body: %v", err)
	}
	if err := body_util.SetBody(r, b); err != nil {
		return fmt.Errorf("failed to set transformed body: %w", err)
	}

	// store changes in context
	if len(changes) > 0 {
		ctxhelper.PutRequestBodyTransformChanges(r.Context(), &changes)
	}

	return nil
}

// applyTransformations applies all transformation rules in the fixed order
func (jbt *JSONBodyTransformer) applyTransformations(
	jsonBody map[string]any,
	config *api_style.BodyTransform,
	changes *[]BodyTransformChange,
) error {
	// 1. rename parameters
	if err := jbt.applyRename(jsonBody, config.Rename, changes); err != nil {
		return fmt.Errorf("failed to apply rename: %v", err)
	}

	// 2. set default values
	if err := jbt.applyDefault(jsonBody, config.Default, changes); err != nil {
		return fmt.Errorf("failed to apply default: %v", err)
	}

	// 3. force set values
	if err := jbt.applyForce(jsonBody, config.Force, changes); err != nil {
		return fmt.Errorf("failed to apply force: %v", err)
	}

	// 4. drop explicitly specified parameters
	if err := jbt.applyDrop(jsonBody, config.Drop, changes); err != nil {
		return fmt.Errorf("failed to apply drop: %v", err)
	}

	// 5. apply numeric clamping
	if err := jbt.applyClamp(jsonBody, config.Clamp, changes); err != nil {
		return fmt.Errorf("failed to apply clamp: %v", err)
	}

	return nil
}

// applyRename renames parameters according to the rename map
func (jbt *JSONBodyTransformer) applyRename(jsonBody map[string]any, renameMap map[string]string, changes *[]BodyTransformChange) error {
	for oldKey, newKey := range renameMap {
		if oldValue, hasOld := jsonBody[oldKey]; hasOld {
			// if new key already exists, skip
			if _, hasNew := jsonBody[newKey]; hasNew {
				continue
			}
			// rename old to new
			jsonBody[newKey] = oldValue
			delete(jsonBody, oldKey)
			*changes = append(*changes, BodyTransformChange{
				Op:   "rename",
				From: oldKey,
				To:   newKey,
			})
		}
	}
	return nil
}

// applyDefault sets default values for missing keys
func (jbt *JSONBodyTransformer) applyDefault(jsonBody map[string]any, defaultMap map[string]any, changes *[]BodyTransformChange) error {
	for key, defaultValue := range defaultMap {
		if _, exists := jsonBody[key]; !exists {
			jsonBody[key] = defaultValue
			*changes = append(*changes, BodyTransformChange{
				Op:    "default",
				Key:   key,
				Value: defaultValue,
			})
		}
	}
	return nil
}

// applyForce sets values, overriding existing ones
func (jbt *JSONBodyTransformer) applyForce(jsonBody map[string]any, forceMap map[string]any, changes *[]BodyTransformChange) error {
	for key, forceValue := range forceMap {
		oldValue := jsonBody[key]
		jsonBody[key] = forceValue
		*changes = append(*changes, BodyTransformChange{
			Op:    "force",
			Key:   key,
			From:  oldValue,
			Value: forceValue,
		})
	}
	return nil
}

// applyDrop removes explicitly specified parameters
func (jbt *JSONBodyTransformer) applyDrop(jsonBody map[string]any, dropList []string, changes *[]BodyTransformChange) error {
	for _, key := range dropList {
		if oldValue, exists := jsonBody[key]; exists {
			delete(jsonBody, key)
			*changes = append(*changes, BodyTransformChange{
				Op:   "drop",
				Key:  key,
				From: oldValue,
			})
		}
	}
	return nil
}

// applyClamp applies numeric constraints to values
func (jbt *JSONBodyTransformer) applyClamp(jsonBody map[string]any, clampMap map[string]api_style.NumericClamp, changes *[]BodyTransformChange) error {
	for key, clamp := range clampMap {
		if value, exists := jsonBody[key]; exists {
			// JSON numbers decode to float64 in map[string]any
			numValue, ok := value.(float64)
			if !ok {
				// not a number in JSON body; skip
				continue
			}

			originalValue := numValue
			clamped := false

			// apply min constraint
			if clamp.Min != nil && numValue < *clamp.Min {
				numValue = *clamp.Min
				clamped = true
			}

			// apply max constraint
			if clamp.Max != nil && numValue > *clamp.Max {
				numValue = *clamp.Max
				clamped = true
			}

			if clamped {
				// keep as float64; JSON encoder handles integral values naturally
				jsonBody[key] = numValue

				*changes = append(*changes, BodyTransformChange{
					Op:   "clamp",
					Key:  key,
					From: originalValue,
					To:   numValue,
				})
			}
		}
	}
	return nil
}
