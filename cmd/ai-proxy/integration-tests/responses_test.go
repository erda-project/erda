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
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

// supportsGetOperations checks if model supports GET operations
func supportsGetOperations(model string) bool {
	// doubao-seed-1.6-250615 model does not support GET /v1/responses/{id} and GET /v1/responses/{id}/input_items
	if strings.Contains(model, "doubao-seed") {
		return false
	}
	// Default supports GET operations
	return true
}

// supportsMetadata checks if model supports metadata field
func supportsMetadata(model string) bool {
	// doubao-seed-1.6-250615 model does not support metadata field
	if strings.Contains(model, "doubao-seed") {
		return false
	}
	// Default supports metadata
	return true
}

// supportsStreaming checks if model supports streaming
func supportsStreaming(model string) bool {
	// Some older versions of models may not support streaming
	// Currently assume all models support it, can add unsupported ones here if found
	return true
}

// TestResponsesNonStream tests responses API in non-streaming mode
func TestResponsesNonStream(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	// Get responses models for testing
	responsesModels := cfg.ResponsesModels
	if len(responsesModels) == 0 {
		t.Skip("No responses models configured for testing")
	}

	for _, model := range responsesModels {
		t.Run(fmt.Sprintf("NonStream_Model_%s", model), func(t *testing.T) {
			testResponsesNonStreamForModel(t, client, model)
		})
	}
}

// TestResponsesStreaming tests responses API in streaming mode
func TestResponsesStreaming(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	// Get responses models for testing
	responsesModels := cfg.ResponsesModels
	if len(responsesModels) == 0 {
		t.Skip("No responses models configured for testing")
	}

	for _, model := range responsesModels {
		t.Run(fmt.Sprintf("Streaming_Model_%s", model), func(t *testing.T) {
			testResponsesStreamingForModel(t, client, model)
		})
	}
}

// TestResponses compatibility test function (runs all tests)
func TestResponses(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	// Get responses models for testing
	responsesModels := cfg.ResponsesModels
	if len(responsesModels) == 0 {
		t.Skip("No responses models configured for testing")
	}

	for _, model := range responsesModels {
		t.Run(fmt.Sprintf("Responses_Model_%s", model), func(t *testing.T) {
			testResponsesWorkflowForModel(t, client, model)
		})
	}
}

// testResponsesWorkflowForModel tests complete responses workflow for single model
func testResponsesWorkflowForModel(t *testing.T, client *common.Client, model string) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	supportsGet := supportsGetOperations(model)
	if !supportsGet {
		t.Logf("Model %s does not support GET operations, testing only POST and DELETE", model)
	}

	// Step 1: Create first Response
	firstResponseID := createResponse(t, client, ctx, model, "Hello, can you help me write a simple Python function to calculate factorial?")
	t.Logf("✓ Created first response ID: %s", firstResponseID)

	// Step 2: Get first Response details (if supported)
	if supportsGet {
		getResponse(t, client, ctx, firstResponseID, model)
		t.Logf("✓ Retrieved first response details")
	} else {
		t.Logf("- Skipped getting first response details (not supported by model)")
	}

	// Step 3: Create second Response for chained conversation (using previous_response_id)
	secondResponseID := createChainedResponse(t, client, ctx, model, "Can you make it more efficient and add error handling?", firstResponseID)
	t.Logf("✓ Created chained response ID: %s", secondResponseID)

	// Step 4: Get second Response details (if supported)
	if supportsGet {
		getResponse(t, client, ctx, secondResponseID, model)
		t.Logf("✓ Retrieved second response details")
	} else {
		t.Logf("- Skipped getting second response details (not supported by model)")
	}

	// Step 5: Get Response input items (if supported)
	if supportsGet {
		getResponseInputItems(t, client, ctx, secondResponseID, model)
		t.Logf("✓ Retrieved response input items")
	} else {
		t.Logf("- Skipped getting response input items (not supported by model)")
	}

	// Step 6: Delete Responses (cleanup)
	deleteResponse(t, client, ctx, secondResponseID, model)
	t.Logf("✓ Deleted second response")

	deleteResponse(t, client, ctx, firstResponseID, model)
	t.Logf("✓ Deleted first response")

	t.Logf("✓ Responses workflow completed successfully for model: %s", model)
}

