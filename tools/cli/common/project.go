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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	pb "github.com/erda-project/erda-proto-go/msp/tenant/project/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/utils"
)

type UserOrgProj struct {
	UserId    string
	OrgId     string
	ProjectId string
}

func GetProjectDetail(ctx *command.Context, orgID, projectID uint64) (apistructs.ProjectDTO, error) {
	var resp apistructs.ProjectDetailResponse
	var b bytes.Buffer

	response, err := ctx.Get().
		Header("Org-ID", strconv.FormatUint(orgID, 10)).
		Path(fmt.Sprintf("/api/projects/%d", projectID)).
		Do().Body(&b)
	if err != nil {
		return apistructs.ProjectDTO{}, fmt.Errorf(utils.FormatErrMsg(
			"get project detail", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.ProjectDTO{}, fmt.Errorf(utils.FormatErrMsg("get project detail",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.ProjectDTO{}, fmt.Errorf(utils.FormatErrMsg("get project detail",
			fmt.Sprintf("failed to unmarshal project detail response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.ProjectDTO{}, fmt.Errorf(utils.FormatErrMsg("get project detail",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp.Data, nil
}

func CreateProject(ctx *command.Context, orgID uint64, name, desc string,
	resourceConfigs *apistructs.ResourceConfigs) (uint64, error) {
	var request apistructs.ProjectCreateRequest
	var response apistructs.ProjectCreateResponse
	var b bytes.Buffer

	request.Name = name
	request.Desc = desc
	request.OrgID = orgID
	request.Template = "DevOps"
	if resourceConfigs != nil {
		request.ResourceConfigs = resourceConfigs
	}

	resp, err := ctx.Post().Path("/api/projects").
		Header("Org-ID", strconv.FormatUint(orgID, 10)).
		JSONBody(request).Do().Body(&b)
	if err != nil {
		return response.Data, fmt.Errorf(
			utils.FormatErrMsg("create", "failed to request ("+err.Error()+")", false))
	}

	if !resp.IsOK() {
		return response.Data, fmt.Errorf(utils.FormatErrMsg("create",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				resp.StatusCode(), resp.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &response); err != nil {
		return response.Data, fmt.Errorf(utils.FormatErrMsg("create",
			fmt.Sprintf("failed to unmarshal project create response ("+err.Error()+")"), false))
	}

	if !response.Success {
		return response.Data, fmt.Errorf(utils.FormatErrMsg("create",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				response.Error.Code, response.Error.Msg), false))
	}

	return response.Data, nil
}

func CreateMSPProject(ctx *command.Context, projectID uint64, name string) (*pb.Project, error) {
	var request pb.CreateProjectRequest
	response := struct {
		apistructs.Header
		Data *pb.Project `json:"data"`
	}{}
	var b bytes.Buffer

	request.Id = strconv.FormatUint(projectID, 10)
	request.Name = name
	request.DisplayName = name
	request.Type = "DOP"

	resp, err := ctx.Post().Path("/api/msp/tenant/project").
		JSONBody(request).Do().Body(&b)

	if err != nil {
		return response.Data, fmt.Errorf(
			utils.FormatErrMsg("create", "failed to request ("+err.Error()+")", false))
	}

	if !resp.IsOK() {
		return response.Data, fmt.Errorf(utils.FormatErrMsg("create",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				resp.StatusCode(), resp.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &response); err != nil {
		return response.Data, fmt.Errorf(utils.FormatErrMsg("create",
			fmt.Sprintf("failed to unmarshal project create response ("+err.Error()+")"), false))
	}

	return response.Data, nil
}

func ImportPackage(ctx *command.Context, orgID, projectID uint64, pkg string) (uint64, error) {
	response := struct {
		apistructs.Header
		Data uint64
	}{}
	var b bytes.Buffer

	f, err := os.Open(pkg)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	fileNameWithExt := filepath.Base(pkg)

	resp, err := ctx.Post().
		Path(fmt.Sprintf("/api/orgs/%d/projects/%d/package/actions/import", orgID, projectID)).
		MultipartFormDataBody(map[string]httpclient.MultipartItem{
			"file": {
				Reader:   f,
				Filename: fileNameWithExt,
			},
		}).Do().Body(&b)
	if err != nil {
		return response.Data, fmt.Errorf(
			utils.FormatErrMsg("create", "failed to request ("+err.Error()+")", false))
	}

	if !resp.IsOK() {
		return response.Data, fmt.Errorf(utils.FormatErrMsg("import",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				resp.StatusCode(), resp.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &response); err != nil {
		return response.Data, fmt.Errorf(utils.FormatErrMsg("import",
			fmt.Sprintf("failed to unmarshal project import response ("+err.Error()+")"), false))
	}

	return response.Data, nil
}

// GetProjectList 获取 org 下的所有 Projects
func GetProjectList(ctx *command.Context, orgId string) ([]apistructs.ProjectDTO, error) {
	var resp apistructs.ProjectListResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/projects").Param("joined", "true").
		Param("orgId", orgId).Do().Body(&b)
	if err != nil {
		return nil, fmt.Errorf(
			utils.FormatErrMsg("list", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return nil, fmt.Errorf(utils.FormatErrMsg("list",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf(utils.FormatErrMsg("list",
			fmt.Sprintf("failed to unmarshal projects list response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return nil, fmt.Errorf(utils.FormatErrMsg("list",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	if resp.Data.Total < 0 {
		return nil, fmt.Errorf(
			utils.FormatErrMsg("list", "critical: the number of projects is less than 0", false))
	}

	if resp.Data.Total == 0 {
		fmt.Printf(utils.FormatErrMsg("list", "no projects created\n", false))
		return nil, nil
	}

	return resp.Data.List, nil
}

// GetUserOrgProjID get UserId,ProjectId,OrgID info
func GetUserOrgProjID(ctx *command.Context, orgName, projectName string) (UserOrgProj, error) {
	var uop UserOrgProj
	var userId string
	var orgId, projectId uint64

	_, orgId, err := GetOrgID(ctx, orgName)
	if err != nil {
		return uop, err
	}

	if sessionInfo, ok := ctx.Sessions[ctx.CurrentOpenApiHost]; ok {
		userId = sessionInfo.ID
	}

	if userId == "" || orgId <= 0 {
		return uop, errors.New("get invalid orgID [" + strconv.FormatUint(orgId, 10) + "] or userID [" + userId + "]")
	}

	projs, err := GetProjectList(ctx, strconv.FormatUint(orgId, 10))
	if err != nil {
		return uop, err
	}
	for _, proj := range projs {
		if proj.Name == projectName {
			projectId = proj.ID
		}
	}

	if projectId <= 0 {
		return uop, errors.New("get invalid projectID [" + strconv.FormatUint(projectId, 10) + "]")
	}
	uop.ProjectId = strconv.FormatUint(projectId, 10)
	uop.UserId = userId
	uop.OrgId = strconv.FormatUint(orgId, 10)

	return uop, nil
}
