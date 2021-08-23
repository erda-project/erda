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
