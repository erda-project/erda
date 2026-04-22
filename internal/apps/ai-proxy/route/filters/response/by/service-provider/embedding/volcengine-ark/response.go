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
)

func (f *VolcengineMultimodalEmbeddingResponseConverter) OnBodyChunk(resp *http.Response, chunk []byte, index int64) ([]byte, error) {
	var raw map[string]any
	if err := json.Unmarshal(chunk, &raw); err != nil {
		return chunk, nil
	}
	if !needsConvert(raw) {
		return chunk, nil
	}
	converted := convertToCanonical(raw)
	out, err := json.Marshal(converted)
	if err != nil {
		return chunk, nil
	}
	return out, nil
}

func needsConvert(raw map[string]any) bool {
	data, ok := raw["data"].(map[string]any)
	if !ok {
		return false
	}
	_, hasEmbedding := data["embedding"]
	return hasEmbedding
}

func convertToCanonical(raw map[string]any) map[string]any {
	data := raw["data"].(map[string]any)
	item := map[string]any{
		"object":    "embedding",
		"index":     0,
		"embedding": data["embedding"],
	}
	if v, ok := data["multi_embedding"]; ok {
		item["multi_embedding"] = v
	}
	if v, ok := data["sparse_embedding"]; ok {
		item["sparse_embedding"] = v
	}
	raw["data"] = []any{item}
	raw["object"] = "list"
	return raw
}
