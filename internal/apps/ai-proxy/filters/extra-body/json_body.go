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

package extra_body

import (
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata"
)

const (
	MetadataPublicKeyExtraJSONBodyForStream    = "extra_json_body_for_stream"
	MetadataPublicKeyExtraJSONBodyForNonStream = "extra_json_body_for_non_stream"
)

// GetExtraJSONBody get extra json body from metadata.
// For stream, field name is: "extra_json_body_for_stream"
// For non-stream, field name is: "extra_json_body_for_non_stream"
func GetExtraJSONBody(m *metadata.Metadata, isStream bool) (map[string]any, error) {
	key := MetadataPublicKeyExtraJSONBodyForNonStream
	if isStream {
		key = MetadataPublicKeyExtraJSONBodyForStream
	}
	return unmarshalJSONFromString(m.Public[key])
}

func FulfillExtraJSONBody(m *metadata.Metadata, isStream bool, body map[string]any) error {
	if body == nil {
		return nil
	}
	extra_json_body, err := GetExtraJSONBody(m, isStream)
	if err != nil {
		return fmt.Errorf("failed to get extra json body, err: %v", err)
	}
	// iterate key in extra_json_body, only set key if key not exist in body
	for k, v := range extra_json_body {
		if _, ok := body[k]; !ok {
			body[k] = v
		}
	}
	return nil
}

// unmarshalJSONFromString unmarshal json string to map[string]any
func unmarshalJSONFromString(s string) (map[string]any, error) {
	if s == "" {
		return nil, nil
	}
	var jsonObj map[string]any
	if err := json.Unmarshal([]byte(s), &jsonObj); err != nil {
		return nil, err
	}
	return jsonObj, nil
}
