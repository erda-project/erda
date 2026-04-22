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

	"github.com/stretchr/testify/require"
)

func TestOnBodyChunk_ConvertFromArkToCanonical(t *testing.T) {
	input := []byte(`{"created":1776838589,"data":{"embedding":[0.1,0.2],"multi_embedding":[[0.3,0.4]],"object":"embedding","sparse_embedding":[{"index":1,"value":0.5}]},"id":"req-1","model":"doubao-embedding-vision-251215","object":"list","usage":{"prompt_tokens":2,"total_tokens":2}}`)

	f := &VolcengineMultimodalEmbeddingResponseConverter{}
	resp := &http.Response{Request: httptest.NewRequest("POST", "http://example", nil)}
	out, err := f.OnBodyChunk(resp, input, 0)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, jsonUnmarshal(out, &got))
	require.Equal(t, "list", got["object"])

	data, ok := got["data"].([]any)
	require.True(t, ok)
	require.Len(t, data, 1)

	item, ok := data[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "embedding", item["object"])
	require.EqualValues(t, 0, item["index"])
	require.NotNil(t, item["embedding"])
	require.NotNil(t, item["multi_embedding"])
	require.NotNil(t, item["sparse_embedding"])
}

func TestOnBodyChunk_PassThroughWhenAlreadyCanonical(t *testing.T) {
	input := []byte(`{"object":"list","data":[{"object":"embedding","index":0,"embedding":[0.1,0.2]}]}`)
	f := &VolcengineMultimodalEmbeddingResponseConverter{}
	resp := &http.Response{Request: httptest.NewRequest("POST", "http://example", nil)}
	out, err := f.OnBodyChunk(resp, input, 0)
	require.NoError(t, err)
	require.JSONEq(t, string(input), string(out))
}

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
