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
	"fmt"
	"strings"

	"github.com/erda-project/erda/internal/apps/ai-proxy/models/metadata/api_segment/api_style"
)

type API struct {
	APIStyle  api_style.APIStyle  `json:"apiStyle,omitempty"`
	APIVendor api_style.APIVendor `json:"apiVendor,omitempty"`

	// APIStyleConfig is the default APIStyleConfig.
	APIStyleConfig *api_style.APIStyleConfig `json:"apiStyleConfig,omitempty"`
	// APIStyleConfigs is route-level APIStyleConfig.
	// Each map value is a completed APIStyleConfig, not patch to default APIStyleConfig.
	// Key can be multiple concatenated strings to share configs, like: `*:/v1/files;*:/v1/files/{file_id};*:/v1/files/{file_id}/content`.
	APIStyleConfigs map[string]api_style.APIStyleConfig `json:"apiStyleConfigs,omitempty"`
}

// GenerateAPIStyleConfigKey .
// generated key, method is always upper-case, but pathMatcher is case-sensitive.
func GenerateAPIStyleConfigKey(method string, pathMatcher string) string {
	return fmt.Sprintf("%s:%s", strings.ToUpper(method), pathMatcher)
}

// GetAPIStyleConfigByMethodAndPathMatcher return default APIStyleConfig is not found by route.
// pathMatcher: same as defined in routes.yml, like `/v1/files/{file_id}/content`, not concrete path like `/v1/files/a/content`.
func (api *API) GetAPIStyleConfigByMethodAndPathMatcher(method string, pathMatcher string) *api_style.APIStyleConfig {
	// multi keys
	singleKeyAPIStyleConfigMap := make(map[string]api_style.APIStyleConfig)
	for concatenatedKey := range api.APIStyleConfigs {
		multiKeys := strings.Split(concatenatedKey, ";")
		for _, key := range multiKeys {
			singleKeyAPIStyleConfigMap[key] = api.APIStyleConfigs[concatenatedKey]
		}
	}

	// get in order
	keys := []string{
		GenerateAPIStyleConfigKey(method, pathMatcher), // e.g., POST:/v1/files
		GenerateAPIStyleConfigKey("*", pathMatcher),    // e.g., *:/v1/files
		GenerateAPIStyleConfigKey(method, "*"),         // e.g., POST:*
		GenerateAPIStyleConfigKey("*", "*"),            // *:*
	}
	for _, key := range keys {
		if v, ok := singleKeyAPIStyleConfigMap[key]; ok {
			return &v
		}
	}
	// TODO return nil after all APIStyleConfig are migrated.
	return api.APIStyleConfig
}
