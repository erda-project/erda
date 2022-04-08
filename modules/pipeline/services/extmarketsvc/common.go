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

package extmarketsvc

import (
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

// MakeActionTypeVersion return ext item.
// Example: git, git@1.0, git@1.1
func MakeActionTypeVersion(action *pipelineyml.Action) string {
	r := action.Type.String()
	if action.Version != "" {
		r = r + "@" + action.Version
	}
	return r
}

func MakeActionLocationsBySource(source apistructs.PipelineSource) []string {
	var locations []string
	switch source {
	case apistructs.PipelineSourceCDPDev, apistructs.PipelineSourceCDPTest, apistructs.PipelineSourceCDPStaging, apistructs.PipelineSourceCDPProd, apistructs.PipelineSourceBigData:
		locations = append(locations, apistructs.PipelineTypeFDP.String()+"/")
	case apistructs.PipelineSourceDice, apistructs.PipelineSourceProject, apistructs.PipelineSourceProjectLocal, apistructs.PipelineSourceOps, apistructs.PipelineSourceQA:
		locations = append(locations, apistructs.PipelineTypeCICD.String()+"/")
	default:
		locations = append(locations, apistructs.PipelineTypeDefault.String()+"/")
	}
	return locations
}

func getActionTypeVersion(nameVersion string) (string, string) {
	splits := strings.SplitN(nameVersion, "@", 2)
	name := splits[0]
	version := ""
	if len(splits) > 1 {
		version = splits[1]
	}
	return name, version
}
