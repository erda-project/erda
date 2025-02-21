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

package openai_v1_models

import (
	"regexp"

	richclientpb "github.com/erda-project/erda-proto-go/apps/aiproxy/client/rich_client/pb"
)

func GenerateModelDisplayName(model *richclientpb.RichModel) string {
	s := model.Model.Name
	attrs := []string{}

	// provider type & location
	providerType := model.Provider.Type
	location := ""

	if model.Provider.Metadata != nil && model.Provider.Metadata.Public != nil {
		if loc := model.Provider.Metadata.Public["location"]; loc != "" {
			location = loc
		}
		if displayProviderType := model.Provider.Metadata.Public["displayProviderType"]; displayProviderType != "" {
			providerType = displayProviderType
		}
	}

	attrs = append(attrs, "T:"+providerType)
	if location != "" {
		attrs = append(attrs, "L:"+location)
	}

	// model id at last
	attrs = append(attrs, "ID:"+model.Model.Id)

	attrs_s := ""
	for _, attr := range attrs {
		attrs_s += "[" + attr + "]"
	}

	return s + " " + attrs_s
}

func ParseModelUUIDFromDisplayName(s string) string {
	regex := regexp.MustCompile(`\[ID:([^]]*)]`)
	matches := regex.FindAllStringSubmatch(s, 1)
	if len(matches) == 0 {
		return ""
	}
	return matches[0][1]
}
