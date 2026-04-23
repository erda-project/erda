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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/erda-project/erda/internal/apps/ai-proxy/common/ctxhelper"
	"github.com/stretchr/testify/require"
)

func TestOnBodyChunk_ConvertFromArkToCanonical(t *testing.T) {
	input := []byte(`{"created":1776838589,"data":{"embedding":[0.1,0.2],"multi_embedding":[[0.3,0.4]],"object":"embedding","sparse_embedding":[{"index":1,"value":0.5}]},"id":"req-1","model":"doubao-embedding-vision-251215","object":"list","usage":{"prompt_tokens":2,"total_tokens":2}}`)

	f := &VolcengineMultimodalEmbeddingResponseConverter{}
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	reqOut := httptest.NewRequest("POST", "http://ark.example", io.NopCloser(bytes.NewReader([]byte(`{"input":[{"type":"text","text":"hi"}]}`))))
	ctxhelper.PutReverseProxyRequestOutSnapshot(ctx, reqOut)
	resp := &http.Response{Request: httptest.NewRequest("POST", "http://example", nil).WithContext(ctx)}
	out, err := f.OnBodyChunk(resp, input, 0)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, jsonUnmarshal(out, &got))
	_, hasObject := got["object"]
	require.False(t, hasObject)
	require.Equal(t, "req-1", got["request_id"])
	_, hasID := got["id"]
	require.False(t, hasID)

	data, ok := got["data"].([]any)
	require.True(t, ok)
	require.Len(t, data, 1)

	item, ok := data[0].(map[string]any)
	require.True(t, ok)
	require.EqualValues(t, 0, item["index"])
	require.Equal(t, "text", item["type"])
	require.NotNil(t, item["embedding"])
	require.NotNil(t, item["multi_embedding"])
	require.NotNil(t, item["sparse_embedding"])

	usage, ok := got["usage"].(map[string]any)
	require.True(t, ok)
	require.EqualValues(t, 2, usage["total_tokens"])
	require.EqualValues(t, 2, usage["input_tokens"])
	require.EqualValues(t, 0, usage["output_tokens"])
}

func TestOnBodyChunk_PassThroughWhenAlreadyCanonical(t *testing.T) {
	input := []byte(`{"object":"list","data":[{"object":"embedding","index":0,"embedding":[0.1,0.2]}]}`)
	f := &VolcengineMultimodalEmbeddingResponseConverter{}
	resp := &http.Response{Request: httptest.NewRequest("POST", "http://example", nil)}
	out, err := f.OnBodyChunk(resp, input, 0)
	require.NoError(t, err)
	require.JSONEq(t, string(input), string(out))
}

func TestOnBodyChunk_ConvertWithoutSingleTypeForMultiInput(t *testing.T) {
	input := []byte(`{"created":1776838589,"data":{"embedding":[0.1,0.2],"object":"embedding"},"id":"req-2","model":"doubao-embedding-vision-251215","object":"list","usage":{"prompt_tokens":2,"total_tokens":2}}`)

	f := &VolcengineMultimodalEmbeddingResponseConverter{}
	ctx := ctxhelper.InitCtxMapIfNeed(context.Background())
	reqOut := httptest.NewRequest("POST", "http://ark.example", io.NopCloser(bytes.NewReader([]byte(`{"input":[{"type":"text","text":"hi"},{"type":"image_url","image_url":{"url":"https://example.com/a.png"}}]}`))))
	ctxhelper.PutReverseProxyRequestOutSnapshot(ctx, reqOut)
	resp := &http.Response{Request: httptest.NewRequest("POST", "http://example", nil).WithContext(ctx)}
	out, err := f.OnBodyChunk(resp, input, 0)
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, jsonUnmarshal(out, &got))
	data, ok := got["data"].([]any)
	require.True(t, ok)
	item, ok := data[0].(map[string]any)
	require.True(t, ok)
	_, hasType := item["type"]
	require.False(t, hasType)
}

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
