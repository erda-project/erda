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

package http_error

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteJSONHTTPError_OpenAIStyleErrorOnly(t *testing.T) {
	rec := httptest.NewRecorder()
	err := &HTTPError{
		Ctx:        context.Background(),
		StatusCode: 400,
		Message:    "input[1]: video_url is required when type=video",
		ErrorCtx: map[string]any{
			"code":    "invalid_request_error",
			"message": "input[1]: video_url is required when type=video",
			"param":   "input[1].video_url",
			"type":    "validation_error",
		},
	}

	err.WriteJSONHTTPError(rec)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	_, hasTopMessage := body["message"]
	require.False(t, hasTopMessage)

	errorBody, ok := body["error"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "invalid_request_error", errorBody["code"])
	require.Equal(t, "validation_error", errorBody["type"])
}

func TestWriteJSONHTTPError_KeepsLegacyEnvelope(t *testing.T) {
	rec := httptest.NewRecorder()
	err := &HTTPError{
		Ctx:        context.Background(),
		StatusCode: 400,
		Message:    "LLM Backend Error",
		ErrorCtx: map[string]any{
			"type": "llm-backend-error",
		},
	}

	err.WriteJSONHTTPError(rec)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "LLM Backend Error", body["message"])
	_, hasError := body["error"]
	require.True(t, hasError)
}
