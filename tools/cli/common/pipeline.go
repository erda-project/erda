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
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

func GetPipeline(ctx *command.Context, orgId, pipelineId uint64) (apistructs.PipelineDetailDTO, error) {
	// fetch pipeline info
	var pipelineInfoResp apistructs.PipelineDetailResponse
	response, err := ctx.Get().
		Header("Org-ID", strconv.FormatUint(orgId, 10)).
		Path(fmt.Sprintf("/api/pipelines/%d", pipelineId)).
		Do().JSON(&pipelineInfoResp)
	if err != nil {
		return apistructs.PipelineDetailDTO{}, err
	}
	if !response.IsOK() {
		return apistructs.PipelineDetailDTO{}, errors.Errorf("status fail, status code: %d, err: %+v", response.StatusCode(), pipelineInfoResp.Error)
	}
	if !pipelineInfoResp.Success {
		return apistructs.PipelineDetailDTO{}, errors.Errorf("status fail: %+v", pipelineInfoResp.Error)
	}

	return *pipelineInfoResp.Data, nil
}
