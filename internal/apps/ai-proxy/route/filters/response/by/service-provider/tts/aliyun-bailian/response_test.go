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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQwenTTSConverter_OnPeekChunkBeforeHeaders(t *testing.T) {
	tests := []struct {
		name           string
		jsonBody       string
		expectError    bool
		expectHeader   string
		expectErrorMsg string
	}{
		{
			name:         "Success - extracts audio URL and sets header",
			jsonBody:     `{"output":{"audio":{"url":"http://example.com/audio.mp3"},"finish_reason":"stop"}}`,
			expectError:  false,
			expectHeader: "audio/mpeg",
		},
		{
			name:           "Missing URL - error response",
			jsonBody:       `{"code":"InvalidParameter","message":"error"}`,
			expectError:    true,
			expectErrorMsg: "missing audio url",
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
				assert.NotEmpty(t, f.audioURL) // audioURL should be extracted
			}
		})
	}
}

func TestQwenTTSConverter_OnBodyChunk(t *testing.T) {
	// Mock audio server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("audio data"))
	}))
	defer ts.Close()

	// Test success case
	f := &BailianTTSConverter{audioURL: ts.URL}
	req := httptest.NewRequest("GET", "/", nil)
	resp := &http.Response{
		Request: req,
	}

	out, err := f.OnBodyChunk(resp, []byte("original chunk"), 0)
	assert.NoError(t, err)
	assert.Equal(t, "audio data", string(out))

	// Test download error case
	ts404 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts404.Close()

	f2 := &BailianTTSConverter{audioURL: ts404.URL}
	out2, err2 := f2.OnBodyChunk(resp, []byte("original chunk"), 0)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "failed to download audio")
	assert.Nil(t, out2)
}