// testResponsesNonStreamForModel tests responses non-streaming mode for single model
func testResponsesNonStreamForModel(t *testing.T, client *common.Client, model string) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	supportsGet := supportsGetOperations(model)
	if !supportsGet {
		t.Logf("Model %s does not support GET operations, testing only POST and DELETE", model)
	}

	// Step 1: Create first Response (non-streaming)
	firstResponseID := createNonStreamResponse(t, client, ctx, model, "Hello, can you help me write a simple Python function to calculate factorial?")
	t.Logf("✓ Created first non-stream response ID: %s", firstResponseID)

	// Step 2: Get first Response details (if supported)
	if supportsGet {
		getResponse(t, client, ctx, firstResponseID, model)
		t.Logf("✓ Retrieved first response details")
	} else {
		t.Logf("- Skipped getting first response details (not supported by model)")
	}

	// Step 3: Create second Response for chained conversation (non-streaming)
	secondResponseID := createChainedNonStreamResponse(t, client, ctx, model, "Can you make it more efficient and add error handling?", firstResponseID)
	t.Logf("✓ Created chained non-stream response ID: %s", secondResponseID)

	// Step 4: Get second Response details (if supported)
	if supportsGet {
		getResponse(t, client, ctx, secondResponseID, model)
		t.Logf("✓ Retrieved second response details")
	} else {
		t.Logf("- Skipped getting second response details (not supported by model)")
	}

	// Step 5: Get Response input items (if supported)
	if supportsGet {
		getResponseInputItems(t, client, ctx, secondResponseID, model)
		t.Logf("✓ Retrieved response input items")
	} else {
		t.Logf("- Skipped getting response input items (not supported by model)")
	}

	// Step 6: Delete Responses (cleanup)
	deleteResponse(t, client, ctx, secondResponseID, model)
	t.Logf("✓ Deleted second response")

	deleteResponse(t, client, ctx, firstResponseID, model)
	t.Logf("✓ Deleted first response")

	t.Logf("✓ Non-stream responses workflow completed successfully for model: %s", model)
}

// testResponsesStreamingForModel tests responses streaming mode for single model
func testResponsesStreamingForModel(t *testing.T, client *common.Client, model string) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	if !supportsStreaming(model) {
		t.Skipf("Model %s does not support streaming, skipping streaming tests", model)
		return
	}

	t.Logf("Testing streaming responses for model: %s", model)

	// Test streaming response creation
	responseID := createStreamingResponse(t, client, ctx, model, "Write a simple Python function to calculate factorial")
	t.Logf("✓ Created streaming response ID: %s", responseID)

	// Cleanup: delete created response (if valid response ID exists)
	if responseID != "streaming_test_fallback_id" {
		deleteResponse(t, client, ctx, responseID, model)
		t.Logf("✓ Deleted streaming response")
	} else {
		t.Logf("- Skipped deleting response (no valid response ID received)")
	}

	t.Logf("✓ Streaming responses workflow completed successfully for model: %s", model)
}

