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

	"github.com/erda-project/erda/tools/cli/httputils"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

//func GetAddonDetail(ctx *command.Context, orgId, projectId, applicationId int) (
//	apistructs.ApplicationFetchResponse, error) {
//	var (
//		resp apistructs.ApplicationFetchResponse
//		b    bytes.Buffer
//	)
//
//	response, err := ctx.Get().Header("Org-ID", strconv.Itoa(orgId)).
//		Path(fmt.Sprintf("/api/applications/%d?projectId=%d", applicationId, projectId)).
//		Do().Body(&b)
//	if err != nil {
//		return apistructs.ApplicationFetchResponse{}, fmt.Errorf(format.FormatErrMsg(
//			"get application detail", "failed to request ("+err.Error()+")", false))
//	}
//
//	if !response.IsOK() {
//		return apistructs.ApplicationFetchResponse{}, fmt.Errorf(format.FormatErrMsg("get application detail",
//			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
//				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
//	}
//
//	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
//		return apistructs.ApplicationFetchResponse{}, fmt.Errorf(format.FormatErrMsg("get application detail",
//			fmt.Sprintf("failed to unmarshal application detail response ("+err.Error()+")"), false))
//	}
//
//	if !resp.Success {
//		return apistructs.ApplicationFetchResponse{}, fmt.Errorf(format.FormatErrMsg("get application detail",
//			fmt.Sprintf("failed to request, error code: %s, error message: %s",
//				resp.Error.Code, resp.Error.Msg), false))
//	}
//
//	return resp, nil
//}

// TODO paging
func GetAddonList(ctx *command.Context, orgId, projectId uint64) (apistructs.AddonListResponse, error) {
	var resp apistructs.AddonListResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/addons").
		Header("Org-ID", strconv.FormatUint(orgId, 10)).
		Param("type", "project").
		Param("value", strconv.FormatUint(projectId, 10)).
		Do().Body(&b)
	if err != nil {
		return apistructs.AddonListResponse{}, fmt.Errorf(
			format.FormatErrMsg("list", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.AddonListResponse{}, fmt.Errorf(format.FormatErrMsg("list",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.AddonListResponse{}, fmt.Errorf(format.FormatErrMsg("list",
			fmt.Sprintf("failed to unmarshal application list response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.AddonListResponse{}, fmt.Errorf(
			format.FormatErrMsg("list",
				fmt.Sprintf("failed to request, error code: %s, error message: %s",
					resp.Error.Code, resp.Error.Msg), false))
	}

	return resp, nil
}

func DeleteAddon(ctx *command.Context, orgID uint64, addonID string) error {
	r := ctx.Delete().
		Header("Org-ID", strconv.FormatUint(orgID, 10)).
		Path(fmt.Sprintf("/api/addons/%s", addonID))
	resp, err := httputils.DoResp(r)
	if err != nil {
		return fmt.Errorf(
			format.FormatErrMsg("remove", "failed to remove addon, error: "+err.Error(), false))
	}
	if err := resp.ParseData(nil); err != nil {
		return fmt.Errorf(
			format.FormatErrMsg(
				"remove", "failed to parse remove addon response, error: "+err.Error(), false))
	}

	return nil
}
