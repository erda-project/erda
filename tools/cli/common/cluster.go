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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/tools/cli/command"
	"github.com/erda-project/erda/tools/cli/format"
)

func GetClusters(ctx *command.Context, orgid string) (apistructs.ClusterListResponse, error) {
	var resp apistructs.ClusterListResponse
	var b bytes.Buffer
	var response *httpclient.Response
	var err error

	if orgid == "" {
		response, err = ctx.Get().Path("/api/clusters").Do().Body(&b)
	} else {
		response, err = ctx.Get().Path("/api/clusters").Param("orgID", orgid).Do().Body(&b)
	}
	if err != nil {
		return apistructs.ClusterListResponse{}, fmt.Errorf(
			format.FormatErrMsg("get clusters", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.ClusterListResponse{}, fmt.Errorf(format.FormatErrMsg("get clusters",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.ClusterListResponse{}, fmt.Errorf(format.FormatErrMsg("get clusters",
			fmt.Sprintf("failed to unmarshal build detail response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.ClusterListResponse{}, fmt.Errorf(format.FormatErrMsg("get clusters",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp, nil
}

func GetClusterDetail(ctx *command.Context, cluster string) (apistructs.GetClusterResponse, error) {
	var resp apistructs.GetClusterResponse
	var b bytes.Buffer

	if cluster == "" {
		return apistructs.GetClusterResponse{}, fmt.Errorf(
			format.FormatErrMsg("get clusters", "missing required arg cluster", false))
	}

	response, err := ctx.Get().Path("/api/clusters/" + cluster).Do().Body(&b)
	if err != nil {
		return apistructs.GetClusterResponse{}, fmt.Errorf(format.FormatErrMsg(
			"get cluster detail", "failed to request ("+err.Error()+")", false))
	}

	if !response.IsOK() {
		return apistructs.GetClusterResponse{}, fmt.Errorf(format.FormatErrMsg("get cluster detail",
			fmt.Sprintf("failed to request, status-code: %d, content-type: %s, raw bod: %s",
				response.StatusCode(), response.ResponseHeader("Content-Type"), b.String()), false))
	}

	if err := json.Unmarshal(b.Bytes(), &resp); err != nil {
		return apistructs.GetClusterResponse{}, fmt.Errorf(format.FormatErrMsg("get cluster detail",
			fmt.Sprintf("failed to unmarshal build detail response ("+err.Error()+")"), false))
	}

	if !resp.Success {
		return apistructs.GetClusterResponse{}, fmt.Errorf(format.FormatErrMsg("get cluster detail",
			fmt.Sprintf("failed to request, error code: %s, error message: %s",
				resp.Error.Code, resp.Error.Msg), false))
	}

	return resp, nil
}
