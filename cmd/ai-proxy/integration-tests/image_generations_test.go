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
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

func TestImageGenerations(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	// Get image models for testing
	imageModels := cfg.ImageModels
	if len(imageModels) == 0 {
		t.Skip("No image models configured for testing")
	}

	for _, model := range imageModels {
		// Test with header method (model passed via X-AI-Proxy-Model header)
		t.Run(fmt.Sprintf("Header_Model_%s", model), func(t *testing.T) {
			testImageGenerationForModelWithHeader(t, client, model)
		})

		// Test with model in request body
		t.Run(fmt.Sprintf("Body_Model_%s", model), func(t *testing.T) {
			testImageGenerationForModelWithBody(t, client, model)
		})
	}
}

func testImageGenerationForModelWithHeader(t *testing.T, client *common.Client, model string) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	// Create image generation request
	request := openai.ImageRequest{
		Prompt: "A simple test image",
		N:      1,
		Size:   "1024x1024",
	}

	// Set different parameters based on model type
	if strings.Contains(strings.ToLower(model), "gpt-image-1") {
		// gpt-image-1: don't set response_format, returns base64
		// Note: don't set Quality and Style as they may not be supported
	} else if strings.Contains(strings.ToLower(model), "dall-e-3") || strings.Contains(strings.ToLower(model), "dalle-3") {
		// DALL-E 3: set response_format to url, supports Quality and Style
		request.ResponseFormat = "url"
		request.Quality = "standard"
		request.Style = "vivid"
	} else {
		// Other models default to url format
		request.ResponseFormat = "url"
	}

	// Pass model via header
	headers := map[string]string{
		"X-AI-Proxy-Model": model,
	}

	t.Logf("Debug: Setting model in header: %s", model)
	t.Logf("Debug: Request params - Size: %s, Format: %s, N: %d", request.Size, request.ResponseFormat, request.N)

	startTime := time.Now()
	resp := client.PostJSONWithHeaders(ctx, "/v1/images/generations", request, headers)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var imageResp openai.ImageResponse
	if err := resp.GetJSON(&imageResp); err != nil {
		t.Fatalf("✗ Failed to parse response: %v", err)
	}

	// Validate response
	validateImageResponse(t, &imageResp, model, responseTime)

	t.Logf("✓ Model %s generated %d image(s) (response time: %v)",
		model, len(imageResp.Data), responseTime)
}