// createStreamingResponse creates streaming Response
func createStreamingResponse(t *testing.T, client *common.Client, ctx context.Context, model, input string) string {
	createRequest := map[string]any{
		"model":  model,
		"input":  input,
		"stream": true, // Enable streaming
	}

	// Only add metadata field for models that support metadata
	if supportsMetadata(model) {
		createRequest["metadata"] = map[string]any{
			"test": "integration_test_streaming",
		}
	}

	var responseID string
	var content strings.Builder
	var streamCount int
	var chunkTimes []time.Time

	resp := client.PostJSONStream(ctx, "/v1/responses", createRequest, func(data []byte) error {
		chunkTimes = append(chunkTimes, time.Now())

		// Parse streaming response
		var streamResp map[string]any
		if err := json.Unmarshal(data, &streamResp); err != nil {
			return fmt.Errorf("failed to parse stream response: %w", err)
		}

		streamCount++

		// Extract response ID (Responses API format is response.id)
		if responseObj, ok := streamResp["response"].(map[string]any); ok {
			if id, ok := responseObj["id"].(string); ok && responseID == "" {
				responseID = id
				t.Logf("   Extracted response ID: %s", id)
			}
		}

		// Log object type (if exists)
		if streamType, ok := streamResp["type"].(string); ok {
			if streamCount == 1 {
				t.Logf("   Stream type: %s", streamType)
			}
		}

		// Extract and accumulate content
		// Try multiple possible content paths
		extractedContent := extractContentFromStreamResponse(streamResp)
		if extractedContent != "" {
			content.WriteString(extractedContent)
		}

		return nil
	})

	if resp.Error != nil {
		t.Fatalf("✗ Streaming response request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Streaming response request failed with status %d", resp.StatusCode)
	}

	// Validate response headers
	contentType := resp.Headers.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/event-stream") {
		t.Errorf("✗ Expected Content-Type starting with text/event-stream, got: %s", contentType)
	}

	if streamCount == 0 {
		t.Error("✗ No stream chunks received")
	}

	if responseID == "" {
		t.Error("✗ No response ID received from stream")
		// Generate fake ID to avoid subsequent test crashes
		responseID = "streaming_test_fallback_id"
	}

	// Analyze streaming timing characteristics
	analyzeStreamingTiming(t, chunkTimes, model)

	contentStr := strings.TrimSpace(content.String())
	t.Logf("✓ Streaming Model %s: %d chunks", model, streamCount)
	t.Logf("  Content preview: %s...", truncateString(contentStr, 100))
	t.Logf("  Input: %s", truncateString(input, 50))

	return responseID
}

// extractContentFromStreamResponse extracts content from streaming response
func extractContentFromStreamResponse(streamResp map[string]any) string {
	var content strings.Builder

	// Responses API streaming format: extract text content from delta
	if delta, ok := streamResp["delta"].(map[string]any); ok {
		if text, ok := delta["text"].(string); ok {
			content.WriteString(text)
		}
	}

	// Try to extract from content_part (if exists)
	if contentPart, ok := streamResp["content_part"].(map[string]any); ok {
		if text, ok := contentPart["text"].(string); ok {
			content.WriteString(text)
		}
	}

	// Extract content from response.output (content for non-streaming events)
	if responseObj, ok := streamResp["response"].(map[string]any); ok {
		if output, ok := responseObj["output"].([]any); ok {
			for _, item := range output {
				if itemMap, ok := item.(map[string]any); ok {
					if contentArr, ok := itemMap["content"].([]any); ok {
						for _, contentItem := range contentArr {
							if contentMap, ok := contentItem.(map[string]any); ok {
								if text, ok := contentMap["text"].(string); ok {
									content.WriteString(text)
								}
							}
						}
					}
				}
			}
		}
	}

	return content.String()
}

// createNonStreamResponse creates first non-streaming Response
func createNonStreamResponse(t *testing.T, client *common.Client, ctx context.Context, model, input string) string {
	createRequest := map[string]any{
		"model":  model,
		"input":  input,
		"stream": false, // Explicitly specify as non-streaming
	}

	// Only add metadata field for models that support metadata
	if supportsMetadata(model) {
		createRequest["metadata"] = map[string]any{
			"test": "integration_test",
		}
	}

	startTime := time.Now()
	resp := client.PostJSON(ctx, "/v1/responses", createRequest)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Create non-stream response failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Create non-stream response failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var responseData map[string]any
	if err := resp.GetJSON(&responseData); err != nil {
		t.Fatalf("✗ Failed to parse non-stream response: %v", err)
	}

	// Validate response
	responseID, ok := responseData["id"].(string)
	if !ok || responseID == "" {
		t.Fatalf("✗ Response ID is empty or invalid: %v", responseData["id"])
	}

	if object, ok := responseData["object"].(string); !ok || object != "response" {
		t.Errorf("✗ Expected object 'response', got '%v'", responseData["object"])
	}

	// Check status field (if exists)
	if status, ok := responseData["status"].(string); ok {
		t.Logf("   Non-stream response status: %s", status)
	}

	t.Logf("   Non-stream response creation time: %v", responseTime)
	t.Logf("   Input: %s", truncateString(input, 50))
	return responseID
}

