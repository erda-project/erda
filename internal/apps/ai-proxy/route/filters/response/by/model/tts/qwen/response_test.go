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

package qwen

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQwenTTSConverter_OnHeaders(t *testing.T) {
	f := &QwenTTSConverter{}
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
	}

	err := f.OnHeaders(resp)
	assert.NoError(t, err)
	assert.Equal(t, "audio/mpeg", resp.Header.Get("Content-Type"))

	resp.StatusCode = http.StatusInternalServerError
	err = f.OnHeaders(resp)
	assert.Error(t, err)
}

func TestQwenTTSConverter_OnBodyChunk(t *testing.T) {
	f := &QwenTTSConverter{}
	chunk := []byte("test chunk")

	_, err := f.OnBodyChunk(nil, chunk, 0)
	assert.NoError(t, err)
	assert.Equal(t, "test chunk", f.buff.String())
}

func TestQwenTTSConverter_OnComplete(t *testing.T) {
	// Mock audio server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("audio data"))
	}))
	defer ts.Close()

	tests := []struct {
		name        string
		qwenResp    QwenTTSResponse
		mockResp    string
		expectError bool
		expectAudio []byte
	}{
		{
			name: "Success",
			qwenResp: QwenTTSResponse{
				Output: QwenTTSOutput{
					Audio: QwenTTSAudio{
						URL: ts.URL,
					},
				},
			},
			expectAudio: []byte("audio data"),
		},
		{
			name: "Missing URL",
			qwenResp: QwenTTSResponse{
				Output: QwenTTSOutput{
					Audio: QwenTTSAudio{
						URL: "",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &QwenTTSConverter{}
			resp := &http.Response{
				Request: httptest.NewRequest("GET", "/", nil).WithContext(context.Background()),
			}

			// Pre-fill buffer with mocked JSON response
			jsonBytes, _ := json.Marshal(tt.qwenResp)
			f.buff.Write(jsonBytes)

			out, err := f.OnComplete(resp)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectAudio, out)
			}
		})
	}
}

func TestQwenTTSConverter_OnComplete_DownloadError(t *testing.T) {
	// Server that returns 404
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	f := &QwenTTSConverter{}
	qwenResp := QwenTTSResponse{
		Output: QwenTTSOutput{
			Audio: QwenTTSAudio{
				URL: ts.URL,
			},
		},
	}
	jsonBytes, _ := json.Marshal(qwenResp)
	f.buff.Write(jsonBytes)

	resp := &http.Response{
		Request: httptest.NewRequest("GET", "/", nil).WithContext(context.Background()),
	}

	_, err := f.OnComplete(resp)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bad status")
}
