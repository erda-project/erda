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

func GetReleaseDetail(ctx *command.Context, orgId uint64, releaseId string) (apistructs.ReleaseGetResponseData, error) {
	var resp apistructs.ReleaseGetResponse
	var b bytes.Buffer

	response, err := ctx.Get().
		Header("Org-ID", strconv.FormatUint(orgId, 10)).
		Path(fmt.Sprintf("/api/releases/%s", releaseId)).
		Do().Body(&b)
	if err != nil {
		return apistructs.ReleaseGetResponseData{}, fmt.Errorf(format.FormatErrMsg(
			"get project detail", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.ReleaseGetResponseData{}, fmt.Errorf(format.FormatErrMsg("get project detail",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.ReleaseGetResponseData{}, fmt.Errorf(format.FormatErrMsg("get project detail",
			fmt.Sprintf("failed to unmarshal project detail response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.ReleaseGetResponseData{}, fmt.Errorf(format.FormatErrMsg("get project detail",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp.Data, nil
}

func GetPagingReleases(ctx *command.Context, orgId, applicationId uint64, branch string, isVersion bool, pageNo, pageSize int) (apistructs.ReleaseListResponseData, error) {
	var resp apistructs.ReleaseListResponse
	var b bytes.Buffer
	req := ctx.Get().Path("/api/releases").
		Header("Org-ID", strconv.FormatUint(orgId, 10)).
		Param("pageNo", strconv.Itoa(pageNo)).
		Param("pageSize", strconv.Itoa(pageSize)).
		Param("applicationId", strconv.FormatUint(applicationId, 10)).
		Param("branchName", branch).
		Param("isVersion", strconv.FormatBool(isVersion))
	response, err := req.Do().Body(&b)
	if err != nil {
		return apistructs.ReleaseListResponseData{}, fmt.Errorf(format.FormatErrMsg(
			"get release detail", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.ReleaseListResponseData{}, fmt.Errorf(format.FormatErrMsg("get release detail",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.ReleaseListResponseData{}, fmt.Errorf(format.FormatErrMsg("get release detail",
			fmt.Sprintf("failed to unmarshal runtime detail response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.ReleaseListResponseData{}, fmt.Errorf(format.FormatErrMsg("get release detail",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp.Data, nil
}

func ChooseRelease(ctx *command.Context, orgId, applicationId uint64, branch, version string) (bool, apistructs.ReleaseGetResponseData, error) {
	var release apistructs.ReleaseGetResponseData
	num := 0
	found := false
	err := dicedir.PagingAll(func(pageNo, pageSize int) (bool, error) {
		paging, err := GetPagingReleases(ctx, orgId, applicationId, branch, true, pageNo, pageSize)
		if err != nil {
			return false, err
		}
		for _, r := range paging.Releases {
			if r.Version == version {
				release = r
				found = true
				return false, nil
			}
		}
		num += len(paging.Releases)
		return paging.Total > int64(num), nil
	}, 10)
	if err != nil {
		return false, apistructs.ReleaseGetResponseData{}, err
	}

	if !found {
		return false, apistructs.ReleaseGetResponseData{}, nil
	}

	return true, release, nil
}