// createChainedNonStreamResponse creates chained conversation non-streaming Response
func createChainedNonStreamResponse(t *testing.T, client *common.Client, ctx context.Context, model, input, previousResponseID string) string {
	createRequest := map[string]any{
		"model":                model,
		"input":                input,
		"previous_response_id": previousResponseID, // Use correct parameter for conversation chaining
		"stream":               false,              // Explicitly specify as non-streaming
	}

	// Only add metadata field for models that support metadata
	if supportsMetadata(model) {
		createRequest["metadata"] = map[string]any{
			"test": "integration_test",
			"type": "chained_conversation",
		}
	}

	startTime := time.Now()
	resp := client.PostJSON(ctx, "/v1/responses", createRequest)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Create chained non-stream response failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Create chained non-stream response failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var responseData map[string]any
	if err := resp.GetJSON(&responseData); err != nil {
		t.Fatalf("✗ Failed to parse chained non-stream response: %v", err)
	}

	// Validate response
	responseID, ok := responseData["id"].(string)
	if !ok || responseID == "" {
		t.Fatalf("✗ Chained non-stream response ID is empty or invalid: %v", responseData["id"])
	}

	if object, ok := responseData["object"].(string); !ok || object != "response" {
		t.Errorf("✗ Expected object 'response', got '%v'", responseData["object"])
	}

	// Check status field (if exists)
	if status, ok := responseData["status"].(string); ok {
		t.Logf("   Chained non-stream response status: %s", status)
	}

	t.Logf("   Chained non-stream response creation time: %v", responseTime)
	t.Logf("   Input: %s", truncateString(input, 50))
	t.Logf("   Previous response ID: %s", previousResponseID)
	return responseID
}

// createResponse creates first Response
func createResponse(t *testing.T, client *common.Client, ctx context.Context, model, input string) string {
	createRequest := map[string]any{
		"model": model,
		"input": input,
	}

	// Only add metadata field for models that support metadata
	if supportsMetadata(model) {
		createRequest["metadata"] = map[string]any{
			"test": "integration_test",
		}
	}

	startTime := time.Now()
	resp := client.PostJSON(ctx, "/v1/responses", createRequest)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Create response failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Create response failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var responseData map[string]any
	if err := resp.GetJSON(&responseData); err != nil {
		t.Fatalf("✗ Failed to parse response: %v", err)
	}

	// Validate response
	responseID, ok := responseData["id"].(string)
	if !ok || responseID == "" {
		t.Fatalf("✗ Response ID is empty or invalid: %v", responseData["id"])
	}

	if object, ok := responseData["object"].(string); !ok || object != "response" {
		t.Errorf("✗ Expected object 'response', got '%v'", responseData["object"])
	}

	// Check status field (if exists)
	if status, ok := responseData["status"].(string); ok {
		t.Logf("   Response status: %s", status)
	}

	t.Logf("   Response creation time: %v", responseTime)
	t.Logf("   Input: %s", truncateString(input, 50))
	return responseID
}

// createChainedResponse creates chained conversation Response
func createChainedResponse(t *testing.T, client *common.Client, ctx context.Context, model, input, previousResponseID string) string {
	createRequest := map[string]any{
		"model":                model,
		"input":                input,
		"previous_response_id": previousResponseID, // Use correct parameter for conversation chaining
	}

	// Only add metadata field for models that support metadata
	if supportsMetadata(model) {
		createRequest["metadata"] = map[string]any{
			"test": "integration_test",
			"type": "chained_conversation",
		}
	}

	startTime := time.Now()
	resp := client.PostJSON(ctx, "/v1/responses", createRequest)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Create chained response failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Create chained response failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var responseData map[string]any
	if err := resp.GetJSON(&responseData); err != nil {
		t.Fatalf("✗ Failed to parse chained response: %v", err)
	}

	// Validate response
	responseID, ok := responseData["id"].(string)
	if !ok || responseID == "" {
		t.Fatalf("✗ Chained response ID is empty or invalid: %v", responseData["id"])
	}

	if object, ok := responseData["object"].(string); !ok || object != "response" {
		t.Errorf("✗ Expected object 'response', got '%v'", responseData["object"])
	}

	// Check status field (if exists)
	if status, ok := responseData["status"].(string); ok {
		t.Logf("   Chained response status: %s", status)
	}

	t.Logf("   Chained response creation time: %v", responseTime)
	t.Logf("   Input: %s", truncateString(input, 50))
	t.Logf("   Previous response ID: %s", previousResponseID)
	return responseID
}

