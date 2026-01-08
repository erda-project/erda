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

package volcengine_ark

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"

	metadatapb "github.com/erda-project/erda-proto-go/apps/aiproxy/metadata/pb"
	modelpb "github.com/erda-project/erda-proto-go/apps/aiproxy/model/pb"
	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
)

func TestVolcengineTTSResponseConverter_OnPeekChunkBeforeHeaders(t *testing.T) {
	// Case 1: Success path logic check
	// Note: This test will fail to poll real Bytedance API, which is expected.
	// We verify that it tries to poll (and fails), which means it passed the TaskID check.
	f1 := &VolcengineTTSConverter{}

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(ctxhelper.InitCtxMapIfNeed(req.Context()))

	// Mock Model and CallID
	mockMeta, _ := structpb.NewStruct(map[string]interface{}{
		"app_id":     "mock_app",
		"access_key": "mock_key",
		"model_name": "mock_model",
	})
	ctxhelper.PutModel(req.Context(), &modelpb.Model{
		Metadata: &metadatapb.Metadata{
			Public: mockMeta.Fields,
		},
	})
	ctxhelper.PutGeneratedCallID(req.Context(), "mock-call-id")

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Request:    req,
	}
	seekedBody := []byte(`{"code":20000000,"message":"success","data":{"task_id":"test-task-123"}}`)

	err := f1.OnPeekChunkBeforeHeaders(resp, seekedBody)
	assert.Error(t, err) // Expect network error
	assert.Contains(t, err.Error(), "failed to poll audio")

	// Case 2: Error response
	f2 := &VolcengineTTSConverter{}
	resp2 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Request:    httptest.NewRequest("GET", "/", nil),
	}
	seekedBody2 := []byte(`{"code":40000001,"message":"error","data":{}}`)

	err = f2.OnPeekChunkBeforeHeaders(resp2, seekedBody2)
	assert.NoError(t, err)                                // Returns nil
	assert.Equal(t, "", resp2.Header.Get("Content-Type")) // Should NOT be audio

	// Case 3: Invalid JSON
	f3 := &VolcengineTTSConverter{}
	resp3 := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Request:    httptest.NewRequest("GET", "/", nil),
	}
	seekedBody3 := []byte(`invalid json`)
	err = f3.OnPeekChunkBeforeHeaders(resp3, seekedBody3)
	assert.Error(t, err)
}

func TestVolcengineTTSResponseConverter_OnBodyChunk(t *testing.T) {
	f := &VolcengineTTSConverter{}
	chunk := []byte("test chunk")

	req := httptest.NewRequest("GET", "/", nil)
	req = req.WithContext(ctxhelper.InitCtxMapIfNeed(req.Context()))
	// ctxhelper.PutIsLastBodyChunk(req.Context(), false) // Simulate NOT last chunk

	resp := &http.Response{
		Request: req,
	}

	out, err := f.OnBodyChunk(resp, chunk, 0)
	assert.NoError(t, err)
	assert.Equal(t, chunk, out)
}

func TestBytedanceTTSSubmitResponse_Parsing(t *testing.T) {
	tests := []struct {
		name        string
		jsonData    string
		expectCode  int64
		expectMsg   string
		expectTask  string
		expectError bool
	}{
		{
			name:       "Success response",
			jsonData:   `{"code":20000000,"message":"success","data":{"task_id":"test-task-123"}}`,
			expectCode: 20000000,
			expectMsg:  "success",
			expectTask: "test-task-123",
		},
		{
			name:        "Error response",
			jsonData:    `{"code":40000001,"message":"invalid request","data":{}}`,
			expectCode:  40000001,
			expectMsg:   "invalid request",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp BytedanceTTSSubmitResponse
			err := json.Unmarshal([]byte(tt.jsonData), &resp)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectCode, resp.Code)
			assert.Equal(t, tt.expectMsg, resp.Message)
			if tt.expectTask != "" {
				assert.Equal(t, tt.expectTask, resp.Data.TaskId)
			}
		})
	}
}

func TestQueryResponse_Parsing(t *testing.T) {
	tests := []struct {
		name         string
		jsonData     string
		expectCode   int64
		expectStatus int64
		expectURL    string
	}{
		{
			name:         "Running status",
			jsonData:     `{"code":20000000,"message":"success","data":{"task_status":1,"audio_url":""}}`,
			expectCode:   20000000,
			expectStatus: 1,
			expectURL:    "",
		},
		{
			name:         "Success status with URL",
			jsonData:     `{"code":20000000,"message":"success","data":{"task_status":2,"audio_url":"https://example.com/audio.mp3"}}`,
			expectCode:   20000000,
			expectStatus: 2,
			expectURL:    "https://example.com/audio.mp3",
		},
		{
			name:         "Failed status",
			jsonData:     `{"code":20000000,"message":"task failed","data":{"task_status":3,"audio_url":""}}`,
			expectCode:   20000000,
			expectStatus: 3,
			expectURL:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp QueryResponse
			err := json.Unmarshal([]byte(tt.jsonData), &resp)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectCode, resp.Code)
			assert.Equal(t, tt.expectStatus, resp.Data.TaskStatus)
			assert.Equal(t, tt.expectURL, resp.Data.AudioURL)
		})
	}
}
