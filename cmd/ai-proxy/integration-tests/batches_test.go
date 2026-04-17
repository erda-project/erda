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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/common"
	"github.com/erda-project/erda/cmd/ai-proxy/integration-tests/config"
)

type fileUploadResponse struct {
	ID       string `json:"id"`
	Object   string `json:"object"`
	Filename string `json:"filename"`
	Purpose  string `json:"purpose"`
}

type batchResponse struct {
	ID               string `json:"id"`
	Object           string `json:"object"`
	InputFileID      string `json:"input_file_id"`
	OutputFileID     string `json:"output_file_id"`
	ErrorFileID      string `json:"error_file_id"`
	Status           string `json:"status"`
	Endpoint         string `json:"endpoint"`
	CompletionWindow string `json:"completion_window"`
}

func TestBatchesWorkflow(t *testing.T) {
	cfg := config.Get()
	if len(cfg.BatchModels) == 0 {
		t.Skip("No batch models configured (BATCH_MODELS)")
	}

	client := common.NewClient()

	for _, model := range cfg.BatchModels {
		model := model
		t.Run(fmt.Sprintf("Model_%s", model), func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
			defer cancel()

			headers := map[string]string{
				"X-AI-Proxy-Model": model,
			}

			fileID := uploadBatchInputFile(t, client, ctx, model, headers)
			batchID := createBatch(t, client, ctx, fileID, cfg.BatchWindow, headers)

			getBatch(t, client, ctx, batchID, headers)
			listBatches(t, client, ctx, headers)
			cancelBatch(t, client, ctx, batchID, headers)
		})
	}
}

func uploadBatchInputFile(t *testing.T, client *common.Client, ctx context.Context, model string, headers map[string]string) string {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("purpose", "batch"); err != nil {
		t.Fatalf("failed to write purpose field: %v", err)
	}

	part, err := writer.CreateFormFile("file", "batch-input.jsonl")
	if err != nil {
		t.Fatalf("failed to create file part: %v", err)
	}

	jsonlLine := fmt.Sprintf(
		"{\"custom_id\":\"case-1\",\"method\":\"POST\",\"url\":\"/v1/chat/completions\",\"body\":{\"model\":\"%s\",\"messages\":[{\"role\":\"user\",\"content\":\"Reply exactly: ok\"}],\"max_tokens\":16}}\n",
		model,
	)
	if _, err = part.Write([]byte(jsonlLine)); err != nil {
		t.Fatalf("failed to write JSONL content: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	resp := client.PostMultipartWithHeaders(ctx, "/v1/files", &body, writer.FormDataContentType(), headers)
	if resp.Error != nil {
		t.Fatalf("upload file failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Fatalf("upload file failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var fileResp fileUploadResponse
	if err := resp.GetJSON(&fileResp); err != nil {
		t.Fatalf("failed to parse file upload response: %v", err)
	}

	if fileResp.ID == "" {
		t.Fatalf("file upload response missing id: %s", string(resp.Body))
	}
	if fileResp.Object != "file" {
		t.Fatalf("expected file object, got %q", fileResp.Object)
	}

	t.Logf("✓ Uploaded batch file for model %s: %s", model, fileResp.ID)
	return fileResp.ID
}

func createBatch(t *testing.T, client *common.Client, ctx context.Context, fileID, completionWindow string, headers map[string]string) string {
	t.Helper()

	payload := map[string]any{
		"input_file_id":     fileID,
		"endpoint":          "/v1/chat/completions",
		"completion_window": completionWindow,
	}

	resp := client.PostJSONWithHeaders(ctx, "/v1/batches", payload, headers)
	if resp.Error != nil {
		t.Fatalf("create batch failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Fatalf("create batch failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var batchResp batchResponse
	if err := resp.GetJSON(&batchResp); err != nil {
		t.Fatalf("failed to parse create batch response: %v", err)
	}
	if batchResp.ID == "" {
		t.Fatalf("create batch response missing id: %s", string(resp.Body))
	}
	if batchResp.Object != "batch" {
		t.Fatalf("expected batch object, got %q", batchResp.Object)
	}
	if batchResp.InputFileID != fileID {
		t.Fatalf("input_file_id mismatch, expected %s got %s", fileID, batchResp.InputFileID)
	}

	t.Logf("✓ Created batch: %s (status=%s)", batchResp.ID, batchResp.Status)
	return batchResp.ID
}

func getBatch(t *testing.T, client *common.Client, ctx context.Context, batchID string, headers map[string]string) {
	t.Helper()

	resp := client.GetWithHeaders(ctx, "/v1/batches/"+batchID, headers)
	if resp.Error != nil {
		t.Fatalf("get batch failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Fatalf("get batch failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var batchResp batchResponse
	if err := resp.GetJSON(&batchResp); err != nil {
		t.Fatalf("failed to parse get batch response: %v", err)
	}
	if batchResp.ID != batchID {
		t.Fatalf("retrieved batch id mismatch, expected %s got %s", batchID, batchResp.ID)
	}

	t.Logf("✓ Retrieved batch: %s (status=%s)", batchResp.ID, batchResp.Status)
}

func listBatches(t *testing.T, client *common.Client, ctx context.Context, headers map[string]string) {
	t.Helper()

	resp := client.GetWithHeaders(ctx, "/v1/batches", headers)
	if resp.Error != nil {
		t.Fatalf("list batches failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Fatalf("list batches failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var listResp map[string]any
	if err := json.Unmarshal(resp.Body, &listResp); err != nil {
		t.Fatalf("failed to parse list batches response: %v", err)
	}

	objectVal, _ := listResp["object"].(string)
	if objectVal != "" && objectVal != "list" {
		t.Fatalf("unexpected list object value: %q", objectVal)
	}
	if _, ok := listResp["data"]; !ok {
		t.Fatalf("list batches response missing data field: %s", string(resp.Body))
	}

	t.Logf("✓ Listed batches")
}

func cancelBatch(t *testing.T, client *common.Client, ctx context.Context, batchID string, headers map[string]string) {
	t.Helper()

	resp := client.PostJSONWithHeaders(ctx, "/v1/batches/"+batchID+"/cancel", nil, headers)
	if resp.Error != nil {
		t.Fatalf("cancel batch failed: %v", resp.Error)
	}
	if !resp.IsSuccess() {
		t.Fatalf("cancel batch failed with status %d: %s", resp.StatusCode, string(resp.Body))
	}

	var batchResp batchResponse
	if err := resp.GetJSON(&batchResp); err != nil {
		t.Fatalf("failed to parse cancel batch response: %v", err)
	}
	if batchResp.ID != batchID {
		t.Fatalf("cancel batch id mismatch, expected %s got %s", batchID, batchResp.ID)
	}
	if batchResp.Status == "" {
		t.Fatalf("cancel batch response missing status: %s", string(resp.Body))
	}

	if !strings.Contains(strings.ToLower(batchResp.Status), "cancel") && !strings.Contains(strings.ToLower(batchResp.Status), "final") {
		// provider implementations may return terminal status values directly.
		t.Logf("cancel response status=%s", batchResp.Status)
	}

	t.Logf("✓ Cancelled batch: %s (status=%s)", batchResp.ID, batchResp.Status)
}

