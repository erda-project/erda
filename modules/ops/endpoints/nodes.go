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

package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) AddNodes(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.AddNodesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("failed to unmarshal to apistructs.AddNodesRequest: %v", err)
		return mkResponse(apistructs.AddNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	userid := r.Header.Get("User-ID")
	if userid == "" {
		errstr := fmt.Sprintf("failed to get user-id in http header")
		return mkResponse(apistructs.AddNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}

	// permission check
	p := apistructs.PermissionCheckRequest{
		UserID:   userid,
		Scope:    apistructs.OrgScope,
		ScopeID:  req.OrgID,
		Resource: apistructs.CloudResourceResource,
		Action:   apistructs.CreateAction,
	}
	rspData, err := e.bdl.CheckPermission(&p)
	if err != nil {
		logrus.Errorf("check permission error: %v", err)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: "check permission internal error"},
			},
		})
	}
	if !rspData.Access {
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: "access denied"},
			},
		})
	}

	recordID, err := e.nodes.AddNodes(req, userid)
	if err != nil {
		errstr := fmt.Sprintf("failed to add nodes: %v", err)
		return mkResponse(apistructs.AddNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.AddNodesResponse{
		Header: apistructs.Header{Success: true},
		Data:   apistructs.AddNodesData{RecordID: recordID},
	})
}

func (e *Endpoints) RmNodes(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.RmNodesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("failed to unmarshal to apistructs.RmNodesRequest: %v", err)
		return mkResponse(apistructs.RmNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	userid := r.Header.Get("User-ID")
	if userid == "" {
		errstr := fmt.Sprintf("failed to get user-id in http header")
		return mkResponse(apistructs.RmNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	orgid := r.Header.Get("Org-ID")
	scopeID, err := strconv.ParseUint(orgid, 10, 64)
	if err != nil {
		logrus.Errorf("parse orgid failed, orgid: %v, error: %v", orgid, err)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: "parse orgid failed"},
			},
		})
	}
	p := apistructs.PermissionCheckRequest{
		UserID:   userid,
		Scope:    apistructs.OrgScope,
		ScopeID:  scopeID,
		Resource: apistructs.CloudResourceResource,
		Action:   apistructs.DeleteAction,
	}
	rspData, err := e.bdl.CheckPermission(&p)
	if err != nil {
		logrus.Errorf("check permission failed, request: %v, error: %v", p, err)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: "check permission internal error"},
			},
		})
	}
	if !rspData.Access {
		logrus.Errorf("check permission failed, request: %v, response: %v", p, rspData)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: "access denied"},
			},
		})
	}

	recordID, err := e.nodes.RmNodes(req, userid, orgid)
	if err != nil {
		errstr := fmt.Sprintf("failed to rm nodes: %v", err)
		return mkResponse(apistructs.RmNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.RmNodesResponse{
		Header: apistructs.Header{Success: true},
		Data:   apistructs.RmNodesData{RecordID: recordID},
	})
}

func mkResponse(content interface{}) (httpserver.Responser, error) {
	return httpserver.HTTPResponse{
		Status:  http.StatusOK,
		Content: content,
	}, nil
}
