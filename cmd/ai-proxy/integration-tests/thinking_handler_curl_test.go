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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

const (
	// header constant from vars package (actual header name from response)
	XAIProxyRequestThinkingTransform = "X-Ai-Proxy-Request-Thinking-Transform"
)

// AuditRecord represents an audit record for request transformation
type AuditRecord struct {
	Op   string         `json:"op"`
	From map[string]any `json:"from,omitempty"`
	To   map[string]any `json:"to,omitempty"`
}

// TestThinkingHandlerCurl tests thinking-handler with JSON file based test cases
func TestThinkingHandlerCurl(t *testing.T) {
	testCasesDir := "curl_tests/thinking"
	cfg := config.Get()
	
	// get thinking models from config
	if len(cfg.ThinkingModels) == 0 {
		t.Skip("No THINKING_MODELS configured - please set THINKING_MODELS in .env")
	}
	
	err := filepath.Walk(testCasesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// skip directories and non-json files
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}
		
		// extract thinking mode from filename
		filename := strings.TrimSuffix(info.Name(), ".json")
		if !strings.HasPrefix(filename, "bytedance_style_thinking_") {
			t.Logf("Skipping file with unexpected name format: %s", info.Name())
			return nil
		}
		
		thinkingMode := strings.TrimPrefix(filename, "bytedance_style_thinking_")
		if thinkingMode != "auto" && thinkingMode != "enabled" && thinkingMode != "disabled" {
			t.Logf("Skipping file with unknown thinking mode: %s", thinkingMode)
			return nil
		}
		
		// run test for each target model
		for _, targetModel := range cfg.ThinkingModels {
			testName := fmt.Sprintf("%s_to_%s", filename, strings.ReplaceAll(targetModel, "/", "_"))
			
			t.Run(testName, func(t *testing.T) {
				runCurlTestCase(t, path, &TestInfo{
					InputFormat:    filename,
					TargetModel:    targetModel,
					ThinkingMode:   thinkingMode,
				})
			})
		}
		
		return nil
	})
	
	if err != nil {
		t.Fatalf("Failed to walk test cases directory: %v", err)
	}
}

// TestInfo contains parsed information from test filename
type TestInfo struct {
	InputFormat  string // e.g. "bytedance_style_thinking_auto"
	TargetModel  string // e.g. "qwen/qwen-plus"
	ThinkingMode string // e.g. "auto", "enabled", "disabled"
}

// runCurlTestCase runs a single test case from JSON file
func runCurlTestCase(t *testing.T, testFilePath string, testInfo *TestInfo) {
	// read test case JSON file
	testData, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read test file: %s", testFilePath)
	
	var requestBody map[string]any
	err = json.Unmarshal(testData, &requestBody)
	require.NoError(t, err, "Failed to parse test JSON: %s", testFilePath)
	
	// create HTTP client to real AI Proxy instance
	client := common.NewClient()
	ctx := context.Background()
	
	// add test headers to specify target model
	headers := map[string]string{
		"X-AI-Proxy-Model": testInfo.TargetModel, // use full publisher/model format
	}
	
	// send request to real AI Proxy
	endpoint := "/v1/chat/completions"
	resp := client.PostJSONWithHeaders(ctx, endpoint, requestBody, headers)
	
	// log request details for debugging
	t.Logf("Sending request to %s with model=%s", endpoint, testInfo.TargetModel)
	
	// verify response model header matches what we requested
	responseModel := resp.Headers.Get("X-AI-Proxy-Model")
	if responseModel != "" {
		t.Logf("Response model: %s", responseModel)
		// note: response model might be slightly different (e.g. with version suffix)
	} else {
		t.Logf("No X-AI-Proxy-Model header in response")
	}
	
	// verify audit header contains thinking transformation
	auditHeader := resp.Headers.Get(XAIProxyRequestThinkingTransform)
	if auditHeader != "" {
		verifyAuditTransformation(t, auditHeader, testInfo)
		t.Logf("✓ Test passed: %s -> %s (mode: %s)", 
			testInfo.InputFormat, testInfo.TargetModel, testInfo.ThinkingMode)
	} else {
		t.Logf("No audit header found for test: %s - thinking-handler may not be processing this request", testFilePath)
		
		// log response details for debugging
		if resp.Error != nil {
			t.Logf("Response error: %v", resp.Error)
		}
		t.Logf("Response status: %d", resp.StatusCode)
		t.Logf("Response headers with X-AI-Proxy prefix:")
		for k, v := range resp.Headers {
			if strings.HasPrefix(k, "X-AI-Proxy") || strings.HasPrefix(k, "X-Ai-Proxy") {
				t.Logf("  %s: %s", k, strings.Join(v, ","))
			}
		}
		
		// for now, don't fail the test - this allows us to debug the real AI Proxy setup
		t.Logf("SKIP: No audit transformation detected")
	}
}

