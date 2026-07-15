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
	"net/url"
	"strconv"

	"github.com/erda-project/erda-proto-go/dop/issue/stream/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/utils"
)

func GetIssue(ctx *command.Context, orgID, projectID, issueID uint64) (*apistructs.Issue, error) {
	var resp apistructs.IssueGetResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path(fmt.Sprintf("/api/issues/%d", issueID)).
		Param("orgID", strconv.FormatUint(orgID, 10)).
		Param("projectID", strconv.FormatUint(projectID, 10)).
		Do().Body(&b)
	if err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg(
			"get issue", "failed to request ("+err.Error()+")", false))

	}

	if !response.IsOK() {
		return nil, formatHTTPFailureFromResponse("get issue", response, b.Bytes())
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("get issue",
			"failed to unmarshal response ("+err.Error()+")", false))

	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("get issue",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))

	}

	return resp.Data, nil
}

func CreateIssue(ctx *command.Context, orgID uint64, request *apistructs.IssueCreateRequest) (uint64, error) {
	var resp apistructs.IssueCreateResponse
	var b bytes.Buffer

	response, err := ctx.Post().Path("/api/issues").JSONBody(request).
		Header("org", strconv.FormatUint(orgID, 10)).
		Do().Body(&b)
	if err != nil {
		return 0, fmt.Errorf("%s", utils.FormatErrMsg(
			"create issue", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return 0, formatHTTPFailureFromResponse("create issue", response, b.Bytes())
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return 0, fmt.Errorf("%s", utils.FormatErrMsg("create issue",
			"failed to unmarshal response ("+err.Error()+")", false))
	}

	if !resp.Success {
		return 0, fmt.Errorf("%s", utils.FormatErrMsg("create issue",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp.Data, nil
}

func ListMyIssueResponse(ctx *command.Context, req *apistructs.IssuePagingRequest) (*apistructs.IssuePagingResponse, error) {
	var resp apistructs.IssuePagingResponse
	var b bytes.Buffer

	values := url.Values{}
	for _, s := range req.State {
		values.Add("state", strconv.FormatInt(s, 10))
	}
	for _, a := range req.Assignees {
		values.Add("assignee", a)
	}
	for _, t := range req.Type {
		values.Add("type", string(t))
	}

	response, err := ctx.Get().Path("/api/issues").
		Header("org", strconv.FormatInt(req.OrgID, 10)).
		Header("Org-ID", strconv.FormatInt(req.OrgID, 10)).
		Param("pageNo", strconv.FormatUint(req.PageNo, 10)).
		Param("pageSize", strconv.FormatUint(req.PageSize, 10)).
		Param("orgID", strconv.FormatInt(req.OrgID, 10)).
		Param("projectID", strconv.FormatUint(req.ProjectID, 10)).
		Param("orderBy", req.OrderBy).Param("asc", "false").
		Params(values).Do().Body(&b)
	if err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg(
			"get issues", "failed to request ("+err.Error()+")", false))

	}

	if !response.IsOK() {
		return nil, formatHTTPFailureFromResponse("list issue", response, b.Bytes())
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("list issue",
			"failed to unmarshal response ("+err.Error()+")", false))

	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("list issue",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))

	}

	return &resp, nil
}

func ListMyIssue(ctx *command.Context, req *apistructs.IssuePagingRequest) (*apistructs.IssuePagingResponseData, error) {
	resp, err := ListMyIssueResponse(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func ListState(ctx *command.Context, orgID uint64, req apistructs.IssueStateRelationGetRequest) ([]apistructs.IssueStateRelation, error) {
	var resp apistructs.IssueStateRelationGetResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/issues/actions/get-state-relations").
		Param("org", strconv.FormatUint(orgID, 10)).
		Param("projectID", strconv.FormatUint(req.ProjectID, 10)).
		Param("issueType", string(req.IssueType)).
		Do().Body(&b)

	if err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg(
			"get state-relations detail", "failed to request ("+err.Error()+")", false))

	}

	if !response.IsOK() {
		return nil, formatHTTPFailureFromResponse("get state-relations", response, b.Bytes())
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("get state-relations",
			"failed to unmarshal response ("+err.Error()+")", false))

	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("get state-relations",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))

	}

	return resp.Data, nil
}

