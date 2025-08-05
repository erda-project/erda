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

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

func TestEmbeddings(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	// Get embedding models for testing
	embeddingModels := cfg.EmbeddingsModels
	if len(embeddingModels) == 0 {
		t.Skip("No embedding models configured for testing")
	}

	// Define test texts
	testTexts := []struct {
		input       string
		description string
	}{
		{"Hello, world!", "Simple_greeting"},
		{"The quick brown fox jumps over the lazy dog.", "Classic_pangram"},
		{"Machine learning is a subset of artificial intelligence.", "Technical_description"},
	}

	for _, model := range embeddingModels {
		for _, testText := range testTexts {
			// Test with header method (existing)
			t.Run(fmt.Sprintf("Header_Model_%s_Text_%s", model, testText.description), func(t *testing.T) {
				testEmbeddingForModelWithHeader(t, client, model, testText.input)
			})

			// Test with JSON body method (new)
			t.Run(fmt.Sprintf("JSONBody_Model_%s_Text_%s", model, testText.description), func(t *testing.T) {
				testEmbeddingForModelWithJSONBody(t, client, model, testText.input)
			})
		}
	}
}

func testEmbeddingForModelWithHeader(t *testing.T, client *common.Client, model, input string) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	// Create embedding request without setting model in request body
	request := struct {
		Input []string `json:"input"`
	}{
		Input: []string{input},
	}

	// Send request with model passed via header
	headers := map[string]string{
		"X-AI-Proxy-Model": model,
	}
	startTime := time.Now()
	resp := client.PostJSONWithHeaders(ctx, "/v1/embeddings", request, headers)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var embeddingResp openai.EmbeddingResponse
	if err := resp.GetJSON(&embeddingResp); err != nil {
		t.Fatalf("✗ Failed to parse response: %v", err)
	}

	// Validate response
	if len(embeddingResp.Data) == 0 {
		t.Error("✗ No embeddings returned")
	}

	if len(embeddingResp.Data) > 0 {
		embedding := embeddingResp.Data[0]
		if len(embedding.Embedding) == 0 {
			t.Error("✗ Empty embedding vector")
		}

		// Validate basic properties of embedding vector
		if embedding.Index != 0 {
			t.Errorf("✗ Expected embedding index 0, got %d", embedding.Index)
		}

		// Check if embedding vector dimension is reasonable (usually between hundreds to thousands)
		if len(embedding.Embedding) < 100 || len(embedding.Embedding) > 10000 {
			t.Logf("⚠ Unusual embedding dimension: %d", len(embedding.Embedding))
		}
	}

	// Validate usage information
	if embeddingResp.Usage.TotalTokens == 0 {
		t.Error("✗ No token usage reported")
	}

	// Check response time reasonableness
	if responseTime > 10*time.Second {
		t.Logf("⚠ Slow embedding response time: %v", responseTime)
	}

	t.Logf("✓ Model %s embedded text via header: '%s' (dimension: %d, tokens: %d, response time: %v)",
		model, truncateString(input, 50), len(embeddingResp.Data[0].Embedding),
		embeddingResp.Usage.TotalTokens, responseTime)
}

func testEmbeddingForModelWithJSONBody(t *testing.T, client *common.Client, model, input string) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	// Create embedding request with model included in request body
	request := struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}{
		Model: model,
		Input: []string{input},
	}

	startTime := time.Now()
	resp := client.PostJSON(ctx, "/v1/embeddings", request)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var embeddingResp openai.EmbeddingResponse
	if err := resp.GetJSON(&embeddingResp); err != nil {
		t.Fatalf("✗ Failed to parse response: %v", err)
	}

	// Validate response
	if len(embeddingResp.Data) == 0 {
		t.Error("✗ No embeddings returned")
	}

	if len(embeddingResp.Data) > 0 {
		embedding := embeddingResp.Data[0]
		if len(embedding.Embedding) == 0 {
			t.Error("✗ Empty embedding vector")
		}

		// Validate basic properties of embedding vector
		if embedding.Index != 0 {
			t.Errorf("✗ Expected embedding index 0, got %d", embedding.Index)
		}

		// Check if embedding vector dimension is reasonable (usually between hundreds to thousands)
		if len(embedding.Embedding) < 100 || len(embedding.Embedding) > 10000 {
			t.Logf("⚠ Unusual embedding dimension: %d", len(embedding.Embedding))
		}
	}

	// Validate usage information
	if embeddingResp.Usage.TotalTokens == 0 {
		t.Error("✗ No token usage reported")
	}

	// Check response time reasonableness
	if responseTime > 10*time.Second {
		t.Logf("⚠ Slow embedding response time: %v", responseTime)
	}

	t.Logf("✓ Model %s embedded text via JSON body: '%s' (dimension: %d, tokens: %d, response time: %v)",
		model, truncateString(input, 50), len(embeddingResp.Data[0].Embedding),
		embeddingResp.Usage.TotalTokens, responseTime)
}

// TestEmbeddingsErrorHandling tests error handling
func TestEmbeddingsErrorHandling(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	embeddingModels := cfg.EmbeddingsModels
	if len(embeddingModels) == 0 {
		t.Skip("No embedding models configured for testing")
	}

	model := embeddingModels[0] // Use first model for error testing

	t.Run("EmptyInput", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		// Create request with empty input
		request := struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}{
			Model: model,
			Input: []string{},
		}

		headers := map[string]string{
			"X-AI-Proxy-Model": model,
		}
		resp := client.PostJSONWithHeaders(ctx, "/v1/embeddings", request, headers)

		// Expect this request to fail
		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with empty input, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected empty input (status: %d)", resp.StatusCode)
		}
	})

	t.Run("InvalidModel", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		// Create request using non-existent model
		request := struct {
			Model string   `json:"model"`
			Input []string `json:"input"`
		}{
			Model: "non-existent-model",
			Input: []string{"test text"},
		}

		headers := map[string]string{
			"X-AI-Proxy-Model": "non-existent-model",
		}
		resp := client.PostJSONWithHeaders(ctx, "/v1/embeddings", request, headers)

		// Expect this request to fail
		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with invalid model, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected invalid model (status: %d)", resp.StatusCode)
		}
	})

	t.Run("MissingModel", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		// Create request without model
		request := struct {
			Input []string `json:"input"`
		}{
			Input: []string{"test text"},
		}

		resp := client.PostJSON(ctx, "/v1/embeddings", request)

		// Expect this request to fail
		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with missing model, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected missing model (status: %d)", resp.StatusCode)
		}
	})

}
