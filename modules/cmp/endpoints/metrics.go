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
	"github.com/erda-project/erda/pkg/http/httpserver"
)

var permissionFailErr = fmt.Errorf("failed to get User-ID or Org-ID from request header")

// MetricsQuery handle query request
func (e *Endpoints) MetricsQuery(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		req apistructs.MetricsRequest
		err error
	)

	// get identity info
	i, resp := e.GetIdentity(r)
	if resp != nil {
		return mkResponse(apistructs.RmNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: permissionFailErr.Error()},
			},
		})
	}
	// permission check
	err = e.PermissionCheck(i.UserID, i.OrgID, "", apistructs.GetAction)
	if err != nil {
		return nil, permissionFailErr
	}
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		logrus.Errorf("failed to unmarshal request: %+v", err)
		return mkResponse(apistructs.RmNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: err.Error()},
			},
		})
	}
	logrus.Infof("query metrics :%s %s %v", req.ClusterName, req.ResourceType, req.HostName)
	res, err := e.metrics.DoQuery(ctx, req)
	if err != nil {
		return mkResponse(apistructs.RmNodesResponse{
			Header: apistructs.Header{
				Success: false,
				Error:   apistructs.ErrorResponse{Msg: err.Error()},
			},
		})
	}
	return mkResponse(res)
}
