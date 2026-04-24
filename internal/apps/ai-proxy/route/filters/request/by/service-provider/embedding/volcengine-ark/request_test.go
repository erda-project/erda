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
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/stretchr/testify/require"

	httperror "github.com/erda-project/erda/internal/apps/ai-proxy/route/http_error"
)

func TestOnProxyRequest_DefaultsAndOutputMapping(t *testing.T) {
	in := map[string]any{
		"model": "doubao-embedding-vision-251215",
		"input": []map[string]any{{
			"type": "text",
			"text": "hello multimodal",
		}},
		"output": map[string]any{
			"primary":    "dense",
			"additional": []string{"multi", "sparse"},
		},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	require.NoError(t, f.OnProxyRequest(pr))
	require.Equal(t, arkMultimodalEmbeddingPath, pr.Out.URL.Path)

	var got map[string]any
	require.NoError(t, json.NewDecoder(pr.Out.Body).Decode(&got))
	require.Equal(t, "doubao-embedding-vision-251215", got["model"])
	require.Equal(t, "Target_modality: text.\nInstruction:Compress the text into one word.\nQuery:", got["instructions"])
	require.EqualValues(t, defaultDimensions, got["dimensions"])
	require.Equal(t, defaultEncodingFormat, got["encoding_format"])

	input, ok := got["input"].([]any)
	require.True(t, ok)
	require.Len(t, input, 1)
	inputItem, ok := input[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "text", inputItem["type"])
	require.Equal(t, "hello multimodal", inputItem["text"])

	multiEmbedding, ok := got["multi_embedding"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "enabled", multiEmbedding["type"])

	sparseEmbedding, ok := got["sparse_embedding"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "enabled", sparseEmbedding["type"])
}

func TestOnProxyRequest_DefaultOutputIsDenseOnly(t *testing.T) {
	in := map[string]any{
		"model": "doubao-embedding-vision-251215",
		"input": []map[string]any{{
			"type": "text",
			"text": "hello multimodal",
		}},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	require.NoError(t, f.OnProxyRequest(pr))

	var got map[string]any
	require.NoError(t, json.NewDecoder(pr.Out.Body).Decode(&got))
	_, hasMulti := got["multi_embedding"]
	_, hasSparse := got["sparse_embedding"]
	require.False(t, hasMulti)
	require.False(t, hasSparse)
}

func TestOnProxyRequest_DefaultInstructionsByModality(t *testing.T) {
	in := map[string]any{
		"model": "doubao-embedding-vision-251215",
		"input": []map[string]any{
			{
				"type":      "image",
				"image_url": "https://example.com/a.png",
			},
			{
				"type":      "video",
				"video_url": "https://example.com/v.mp4",
			},
		},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	require.NoError(t, f.OnProxyRequest(pr))

	var got map[string]any
	require.NoError(t, json.NewDecoder(pr.Out.Body).Decode(&got))
	require.Equal(t, "Target_modality: image and video.\nInstruction:Compress the image/video into one word.\nQuery:", got["instructions"])

	input, ok := got["input"].([]any)
	require.True(t, ok)
	require.Len(t, input, 2)
	imageItem, ok := input[0].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "image_url", imageItem["type"])
	imageURL, ok := imageItem["image_url"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "https://example.com/a.png", imageURL["url"])

	videoItem, ok := input[1].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "video_url", videoItem["type"])
	videoURL, ok := videoItem["video_url"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "https://example.com/v.mp4", videoURL["url"])
}

func TestOnProxyRequest_DimensionsPassThrough(t *testing.T) {
	in := map[string]any{
		"model":       "doubao-embedding-vision-251215",
		"dimensions":  1024,
		"instruction": "compress",
		"options": map[string]any{
			"encoding_format": "float",
		},
		"input": []map[string]any{{
			"type": "text",
			"text": "test",
		}},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	require.NoError(t, f.OnProxyRequest(pr))

	var got map[string]any
	require.NoError(t, json.NewDecoder(pr.Out.Body).Decode(&got))
	require.EqualValues(t, 1024, got["dimensions"])
	require.Equal(t, "compress", got["instructions"])
	require.Equal(t, "float", got["encoding_format"])
}

func TestOnProxyRequest_InvalidImageInput(t *testing.T) {
	in := map[string]any{
		"model": "doubao-embedding-vision-251215",
		"input": []map[string]any{{
			"type": "image",
		}},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	err := f.OnProxyRequest(pr)
	require.Error(t, err)
	requireValidationError(t, err, "input[0].image_url", "input[0]: image_url is required when type=image")
}

func TestOnProxyRequest_InvalidDimensions(t *testing.T) {
	in := map[string]any{
		"model":      "doubao-embedding-vision-251215",
		"dimensions": 1536,
		"input": []map[string]any{{
			"type": "text",
			"text": "test",
		}},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	err := f.OnProxyRequest(pr)
	require.Error(t, err)
	requireValidationError(t, err, "dimensions", "dimensions must be one of [1024, 2048]")
}

func TestOnProxyRequest_FusionUnsupported(t *testing.T) {
	in := map[string]any{
		"model": "doubao-embedding-vision-251215",
		"input": []map[string]any{{
			"type": "text",
			"text": "test",
		}},
		"output": map[string]any{
			"primary": "fusion",
		},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	err := f.OnProxyRequest(pr)
	require.Error(t, err)
	requireValidationError(t, err, "output.primary", "output.primary=fusion is not supported by volcengine-ark multimodal embedding")
}

func TestOnProxyRequest_SparseWithImageRejected(t *testing.T) {
	in := map[string]any{
		"model": "doubao-embedding-vision-251215",
		"input": []map[string]any{{
			"type":      "image",
			"image_url": "https://example.com/a.png",
		}},
		"output": map[string]any{
			"primary":    "dense",
			"additional": []string{"sparse"},
		},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	err := f.OnProxyRequest(pr)
	require.Error(t, err)
	requireValidationError(t, err, "output.additional", "output.additional=sparse is only supported for text input")
}

func TestOnProxyRequest_SparseEnabledByOptionsWithImageRejected(t *testing.T) {
	in := map[string]any{
		"model": "doubao-embedding-vision-251215",
		"input": []map[string]any{{
			"type":      "image",
			"image_url": "https://example.com/a.png",
		}},
		"options": map[string]any{
			"sparse_embedding": map[string]any{
				"type": "enabled",
			},
		},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	err := f.OnProxyRequest(pr)
	require.Error(t, err)
	requireValidationError(t, err, "options.sparse_embedding", "options.sparse_embedding=enabled is only supported for text input")
}

func TestOnProxyRequest_SparseDisabledByOptions(t *testing.T) {
	in := map[string]any{
		"model": "doubao-embedding-vision-251215",
		"input": []map[string]any{{
			"type": "text",
			"text": "test",
		}},
		"options": map[string]any{
			"sparse_embedding": map[string]any{
				"type": "disabled",
			},
		},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	require.NoError(t, f.OnProxyRequest(pr))

	var got map[string]any
	require.NoError(t, json.NewDecoder(pr.Out.Body).Decode(&got))
	sparseEmbedding, ok := got["sparse_embedding"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "disabled", sparseEmbedding["type"])
}

func TestOnProxyRequest_SparseConflictWithOutputAdditional(t *testing.T) {
	in := map[string]any{
		"model": "doubao-embedding-vision-251215",
		"input": []map[string]any{{
			"type": "text",
			"text": "test",
		}},
		"output": map[string]any{
			"primary":    "dense",
			"additional": []string{"sparse"},
		},
		"options": map[string]any{
			"sparse_embedding": map[string]any{
				"type": "disabled",
			},
		},
	}
	pr := buildProxyRequest(t, in)

	f := &VolcengineMultimodalEmbeddingConverter{}
	err := f.OnProxyRequest(pr)
	require.Error(t, err)
	requireValidationError(t, err, "output.additional", "output.additional contains sparse but options.sparse_embedding.type=disabled")
}

func requireValidationError(t *testing.T, err error, param string, message string) {
	t.Helper()
	var httpErr *httperror.HTTPError
	require.True(t, errors.As(err, &httpErr), "expected HTTPError, got %T", err)
	require.Equal(t, 400, httpErr.StatusCode)
	require.Equal(t, message, httpErr.Message)
	require.Equal(t, "invalid_request_error", httpErr.ErrorCtx["code"])
	require.Equal(t, message, httpErr.ErrorCtx["message"])
	require.Equal(t, param, httpErr.ErrorCtx["param"])
	require.Equal(t, "validation_error", httpErr.ErrorCtx["type"])
}

func buildProxyRequest(t *testing.T, payload map[string]any) *httputil.ProxyRequest {
	t.Helper()
	b, err := json.Marshal(payload)
	require.NoError(t, err)

	inReq := httptest.NewRequest("POST", "http://ai-proxy.local/v1/multimodal/embeddings", bytes.NewReader(b))
	outReq := httptest.NewRequest("POST", "https://ark.cn-beijing.volces.com/v1/multimodal/embeddings", bytes.NewReader(b))

	return &httputil.ProxyRequest{In: inReq, Out: outReq}
}
