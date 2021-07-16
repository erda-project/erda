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

package server

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/modules/hepa/common/util"
	"github.com/erda-project/erda/modules/hepa/gateway/dto"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func CheckAuth(c *gin.Context, projectIDStr string) (bool, error) {
	internal := c.GetHeader(httputil.InternalHeader)
	if internal == "bundle" {
		return true, nil
	}
	userID := c.GetHeader("User-ID")
	if userID == "" {
		return false, errors.New("can't get userId from header User-ID")
	}
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		return false, errors.WithStack(err)
	}
	code, body, err := util.CommonRequest("POST", discover.CoreServices()+"/api/permissions/actions/check",
		dto.ApiAuthReqDto{
			UserID:   userID,
			Scope:    "project",
			ScopeID:  projectID,
			Resource: "project",
			Action:   "GET",
		}, map[string]string{"Internal-Client": "hepa-gateway"})
	if err != nil {
		return false, errors.WithStack(err)
	}
	if code >= 300 {
		return false, errors.Errorf("check auth failed, code:%d, body:%s", code, body)
	}
	authResp := &dto.ApiAuthRespDto{}
	err = json.Unmarshal(body, authResp)
	if err != nil {
		return false, errors.WithStack(err)
	}
	return authResp.HasPermission(), nil
}
