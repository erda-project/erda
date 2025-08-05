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
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

func TestV1Models(t *testing.T) {
	client := common.NewClient()

	t.Run("ListModels", func(t *testing.T) {
		testListModels(t, client)
	})
}

func testListModels(t *testing.T, client *common.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	startTime := time.Now()
	resp := client.Get(ctx, "/v1/models")
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var modelsList openai.ModelsList
	if err := resp.GetJSON(&modelsList); err != nil {
		t.Fatalf("✗ Failed to parse response: %v", err)
	}

	// Validate response structure
	if len(modelsList.Models) == 0 {
		t.Error("✗ No models returned in response")
	}

	// Validate basic fields of each model
	for i, model := range modelsList.Models {
		if model.ID == "" {
			t.Errorf("✗ Model %d: ID is empty", i)
		}
		if model.Object == "" {
			t.Errorf("✗ Model %d: Object is empty", i)
		}
		if model.OwnedBy == "" {
			t.Errorf("✗ Model %d: OwnedBy is empty", i)
		}
		if model.Object != "model" {
			t.Errorf("✗ Model %d: Expected object 'model', got '%s'", i, model.Object)
		}
	}

	t.Logf("✓ Models List: Found %d models (response time: %v)", len(modelsList.Models), responseTime)

	// Log information of first few models
	for i, model := range modelsList.Models {
		if i >= 3 { // Only show first 3 models
			break
		}
		t.Logf("  - Model %d: %s (owned by: %s)", i+1, model.ID, model.OwnedBy)
	}

	if len(modelsList.Models) > 3 {
		t.Logf("  ... and %d more models", len(modelsList.Models)-3)
	}
}