func ListIssueLabels(ctx *command.Context, projectID uint64) ([]apistructs.ProjectLabel, error) {
	var resp apistructs.ProjectLabelListResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/labels").
		Param("projectID", strconv.FormatUint(projectID, 10)).
		Param("type", string(apistructs.LabelTypeIssue)).
		Param("pageNo", "1").
		Param("pageSize", "1000").
		Do().Body(&b)
	if err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg(
			"list issue labels", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return nil, formatHTTPFailureFromResponse("list issue labels", response, b.Bytes())
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("list issue labels",
			"failed to unmarshal response ("+err.Error()+")", false))
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("list issue labels",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	if resp.Data == nil {
		return nil, nil
	}
	return resp.Data.List, nil
}

func ListIssueProperties(ctx *command.Context, orgID uint64, projectID uint64, issueType apistructs.IssueType) ([]apistructs.IssuePropertyIndex, error) {
	var resp apistructs.IssuePropertiesResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/issues/actions/get-properties").
		Param("orgID", strconv.FormatUint(orgID, 10)).
		Param("scopeID", strconv.FormatUint(projectID, 10)).
		Param("propertyIssueType", string(issueType)).
		Do().Body(&b)
	if err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg(
			"list issue properties", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return nil, formatHTTPFailureFromResponse("list issue properties", response, b.Bytes())
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("list issue properties",
			"failed to unmarshal response ("+err.Error()+")", false))
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("list issue properties",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp.Data, nil
}

func ListIssueStages(ctx *command.Context, orgID uint64, issueType apistructs.IssueType) ([]apistructs.IssueStage, error) {
	var resp apistructs.IssueStageResponse
	var b bytes.Buffer

	response, err := ctx.Get().Path("/api/issues/action/get-stage").
		Param("orgID", strconv.FormatUint(orgID, 10)).
		Param("issueType", string(issueType)).
		Do().Body(&b)
	if err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg(
			"list issue stages", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return nil, formatHTTPFailureFromResponse("list issue stages", response, b.Bytes())
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("list issue stages",
			"failed to unmarshal response ("+err.Error()+")", false))
	}

	if !resp.Success {
		return nil, fmt.Errorf("%s", utils.FormatErrMsg("list issue stages",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp.Data, nil
}

type IssuePropertyInstanceCreateRequest struct {
	OrgID     int64                   `json:"orgID"`
	ProjectID int64                   `json:"projectID"`
	IssueID   int64                   `json:"issueID"`
	Property  []IssuePropertyInstance `json:"property"`
}

type IssuePropertyInstance struct {
	PropertyID        int64                        `json:"propertyID"`
	ScopeID           int64                        `json:"scopeID"`
	ScopeType         apistructs.ScopeType         `json:"scopeType"`
	OrgID             int64                        `json:"orgID"`
	PropertyName      string                       `json:"propertyName"`
	DisplayName       string                       `json:"displayName"`
	PropertyType      apistructs.PropertyType      `json:"propertyType"`
	Required          bool                         `json:"required"`
	PropertyIssueType apistructs.PropertyIssueType `json:"propertyIssueType"`
	Relation          int64                        `json:"relation"`
	Index             int64                        `json:"index"`
	EnumeratedValues  []apistructs.Enumerate       `json:"enumeratedValues"`
	Values            []int64                      `json:"values,omitempty"`
	ArbitraryValue    interface{}                  `json:"arbitraryValue,omitempty"`
}

type issuePropertyInstanceCreateResponse struct {
	apistructs.Header
	Data int64 `json:"data"`
}

func CreateIssuePropertyInstance(ctx *command.Context, request *IssuePropertyInstanceCreateRequest) error {
	var resp issuePropertyInstanceCreateResponse
	var b bytes.Buffer

	response, err := ctx.Post().Path("/api/issues/actions/create-property-instance").JSONBody(request).
		Header("org", strconv.FormatInt(request.OrgID, 10)).
		Do().Body(&b)
	if err != nil {
		return fmt.Errorf("%s", utils.FormatErrMsg(
			"create issue property instance", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return formatHTTPFailureFromResponse("create issue property instance", response, b.Bytes())
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return fmt.Errorf("%s", utils.FormatErrMsg("create issue property instance",
			"failed to unmarshal response ("+err.Error()+")", false))
	}

	if !resp.Success {
		return fmt.Errorf("%s", utils.FormatErrMsg("create issue property instance",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return nil
}

func CreateIssueRelation(ctx *command.Context, orgID uint64, parentID uint64, request *apistructs.IssueRelationCreateRequest) error {
	var resp apistructs.Header
	var b bytes.Buffer

	response, err := ctx.Post().Path(fmt.Sprintf("/api/issues/%d/relations", parentID)).
		JSONBody(request).
		Header("org", strconv.FormatUint(orgID, 10)).
		Do().Body(&b)
	if err != nil {
		return fmt.Errorf("%s", utils.FormatErrMsg(
			"create issue relation", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return formatHTTPFailureFromResponse("create issue relation", response, b.Bytes())
	}

	if len(b.Bytes()) == 0 {
		return nil
	}
	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return fmt.Errorf("%s", utils.FormatErrMsg("create issue relation",
			"failed to unmarshal response ("+err.Error()+")", false))
	}

	if !resp.Success {
		return fmt.Errorf("%s", utils.FormatErrMsg("create issue relation",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return nil
}

func GetTodoStateIds(states []apistructs.IssueStateRelation) ([]int64, error) {
	var todoStates []int64
	for _, s := range states {
		switch s.IssueType {
		case apistructs.IssueTypeRequirement,
			apistructs.IssueTypeBug,
			apistructs.IssueTypeTask:
			// goto next switch
		default:
			continue
		}

		switch s.StateBelong {
		case apistructs.IssueStateBelongOpen,
			apistructs.IssueStateBelongWorking,
			apistructs.IssueStateBelongReopen:
			todoStates = append(todoStates, s.StateID)
		}
	}

	return todoStates, nil
}

func CreateIssueComment(ctx *command.Context, orgID uint64, request *pb.CommentIssueStreamBatchCreateRequest) error {
	var resp apistructs.Header
	var b bytes.Buffer

	respponse, err := ctx.Post().Path("/api/issues/actions/batch-create-comment-stream").JSONBody(request).
		Header("org", strconv.FormatUint(orgID, 10)).
		Do().Body(&b)
	if err != nil {
		return fmt.Errorf("%s", utils.FormatErrMsg(
			"create", "failed to request ("+err.Error()+")", false))

	}

	if !respponse.IsOK() {
		return formatHTTPFailureFromResponse("create issue comments", respponse, b.Bytes())
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return fmt.Errorf("%s", utils.FormatErrMsg("create issue comments",
			"failed to unmarshal create response ("+err.Error()+")", false))

	}

	if !resp.Success {
		return fmt.Errorf("%s", utils.FormatErrMsg("create issue comments",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))

	}

	return nil
}

func UpdateIssue(ctx *command.Context, orgID uint64, request *apistructs.IssueUpdateRequest) error {
	var resp apistructs.IssueUpdateResponse
	var b bytes.Buffer

	path := fmt.Sprintf("/api/issues/%d", request.ID)
	response, err := ctx.Put().Path(path).JSONBody(request).
		Header("org", strconv.FormatUint(orgID, 10)).
		Do().Body(&b)
	if err != nil {
		return fmt.Errorf("%s", utils.FormatErrMsg(
			"get issues", "failed to request ("+err.Error()+")", false))

	}

	if !response.IsOK() {
		return formatHTTPFailureFromResponse("update issue", response, b.Bytes())
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return fmt.Errorf("%s", utils.FormatErrMsg("update issue",
			"failed to unmarshal response ("+err.Error()+")", false))

	}

	if !resp.Success {
		return fmt.Errorf("%s", utils.FormatErrMsg("update issue",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))

	}

	return nil
}
