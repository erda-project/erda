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
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/dicedir"
	"github.com/erda-project/erda/tools/cli/format"
)

func GetApplicationDetail(ctx *command.Context, orgId, projectId, applicationId int) (
	apistructs.ApplicationFetchResponse, error) {
	var (
		resp apistructs.ApplicationFetchResponse
		b    bytes.Buffer
	)

	response, err := ctx.Get().Header("Org-ID", strconv.Itoa(orgId)).
		Path(fmt.Sprintf("/api/applications/%d?projectId=%d", applicationId, projectId)).
		Do().Body(&b)
	if err != nil {
		return apistructs.ApplicationFetchResponse{}, fmt.Errorf(format.FormatErrMsg(
			"get application detail", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.ApplicationFetchResponse{}, fmt.Errorf(format.FormatErrMsg("get application detail",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.ApplicationFetchResponse{}, fmt.Errorf(format.FormatErrMsg("get application detail",
			fmt.Sprintf("failed to unmarshal application detail response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.ApplicationFetchResponse{}, fmt.Errorf(format.FormatErrMsg("get application detail",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp, nil
}

func GetApplications(ctx *command.Context, orgId, projectId uint64) ([]apistructs.ApplicationDTO, error) {
	var apps []apistructs.ApplicationDTO
	err := dicedir.PagingAll(func(pageNo, pageSize int) (bool, error) {
		page, err := GetPagingApplications(ctx, orgId, projectId, pageNo, pageSize)
		if err != nil {
			return false, err
		}
		apps = append(apps, page.List...)
		return page.Total > len(apps), nil
	}, 1)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func GetPagingApplications(ctx *command.Context, orgId, projectId uint64, pageNo, pageSize int) (apistructs.ApplicationListResponseData, error) {
	var resp apistructs.ApplicationListResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/applications").
		Header("Org-ID", strconv.FormatUint(orgId, 10)).
		Param("projectId", strconv.FormatUint(projectId, 10)).
		Param("pageNo", strconv.Itoa(pageNo)).Param("pageSize", strconv.Itoa(pageSize)).
		Do().Body(&b)
	if err != nil {
		return apistructs.ApplicationListResponseData{}, fmt.Errorf(
			format.FormatErrMsg("list", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.ApplicationListResponseData{}, fmt.Errorf(format.FormatErrMsg("list",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.ApplicationListResponseData{}, fmt.Errorf(format.FormatErrMsg("list",
			fmt.Sprintf("failed to unmarshal application list response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.ApplicationListResponseData{}, fmt.Errorf(
			format.FormatErrMsg("list",
				fmt.Sprintf("failed to request, error code: %s, error message: %s",
					resp.Error.Code, resp.Error.Msg), false))
	}

	return resp.Data, nil
}
