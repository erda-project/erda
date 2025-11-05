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

package metadata

import (
	"context"
	"fmt"

	"github.com/ohler55/ojg/jp"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
)

// OverrideMetadata merges override metadata into base metadata with leaf-level override support.
//   - Base metadata acts as the starting point.
//   - Override metadata overwrites base metadata on matching paths.
//   - Overrides work at leaf granularity: e.g., overriding `public.context.context-length`
//     keeps sibling fields untouched. If an override node is an empty map, the template map
//     is replaced with an empty one.
func OverrideMetadata(ctx context.Context, baseMetadata, overrideMetadata *metadatapb.Metadata) (*metadatapb.Metadata, error) {
	_ = ctx
	if baseMetadata == nil && overrideMetadata == nil {
		return nil, nil
	}

	baseMeta := FromProtobuf(baseMetadata)
	baseMap := metadataToAnyMap(baseMeta)

	overrideMeta := FromProtobuf(overrideMetadata)
	overrideMap := metadataToAnyMap(overrideMeta)

	var setErr error
	jp.Walk(overrideMap, func(path jp.Expr, value any) {
		if setErr != nil || len(path) <= 1 {
			return
		}

		switch typed := value.(type) {
		case map[string]any:
			if len(typed) == 0 {
				if isEmptyRootBucket(path) {
					return
				}
				// Empty override map should replace the existing map entirely.
				expr := copyExpr(path)
				if err := expr.Set(baseMap, cloneValue(typed)); err != nil {
					setErr = fmt.Errorf("failed to set metadata path %s: %w", expr.String(), err)
				}
			}
			// For non-empty maps we rely on leaf nodes to perform updates,
			// so skip here to avoid wiping siblings.
			return
		default:
			expr := copyExpr(path)
			if err := expr.Set(baseMap, cloneValue(value)); err != nil {
				setErr = fmt.Errorf("failed to set metadata path %s: %w", expr.String(), err)
			}
		}
	}, false)
	if setErr != nil {
		return nil, setErr
	}

	var merged Metadata
	cputil.MustObjJSONTransfer(baseMap, &merged)
	result := merged.ToProtobuf()
	if result == nil {
		return nil, nil
	}
	return result, nil
}

func metadataToAnyMap(meta Metadata) map[string]any {
	result := make(map[string]any)
	cputil.MustObjJSONTransfer(meta, &result)
	result["public"] = ensureAnyMap(result["public"])
	result["secret"] = ensureAnyMap(result["secret"])
	return result
}

func ensureAnyMap(v any) map[string]any {
	switch typed := v.(type) {
	case nil:
		return map[string]any{}
	case map[string]any:
		return typed
	default:
		var out map[string]any
		cputil.MustObjJSONTransfer(typed, &out)
		if out == nil {
			out = map[string]any{}
		}
		return out
	}
}

func copyExpr(expr jp.Expr) jp.Expr {
	cloned := make(jp.Expr, len(expr))
	copy(cloned, expr)
	return cloned
}

func cloneValue(v any) any {
	if v == nil {
		return nil
	}
	var out any
	cputil.MustObjJSONTransfer(v, &out)
	return out
}

func isEmptyRootBucket(path jp.Expr) bool {
	if len(path) != 2 {
		return false
	}
	if _, ok := path[0].(jp.Root); !ok {
		return false
	}
	child, ok := path[1].(jp.Child)
	if !ok {
		return false
	}
	switch string(child) {
	case "public", "secret":
		return true
	default:
		return false
	}
}
