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

package integration_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

func TestMultimodalEmbeddings(t *testing.T) {
	cfg := config.Get()
	models := cfg.MultimodalEmbeddingModels
	if len(models) == 0 {
		t.Skip("No multimodal embedding models configured for testing")
	}

	client := common.NewClient()
	for _, model := range models {
		t.Run(fmt.Sprintf("Model_%s", model), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			req := map[string]any{
				"model": model,
				"input": []map[string]any{{
					"type": "text",
					"text": "十字花科植物是一类广泛分布的植物。",
				}},
				"dimensions": 1024,
				"output": map[string]any{
					"primary":    "dense",
					"additional": []string{"multi"},
				},
			}

			start := time.Now()
			resp := client.PostJSON(ctx, "/v1/multimodal/embeddings", req)
			elapsed := time.Since(start)

			if resp.Error != nil {
				t.Fatalf("request failed: %v", resp.Error)
			}
			if !resp.IsSuccess() {
				t.Fatalf("request failed with status %d: %s", resp.StatusCode, string(resp.Body))
			}

			var out map[string]any
			if err := resp.GetJSON(&out); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			data, ok := out["data"].([]any)
			if !ok || len(data) == 0 {
				t.Fatalf("invalid data field: %#v", out["data"])
			}
			item, ok := data[0].(map[string]any)
			if !ok {
				t.Fatalf("invalid data[0] field: %#v", data[0])
			}
			if _, ok := item["embedding"].([]any); !ok {
				t.Fatalf("missing or invalid embedding field: %#v", item["embedding"])
			}

			t.Logf("model %s multimodal embedding success, elapsed=%v", model, elapsed)
		})
	}
}
