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
	"testing"

	"github.com/alecthomas/assert"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
)

func TestOverrideMetadata(t *testing.T) {
	baseFactory := func() *metadatapb.Metadata {
		return (&Metadata{
			Public: map[string]any{
				"context": map[string]any{
					"context-length":        131072,
					"max-completion-tokens": 16384,
					"max-prompt-tokens":     129024,
				},
				"abilities": map[string]any{
					"abilities":         []any{"reasoning", "enable_thinking", "tool"},
					"input_modalities":  []any{"text"},
					"output_modalities": []any{"text"},
				},
				"publisher": "tpl-publisher",
			},
			Secret: map[string]any{
				"api_key": "tpl-secret",
			},
		}).ToProtobuf()
	}

	t.Run("override leaf keeps siblings", func(t *testing.T) {
		overrideMeta := &Metadata{
			Public: map[string]any{
				"context": map[string]any{
					"context-length": 130000,
				},
			},
		}

		base := baseFactory()

		result, err := OverrideMetadata(context.Background(), base, overrideMeta.ToProtobuf())
		assert.NoError(t, err)

		merged := FromProtobuf(result)
		contextInfo := ensureMapAny(t, merged.Public["context"])
		assert.Equal(t, float64(130000), contextInfo["context-length"])
		assert.Equal(t, float64(16384), contextInfo["max-completion-tokens"])
		assert.Equal(t, float64(129024), contextInfo["max-prompt-tokens"])

		secret := ensureMapAny(t, merged.Secret)
		assert.Equal(t, "tpl-secret", secret["api_key"])
		assert.Equal(t, "tpl-publisher", merged.MustGetValueByKey("publisher"))
	})

	t.Run("override nested array element", func(t *testing.T) {
		overrideMeta := &Metadata{
			Public: map[string]any{
				"abilities": map[string]any{
					"output_modalities": []any{"vision"},
				},
			},
		}

		base := baseFactory()

		result, err := OverrideMetadata(context.Background(), base, overrideMeta.ToProtobuf())
		assert.NoError(t, err)

		merged := FromProtobuf(result)
		abilities := ensureMapAny(t, merged.Public["abilities"])
		assert.Equal(t, []any{"vision"}, abilities["output_modalities"])
		assert.Equal(t, []any{"reasoning", "enable_thinking", "tool"}, abilities["abilities"])
		assert.Equal(t, []any{"text"}, abilities["input_modalities"])
	})

	t.Run("override new field and publisher", func(t *testing.T) {
		overrideMeta := &Metadata{
			Public: map[string]any{
				"publisher": "custom-publisher",
				"extra": map[string]any{
					"enabled": true,
				},
			},
			Secret: map[string]any{
				"api_key": "override-secret",
			},
		}

		base := baseFactory()

		result, err := OverrideMetadata(context.Background(), base, overrideMeta.ToProtobuf())
		assert.NoError(t, err)

		merged := FromProtobuf(result)
		assert.Equal(t, "custom-publisher", merged.MustGetValueByKey("publisher"))
		extra := ensureMapAny(t, merged.Public["extra"])
		assert.Equal(t, true, extra["enabled"])
		secret := ensureMapAny(t, merged.Secret)
		assert.Equal(t, "override-secret", secret["api_key"])
	})

	t.Run("base only metadata when override nil", func(t *testing.T) {
		base := baseFactory()

		result, err := OverrideMetadata(context.Background(), base, nil)
		assert.NoError(t, err)

		merged := FromProtobuf(result)
		contextInfo := ensureMapAny(t, merged.Public["context"])
		assert.Equal(t, float64(131072), contextInfo["context-length"])
		assert.Equal(t, "tpl-publisher", merged.MustGetValueByKey("publisher"))
	})

	t.Run("override metadata without base", func(t *testing.T) {
		overrideMeta := &Metadata{
			Public: map[string]any{
				"publisher": "override-only",
			},
			Secret: map[string]any{
				"token": "secret-token",
			},
		}

		result, err := OverrideMetadata(context.Background(), nil, overrideMeta.ToProtobuf())
		assert.NoError(t, err)

		merged := FromProtobuf(result)
		public := ensureMapAny(t, merged.Public)
		assert.Equal(t, "override-only", public["publisher"])
		secret := ensureMapAny(t, merged.Secret)
		assert.Equal(t, "secret-token", secret["token"])
		assert.Equal(t, "override-only", merged.MustGetValueByKey("publisher"))
	})

	t.Run("nil base and override returns nil", func(t *testing.T) {
		result, err := OverrideMetadata(context.Background(), nil, nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func ensureMapAny(t *testing.T, v any) map[string]any {
	t.Helper()
	switch typed := v.(type) {
	case map[string]any:
		return typed
	default:
		t.Fatalf("expected map, got %T", v)
		return nil
	}
}
