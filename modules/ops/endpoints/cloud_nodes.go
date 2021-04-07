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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) AddCloudNodes(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.CloudNodesRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errstr := fmt.Sprintf("failed to unmarshal to apistructs.CloudNodesRequest: %v", err)
		return mkResponse(apistructs.CloudNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	logrus.Debugf("cloud-node request: %v", req)

	// get identity info: user/org id
	i, resp := e.GetIdentity(r)
	if resp != nil {
		return resp, nil
	}
	// permission check
	err := e.PermissionCheck(i.UserID, i.OrgID, "", apistructs.CreateAction)
	if err != nil {
		logrus.Errorf("check permission error: %v", err)
		return mkResponse(apistructs.CloudClusterResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: "check permission internal error"},
			},
		})
	}

	var recordID uint64
	if req.CloudVendor == string(apistructs.CloudVendorAliEcs) {
		recordID, err = e.nodes.AddCloudNodes(req, i.UserID)
	} else if req.CloudVendor == string(apistructs.CloudVendorAliAck) { // TODO remove
		recordID, err = e.nodes.AddCSNodes(req, i.UserID)
	} else if req.CloudVendor == string(apistructs.CloudVendorAliCS) {
		recordID, err = e.nodes.AddCSNodes(req, i.UserID)
	} else if req.CloudVendor == string(apistructs.CloudVendorAliCSManaged) {
		recordID, err = e.nodes.AddCSNodes(req, i.UserID)
	} else {
		err = fmt.Errorf("cloud vendor:%v is not valid", req.CloudVendor)
	}
	if err != nil {
		errstr := fmt.Sprintf("failed to add nodes: %v", err)
		return mkResponse(apistructs.CloudNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: errstr},
			},
		})
	}
	return mkResponse(apistructs.CloudNodesResponse{
		Header: apistructs.Header{Success: true},
		Data:   apistructs.AddNodesData{RecordID: recordID},
	})
}