func testImageGenerationForModelWithBody(t *testing.T, client *common.Client, model string) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
	defer cancel()

	// Create image generation request with model included in request body
	request := openai.ImageRequest{
		Prompt: "A simple test image",
		Model:  model,
		N:      1,
		Size:   "1024x1024",
	}

	// Set different parameters based on model type
	if strings.Contains(strings.ToLower(model), "gpt-image-1") {
		// gpt-image-1: don't set response_format, returns base64
		// Note: don't set Quality and Style
	} else if strings.Contains(strings.ToLower(model), "dall-e-3") || strings.Contains(strings.ToLower(model), "dalle-3") {
		// DALL-E 3: set response_format to url, supports Quality and Style
		request.ResponseFormat = "url"
		request.Quality = "standard"
		request.Style = "vivid"
	} else {
		// Other models default to url format
		request.ResponseFormat = "url"
	}

	t.Logf("Debug: Setting model in request body: %s", model)
	t.Logf("Debug: Request params - Size: %s, Format: %s, N: %d", request.Size, request.ResponseFormat, request.N)

	startTime := time.Now()
	resp := client.PostJSON(ctx, "/v1/images/generations", request)
	responseTime := time.Since(startTime)

	if resp.Error != nil {
		t.Fatalf("✗ Request failed: %v", resp.Error)
	}

	if !resp.IsSuccess() {
		t.Fatalf("✗ Request failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response
	var imageResp openai.ImageResponse
	if err := resp.GetJSON(&imageResp); err != nil {
		t.Fatalf("✗ Failed to parse response: %v", err)
	}

	// Validate response
	validateImageResponse(t, &imageResp, model, responseTime)

	t.Logf("✓ Model %s generated %d image(s) with model in request body (response time: %v)",
		model, len(imageResp.Data), responseTime)
}

func validateImageResponse(t *testing.T, imageResp *openai.ImageResponse, model string, responseTime time.Duration) {
	// Validate basic response structure
	if len(imageResp.Data) == 0 {
		t.Error("✗ Empty image response data")
		return
	}

	// Validate image count
	expectedCount := 1
	if len(imageResp.Data) != expectedCount {
		t.Errorf("✗ Expected %d images, got %d", expectedCount, len(imageResp.Data))
	}

	// Validate each image data
	for i, imageData := range imageResp.Data {
		// Validate different response formats based on model type
		if strings.Contains(strings.ToLower(model), "gpt-image-1") {
			// gpt-image-1 should return base64 data
			if imageData.B64JSON == "" {
				t.Errorf("✗ Empty base64 image data for image %d from gpt-image-1", i)
			} else {
				// Validate base64 data format
				if validateBase64ImageData(imageData.B64JSON) {
					t.Logf("  ✓ Image %d base64 data is valid (%d bytes)", i, len(imageData.B64JSON))
				} else {
					t.Errorf("✗ Invalid base64 image data for image %d", i)
				}
			}
		} else {
			// Other models should return URL
			if imageData.URL == "" {
				t.Errorf("✗ Empty image URL for image %d", i)
			} else {
				// Validate URL format
				if !strings.HasPrefix(imageData.URL, "http") {
					t.Errorf("✗ Invalid image URL format for image %d: %s", i, imageData.URL)
				} else {
					// Optional: test URL accessibility
					if validateImageURL(imageData.URL) {
						t.Logf("  ✓ Image %d URL is accessible: %s", i, truncateString(imageData.URL, 80))
					} else {
						t.Logf("  ⚠ Image %d URL may not be immediately accessible: %s", i, truncateString(imageData.URL, 80))
					}
				}
			}
		}

		// Validate revised prompt (if exists)
		if imageData.RevisedPrompt != "" {
			t.Logf("  → Revised prompt for image %d: %s", i, truncateString(imageData.RevisedPrompt, 100))
		}
	}

	// Check response time reasonableness
	if responseTime > 60*time.Second {
		t.Logf("⚠ Slow image generation response time: %v", responseTime)
	}

	// Validate creation timestamp
	if imageResp.Created == 0 {
		t.Logf("⚠ Missing creation timestamp in response")
	}
}

// TestImageGenerationsErrorHandling tests error handling
func TestImageGenerationsErrorHandling(t *testing.T) {
	cfg := config.Get()
	client := common.NewClient()

	imageModels := cfg.ImageModels
	if len(imageModels) == 0 {
		t.Skip("No image models configured for testing")
	}

	model := imageModels[0] // Use first model for error testing

	t.Run("EmptyPrompt", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		request := openai.ImageRequest{
			Prompt: "",
			N:      1,
			Size:   "1024x1024",
		}

		headers := map[string]string{
			"X-AI-Proxy-Model": model,
		}
		resp := client.PostJSONWithHeaders(ctx, "/v1/images/generations", request, headers)

		// Expect this request to fail
		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with empty prompt, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected empty prompt (status: %d)", resp.StatusCode)
		}
	})

	t.Run("InvalidSize", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		request := openai.ImageRequest{
			Prompt: "A simple test image",
			N:      1,
			Size:   "999x999", // Invalid size
		}

		headers := map[string]string{
			"X-AI-Proxy-Model": model,
		}
		resp := client.PostJSONWithHeaders(ctx, "/v1/images/generations", request, headers)

		// Expect this request to fail
		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with invalid size, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected invalid size (status: %d)", resp.StatusCode)
		}
	})

	t.Run("InvalidModel", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		request := openai.ImageRequest{
			Prompt: "A simple test image",
			N:      1,
			Size:   "1024x1024",
		}

		headers := map[string]string{
			"X-AI-Proxy-Model": "non-existent-image-model",
		}
		resp := client.PostJSONWithHeaders(ctx, "/v1/images/generations", request, headers)

		// Expect this request to fail
		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with invalid model, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected invalid model (status: %d)", resp.StatusCode)
		}
	})

	t.Run("TooManyImages", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), config.Get().Timeout)
		defer cancel()

		request := openai.ImageRequest{
			Prompt: "A simple test image",
			N:      11, // Exceeds maximum allowed count
			Size:   "1024x1024",
		}

		headers := map[string]string{
			"X-AI-Proxy-Model": model,
		}
		resp := client.PostJSONWithHeaders(ctx, "/v1/images/generations", request, headers)

		// Expect this request to fail
		if resp.IsSuccess() {
			t.Error("✗ Expected request to fail with too many images requested, but it succeeded")
		} else {
			t.Logf("✓ Correctly rejected too many images request (status: %d)", resp.StatusCode)
		}
	})
}

// validateImageURL checks if image URL is accessible
func validateImageURL(url string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Head(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// validateBase64ImageData checks if base64 image data is valid
func validateBase64ImageData(data string) bool {
	if data == "" {
		return false
	}

	// Try to decode base64 data
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return false
	}

	// Check if decoded data has reasonable size (at least 100 bytes)
	return len(decoded) > 100
}