// getResponse gets Response details
func getResponse(t *testing.T, client *common.Client, ctx context.Context, responseID, model string) {
	url := fmt.Sprintf("/v1/responses/%s", responseID)

	// Use header to pass model name
	headers := map[string]string{
		"x-ai-proxy-model-name": model,
	}

	startTime := time.Now()
	resp := client.GetWithHeaders(ctx, url, headers)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Get response failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Get response failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var responseData map[string]any
	if err := resp.GetJSON(&responseData); err != nil {
		t.Fatalf("✗ Failed to parse response: %v", err)
	}

	// Validate response
	if id, ok := responseData["id"].(string); !ok || id != responseID {
		t.Errorf("✗ Response ID mismatch, expected '%s', got '%v'", responseID, responseData["id"])
	}

	if object, ok := responseData["object"].(string); !ok || object != "response" {
		t.Errorf("✗ Expected object 'response', got '%v'", responseData["object"])
	}

	// Check status field (if exists)
	if status, ok := responseData["status"].(string); ok {
		t.Logf("   Response status: %s", status)
	}

	// Check model field (if exists)
	if model, ok := responseData["model"].(string); ok {
		t.Logf("   Response model: %s", model)
	}

	// Check output content preview (if exists)
	if output, ok := responseData["output"].(string); ok && output != "" {
		t.Logf("   Output preview: %s...", truncateString(output, 100))
	}

	t.Logf("   Get response time: %v", responseTime)
}

// getResponseInputItems gets Response input items
func getResponseInputItems(t *testing.T, client *common.Client, ctx context.Context, responseID, model string) {
	url := fmt.Sprintf("/v1/responses/%s/input_items", responseID)

	// Use header to pass model name
	headers := map[string]string{
		"x-ai-proxy-model-name": model,
	}

	startTime := time.Now()
	resp := client.GetWithHeaders(ctx, url, headers)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Get response input items failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Get response input items failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var inputItemsData map[string]any
	if err := resp.GetJSON(&inputItemsData); err != nil {
		t.Fatalf("✗ Failed to parse input items response: %v", err)
	}

	// Validate response format
	if object, ok := inputItemsData["object"].(string); !ok || !strings.Contains(object, "list") {
		t.Errorf("✗ Expected object containing 'list', got '%v'", inputItemsData["object"])
	}

	// Check data field
	if data, ok := inputItemsData["data"].([]any); ok {
		t.Logf("   Found %d input items", len(data))

		// Log preview of first item (if exists)
		if len(data) > 0 {
			if firstItem, ok := data[0].(map[string]any); ok {
				if itemType, exists := firstItem["type"]; exists {
					t.Logf("   First item type: %v", itemType)
				}
				if content, exists := firstItem["content"]; exists {
					t.Logf("   First item content preview: %v", truncateString(fmt.Sprintf("%v", content), 80))
				}
			}
		}
	} else {
		t.Logf("   No input items data found or invalid format")
	}

	t.Logf("   Get input items time: %v", responseTime)
}

// deleteResponse deletes Response
func deleteResponse(t *testing.T, client *common.Client, ctx context.Context, responseID, model string) {
	url := fmt.Sprintf("/v1/responses/%s", responseID)

	// Use header to pass model name
	headers := map[string]string{
		"x-ai-proxy-model-name": model,
	}

	startTime := time.Now()
	resp := client.DeleteWithHeaders(ctx, url, headers)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Delete response failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Delete response failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var deleteResult map[string]any
	if err := resp.GetJSON(&deleteResult); err != nil {
		t.Fatalf("✗ Failed to parse delete response: %v", err)
	}

	// Validate deletion result
	if id, ok := deleteResult["id"].(string); !ok || id != responseID {
		t.Errorf("✗ Delete response ID mismatch, expected '%s', got '%v'", responseID, deleteResult["id"])
	}

	if deleted, ok := deleteResult["deleted"].(bool); !ok || !deleted {
		t.Errorf("✗ Response not marked as deleted: %v", deleteResult["deleted"])
	}

	// No longer check object field as different models return different formats
	// gpt-4o returns "response.deleted", doubao returns "response"

	t.Logf("   Delete response time: %v", responseTime)
}