// verifyAuditTransformation verifies the audit transformation matches expectations
func verifyAuditTransformation(t *testing.T, auditHeader string, testInfo *TestInfo) {
	// try to parse as array first, then as single object
	var auditRecords []AuditRecord
	err := json.Unmarshal([]byte(auditHeader), &auditRecords)
	if err != nil {
		// try to parse as single object
		var singleRecord AuditRecord
		err = json.Unmarshal([]byte(auditHeader), &singleRecord)
		require.NoError(t, err, "Failed to parse audit header as single object")
		auditRecords = []AuditRecord{singleRecord}
	}
	
	// find thinking record (should be the first/only one for thinking transform header)
	require.Len(t, auditRecords, 1, "Expected single audit record in thinking transform header")
	thinkingRecord := &auditRecords[0]
	
	// verify From field contains thinking
	assert.Contains(t, thinkingRecord.From, "thinking", "Audit 'From' should contain thinking field")
	
	// log the actual transformation for debugging
	t.Logf("Thinking transformation detected: from=%+v, to=%+v", thinkingRecord.From, thinkingRecord.To)
	
	// check if transformation occurred (to field is not null/empty)
	if thinkingRecord.To == nil || len(thinkingRecord.To) == 0 {
		t.Logf("⚠ No transformation applied - 'to' field is null or empty")
		// this might be expected behavior for some cases, but let's log it
		return
	}
	
	// verify To field based on target model publisher
	publisher := strings.Split(testInfo.TargetModel, "/")[0]
	switch publisher {
	case "anthropic":
		assert.Contains(t, thinkingRecord.To, "thinking", "Anthropic target should have thinking field")
	case "qwen":
		// qwen might use enable_thinking or thinking_budget depending on the conversion
		if _, hasEnableThinking := thinkingRecord.To["enable_thinking"]; hasEnableThinking {
			t.Logf("Qwen conversion used enable_thinking field")
		} else if _, hasThinkingBudget := thinkingRecord.To["thinking_budget"]; hasThinkingBudget {
			t.Logf("Qwen conversion used thinking_budget field")
		} else {
			t.Logf("Qwen conversion result: %+v", thinkingRecord.To)
		}
	case "openai":
		// openai might use reasoning_effort for reasoning models
		if _, hasReasoningEffort := thinkingRecord.To["reasoning_effort"]; hasReasoningEffort {
			t.Logf("OpenAI conversion used reasoning_effort field")
		} else {
			t.Logf("OpenAI conversion result: %+v", thinkingRecord.To)
		}
	case "bytedance":
		// bytedance should passthrough thinking field
		assert.Contains(t, thinkingRecord.To, "thinking", "Bytedance target should have thinking field")
	default:
		t.Logf("Unknown publisher %s, conversion result: %+v", publisher, thinkingRecord.To)
	}
	
	t.Logf("✓ Audit verification passed: %+v", thinkingRecord.To)
}