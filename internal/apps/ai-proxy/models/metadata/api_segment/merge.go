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

package api_segment

import (
	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
)

// MergeAPIStyleConfig merges multiple API segments by priority (later segments override earlier ones).
// It first uses api.GetAPIStyleConfigByMethodAndPathMatcher to find the corresponding APIStyleConfig,
// then merges them field by field, where non-empty fields from later configs override earlier ones.
func MergeAPIStyleConfig(method string, pathMatcher string, apiSegments ...*API) *api_style.APIStyleConfig {
	if len(apiSegments) == 0 {
		return nil
	}

	var result *api_style.APIStyleConfig

	// process each API segment in order (later ones have higher priority)
	for _, apiSeg := range apiSegments {
		if apiSeg == nil {
			continue
		}

		config := apiSeg.GetAPIStyleConfigByMethodAndPathMatcher(method, pathMatcher)
		if config == nil {
			continue
		}

		if result == nil {
			result = config
			continue
		}

		mergeAPIStyleConfigFields(result, config)
	}

	return result
}

// mergeAPIStyleConfigFields merges non-empty fields from high priority config into low priority config
func mergeAPIStyleConfigFields(low, high *api_style.APIStyleConfig) {
	if high == nil {
		return
	}

	if high.Method != "" {
		low.Method = high.Method
	}
	if high.Scheme != "" {
		low.Scheme = high.Scheme
	}
	if high.Host != "" {
		low.Host = high.Host
	}
	if high.Path != nil {
		low.Path = high.Path
	}
	if len(high.PathOp) > 0 {
		// replace the entire PathOp
		low.PathOp = high.PathOp
	}

	// merge map fields
	if len(high.QueryParams) > 0 {
		if low.QueryParams == nil {
			low.QueryParams = make(map[string][]string)
		}
		for key, value := range high.QueryParams {
			if len(value) > 0 {
				// replace the entire value for this key
				low.QueryParams[key] = value
			}
		}
	}

	if len(high.Headers) > 0 {
		if low.Headers == nil {
			low.Headers = make(map[string][]string)
		}
		for key, value := range high.Headers {
			if len(value) > 0 {
				// replace the entire value for this key
				low.Headers[key] = value
			}
		}
	}

	// merge body field
	if high.Body != nil {
		// replace the entire Body field
		low.Body = high.Body
	}
}
