// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
