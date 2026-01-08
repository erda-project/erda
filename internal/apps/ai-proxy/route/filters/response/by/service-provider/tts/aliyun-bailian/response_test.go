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

package aliyun_bailian

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQwenTTSConverter_OnPeekChunkBeforeHeaders(t *testing.T) {
	// Mock audio server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("audio data"))
	}))
	defer ts.Close()

	// Mock 404 server
	ts404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts404.Close()

	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectHeader   string
		expectBody     string
		expectErrorMsg string
	}{
		{
			name:         "Success",
			jsonBody:     `{"output":{"audio":{"url":"` + ts.URL + `"},"finish_reason":"stop"}}`,
			expectError:  false,
			expectHeader: "audio/mpeg",
			expectBody:   "audio data",
		},
		{
			name:         "Missing URL (Error Response)",
			jsonBody:     `{"code":"InvalidParameter","message":"error"}`,
			expectError:  false,
			expectHeader: "",
			expectBody:   "", // Body not replaced
		},
		{
			name:           "Download Error",
			jsonBody:       `{"output":{"audio":{"url":"` + ts404.URL + `"},"finish_reason":"stop"}}`,
			expectError:    true,
			expectErrorMsg: "failed to download audio",
		},
		{
			name:           "Invalid JSON",
			jsonBody:       `{invalid json`,
			expectError:    true,
			expectErrorMsg: "failed to unmarshal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &BailianTTSConverter{}
			req := httptest.NewRequest("GET", "/", nil)
			resp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Request:    req,
				Body:       io.NopCloser(nil), // Initial body, irrelevant for peek logic itself but good for structure
			}

			err := f.OnPeekChunkBeforeHeaders(resp, []byte(tt.jsonBody))

			if tt.expectError {
				assert.Error(t, err)
				if tt.expectErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectErrorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectHeader, resp.Header.Get("Content-Type"))
				if tt.expectBody != "" {
					bodyBytes, _ := io.ReadAll(resp.Body)
					assert.Equal(t, tt.expectBody, string(bodyBytes))
				}
			}
		})
	}
}

func TestQwenTTSConverter_OnBodyChunk(t *testing.T) {
	f := &BailianTTSConverter{}
	chunk := []byte("test chunk")

	out, err := f.OnBodyChunk(nil, chunk, 0)
	assert.NoError(t, err)
	assert.Equal(t, chunk, out)
}
