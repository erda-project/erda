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
	"time"

	"github.com/erda-project/erda/pkg/http/httpclient"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/dicedir"
	"github.com/erda-project/erda/tools/cli/format"
	"github.com/erda-project/erda/tools/cli/httputils"
)

func GetAddonResp(ctx *command.Context, orgId uint64, addonId string) (
	*httpclient.Response, bytes.Buffer, error) {
	var b bytes.Buffer

	response, err := ctx.Get().Header("Org-ID", strconv.FormatUint(orgId, 10)).
		Path(fmt.Sprintf("/api/addons/%s", addonId)).
		Do().Body(&b)
	if err != nil {
		return nil, b, fmt.Errorf(format.FormatErrMsg(
			"get addon detail", "failed to request ("+err.Error()+")", false))
	}

	return response, b, nil
}

func GetAddonDetail(ctx *command.Context, orgId uint64, addonId string) (
	apistructs.AddonFetchResponseData, error) {
	var resp apistructs.AddonFetchResponse
	var b bytes.Buffer

	response, b, err := GetAddonResp(ctx, orgId, addonId)
	if err != nil {
		return apistructs.AddonFetchResponseData{}, err
	}

	if !response.IsOK() {
		return apistructs.AddonFetchResponseData{}, fmt.Errorf(format.FormatErrMsg("get addon detail",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.AddonFetchResponseData{}, fmt.Errorf(format.FormatErrMsg("get addon detail",
			fmt.Sprintf("failed to unmarshal addon detail response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.AddonFetchResponseData{}, fmt.Errorf(format.FormatErrMsg("get addon detail",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp.Data, nil
}

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

func DeleteAddon(ctx *command.Context, orgId uint64, addonId string) error {
	r := ctx.Delete().
		Header("Org-ID", strconv.FormatUint(orgId, 10)).
		Path(fmt.Sprintf("/api/addons/%s", addonId))
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

func CreateErdaAddon(ctx *command.Context, orgId, projectId uint64, clusterName, workspace, name, plan string, timeout int) error {
	var request apistructs.AddonDirectCreateRequest
	request.OrgID = orgId
	request.ProjectID = projectId
	request.ClusterName = clusterName
	request.Workspace = workspace
	request.Addons = diceyml.AddOns{name: &diceyml.AddOn{Plan: plan}}
	request.ShareScope = "PROJECT"

	var response apistructs.AddonCreateResponse
	var b bytes.Buffer

	resp, err := ctx.Post().Path("/api/addons/actions/create-addon").
		Header("Org-ID", strconv.FormatUint(orgId, 10)).
		JSONBody(request).
		Do().Body(&b)
	if err != nil {
		return fmt.Errorf(
			format.FormatErrMsg("create", "failed to request ("+err.Error()+")", false))
	}

	if !resp.IsOK() {
		return fmt.Errorf(format.FormatErrMsg("create",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				resp.StatusCode(), resp.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &response); err != nil {
		return fmt.Errorf(format.FormatErrMsg("create",
			fmt.Sprintf("failed to unmarshal application create response ("+err.Error()+")"), false))
	}

	if !response.Success {
		return fmt.Errorf(format.FormatErrMsg("create",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				response.Error.Code, response.Error.Msg), false))
	}

	aId := response.Data
	err = dicedir.DoTaskWithTimeout(func() (bool, error) {
		s, err := GetAddonDetail(ctx, orgId, aId)
		if err == nil && s.Status == "ATTACHED" {
			return true, nil
		}
		return false, nil
	}, time.Duration(timeout)*time.Minute)
	if err != nil {
		return err
	}

	return nil
}

func CreateCustomAddon(ctx *command.Context, orgId, projectId uint64, workspace,
	addonName, name string, configs map[string]interface{}) error {
	var request apistructs.CustomAddonCreateRequest
	request.AddonName = addonName
	request.Name = name
	request.ProjectID = projectId
	request.Workspace = workspace
	request.Configs = configs

	var response apistructs.AddonCreateResponse
	var b bytes.Buffer
	resp, err := ctx.Post().Path("/api/addons/actions/create-custom").
		Header("Org-ID", strconv.FormatUint(orgId, 10)).
		JSONBody(request).
		Do().Body(&b)
	if err != nil {
		return fmt.Errorf(
			format.FormatErrMsg("create", "failed to request ("+err.Error()+")", false))
	}

	if !resp.IsOK() {
		return fmt.Errorf(format.FormatErrMsg("create",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				resp.StatusCode(), resp.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &response); err != nil {
		return fmt.Errorf(format.FormatErrMsg("create",
			fmt.Sprintf("failed to unmarshal application create response ("+err.Error()+")"), false))
	}

	if !response.Success {
		return fmt.Errorf(format.FormatErrMsg("create",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				response.Error.Code, response.Error.Msg), false))
	}

	return nil
}
