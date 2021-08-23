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

package common

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

func GetBuildDetail(ctx *command.Context, buildID string) (*apistructs.PipelineDetailDTO, error) {
	if buildID == "" {
		return nil, fmt.Errorf(
			format.FormatErrMsg("pipeline info", "pipelineID is empty", false))
	}

	var pipelineInfoResp apistructs.PipelineDetailResponse
	response, err := ctx.Get().Path("/api/pipelines/" + buildID).Do().JSON(&pipelineInfoResp)
	if err != nil {
		return nil, fmt.Errorf(format.FormatErrMsg(
			"pipeline info", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return nil, fmt.Errorf(format.FormatErrMsg("pipeline info",
			fmt.Sprintf("failed to request, status code: %d", response.StatusCode()), false))
	}

	if !pipelineInfoResp.Success {
		return nil, fmt.Errorf(format.FormatErrMsg("pipeline info",
			fmt.Sprintf("failed to request: %+v", pipelineInfoResp.Error), false))
	}

	return pipelineInfoResp.Data, nil
}
