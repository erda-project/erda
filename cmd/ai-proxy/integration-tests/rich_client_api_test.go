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

	clientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/pb"
	richclientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/rich_client/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	modelproviderpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model_provider/pb"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

// Type aliases for protobuf types to maintain compatibility
type RichClient = richclientpb.RichClient
type Client = clientpb.Client
type RichModel = richclientpb.RichModel
type Model = modelpb.Model
type Provider = modelproviderpb.ModelProvider

// TestRichClientAPI tests all Rich Client API endpoints
func TestRichClientAPI(t *testing.T) {
	cfg := config.Get()
	if cfg.Token == "" {
		t.Skip("No token configured, skipping Rich Client API tests")
	}

	t.Run("GetByAccessKeyId_Success", testGetByAccessKeyIdSuccess)
	t.Run("GetByAccessKeyId_WithoutAccessKeyId", testGetByAccessKeyIdWithoutAccessKeyId)
	t.Run("GetByAccessKeyId_WithInvalidToken", testGetByAccessKeyIdWithInvalidToken)
	t.Run("GetByAccessKeyId_MultiLanguage", testGetByAccessKeyIdMultiLanguage)
}

// testGetByAccessKeyIdSuccess tests successful retrieval of rich client information
func testGetByAccessKeyIdSuccess(t *testing.T) {
	client := common.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with empty access key id (should use token's access key id)
	resp := client.Get(ctx, "/api/ai-proxy/clients/actions/get-by-access-key-id")
	if !resp.IsSuccess() {
		t.Fatalf("Request failed: %v", resp.Error)
	}

	var result struct {
		Success bool        `json:"success"`
		Data    *RichClient `json:"data"`
		Message string      `json:"message"`
	}

	if err := resp.GetJSON(&result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !result.Success {
		t.Fatalf("API returned error: %s", result.Message)
	}

	if result.Data == nil {
		t.Fatal("No data returned")
	}

	// Validate client information
	client_info := result.Data.Client
	if client_info == nil {
		t.Fatal("No client information returned")
	}

	if client_info.Id == "" {
		t.Error("Client ID should not be empty")
	}

	if client_info.Name == "" {
		t.Error("Client name should not be empty")
	}

	// Verify data desensitization: sensitive keys should be empty for security
	if client_info.AccessKeyId != "" {
		t.Errorf("Client AccessKeyId should be empty (desensitized) for security, got: %s", client_info.AccessKeyId)
	} else {
		t.Logf("✓ Client AccessKeyId properly desensitized (empty)")
	}

	if client_info.SecretKeyId != "" {
		t.Errorf("Client SecretKeyId should be empty (desensitized) for security, got: %s", client_info.SecretKeyId)
	} else {
		t.Logf("✓ Client SecretKeyId properly desensitized (empty)")
	}

	// Validate models information
	if len(result.Data.Models) == 0 {
		t.Error("Client should have at least one model")
	}

	for i, model := range result.Data.Models {
		if model.Model == nil {
			t.Errorf("Model %d should not be nil", i)
			continue
		}

		if model.Model.Id == "" {
			t.Errorf("Model %d ID should not be empty", i)
		}

		if model.Model.Name == "" {
			t.Errorf("Model %d name should not be empty", i)
		}

		if model.Provider == nil {
			t.Errorf("Model %d provider should not be nil", i)
			continue
		}

		if model.Provider.Id == "" {
			t.Errorf("Model %d provider ID should not be empty", i)
		}

		if model.Provider.Name == "" {
			t.Errorf("Model %d provider name should not be empty", i)
		}

		// Verify provider ID matches
		if model.Model.ProviderId != model.Provider.Id {
			t.Errorf("Model %d provider ID mismatch: model.providerId=%s, provider.id=%s",
				i, model.Model.ProviderId, model.Provider.Id)
		}

		// Verify sensitive data desensitization
		if model.Model.ApiKey != "" {
			t.Errorf("Model %d ApiKey should be empty (desensitized), got: %s", i, model.Model.ApiKey)
		}

		if model.Provider.ApiKey != "" {
			t.Errorf("Provider %d ApiKey should be empty (desensitized), got: %s", i, model.Provider.ApiKey)
		}
	}

	t.Logf("Successfully retrieved rich client info: client=%s, models=%d",
		client_info.Name, len(result.Data.Models))
}

// testGetByAccessKeyIdWithoutAccessKeyId tests the API without providing access key id
func testGetByAccessKeyIdWithoutAccessKeyId(t *testing.T) {
	client := common.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test with empty request (should use token's access key id)
	resp := client.Get(ctx, "/api/ai-proxy/clients/actions/get-by-access-key-id")
	if !resp.IsSuccess() {
		t.Fatalf("Request should succeed when access key id is empty and token is valid: %v", resp.Error)
	}

	var result struct {
		Success bool        `json:"success"`
		Data    *RichClient `json:"data"`
		Message string      `json:"message"`
	}

	if err := resp.GetJSON(&result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !result.Success {
		t.Fatalf("API should succeed with valid token: %s", result.Message)
	}

	if result.Data == nil {
		t.Fatal("Data should not be nil when token is valid")
	}

	t.Logf("Successfully retrieved client info without explicit access key id")
}

// testGetByAccessKeyIdWithInvalidToken tests API with invalid token
func testGetByAccessKeyIdWithInvalidToken(t *testing.T) {
	cfg := config.Get()
	invalidClient := &common.Client{}
	// Create client with invalid token
	if client := common.NewClient(); client != nil {
		// Create a copy with invalid token
		invalidClient = client
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Override config temporarily for this test
	originalToken := cfg.Token
	cfg.Token = "invalid-token"
	defer func() { cfg.Token = originalToken }()

	resp := invalidClient.GetWithHeaders(ctx, "/api/ai-proxy/clients/actions/get-by-access-key-id", map[string]string{
		"Authorization": "Bearer invalid-token",
	})

	if resp.IsSuccess() {
		t.Error("Request should fail with invalid token")
		return
	}

	if resp.StatusCode != 401 && resp.StatusCode != 403 {
		t.Errorf("Expected 401 or 403 status code, got %d", resp.StatusCode)
	}

	t.Logf("Correctly rejected invalid token with status %d", resp.StatusCode)
}

// testGetByAccessKeyIdMultiLanguage tests multi-language support
func testGetByAccessKeyIdMultiLanguage(t *testing.T) {
	client := common.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test different language headers
	languages := []string{"en", "zh-CN", "zh-TW"}

	for _, lang := range languages {
		t.Run(fmt.Sprintf("Language_%s", lang), func(t *testing.T) {
			resp := client.GetWithHeaders(ctx, "/api/ai-proxy/clients/actions/get-by-access-key-id", map[string]string{
				"Accept-Language": lang,
			})

			if !resp.IsSuccess() {
				t.Fatalf("Request failed for language %s: %v", lang, resp.Error)
			}

			var result struct {
				Success bool        `json:"success"`
				Data    *RichClient `json:"data"`
				Message string      `json:"message"`
			}

			if err := resp.GetJSON(&result); err != nil {
				t.Fatalf("Failed to parse response for language %s: %v", lang, err)
			}

			if !result.Success {
				t.Fatalf("API returned error for language %s: %s", lang, result.Message)
			}

			if result.Data == nil {
				t.Fatalf("No data returned for language %s", lang)
			}

			// Check if metadata is enhanced based on language
			for _, model := range result.Data.Models {
				if model.Model != nil && model.Model.Metadata != nil {
					t.Logf("Language %s: Model %s has enhanced metadata", lang, model.Model.Name)
				}
			}

			t.Logf("Successfully tested language support: %s", lang)
		})
	}
}
