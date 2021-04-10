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
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/hepa/bundle"
)

type ScopeInfo struct {
	ProjectId   string
	Workspace   string
	AppId       string
	ServiceName string
	RuntimeName string
}

func MakeAuditInfo(reqCtx *gin.Context, scopeInfo ScopeInfo, name apistructs.TemplateName, res *StandardResult, ctx map[string]interface{}) *apistructs.Audit {
	if reqCtx == nil {
		return nil
	}
	orgId := reqCtx.GetHeader("Org-ID")
	if orgId == "" {
		return nil
	}
	userId := reqCtx.GetHeader("User-ID")
	if userId == "" {
		return nil
	}
	id, err := strconv.ParseUint(scopeInfo.ProjectId, 10, 64)
	if err != nil {
		log.Errorf("parse failed, err:%+v", errors.WithStack(err))
		return nil
	}
	projectInfo, err := bundle.Bundle.GetProject(id)
	if err != nil {
		log.Errorf("get project failed, err:%+v", errors.WithStack(err))
		return nil
	}
	ctx["projectId"] = scopeInfo.ProjectId
	ctx["project"] = projectInfo.Name
	audit := &apistructs.Audit{
		UserID:       userId,
		ScopeType:    "project",
		ScopeID:      id,
		TemplateName: name,
		StartTime:    strconv.FormatInt(reqCtx.MustGet("startTime").(int64), 10),
		ClientIP:     reqCtx.ClientIP(),
	}
	audit.ProjectID = id
	id, err = strconv.ParseUint(orgId, 10, 64)
	if err != nil {
		return nil
	}
	audit.OrgID = id
	if scopeInfo.AppId != "" {
		ctx["appId"] = scopeInfo.AppId
		id, _ = strconv.ParseUint(scopeInfo.AppId, 10, 64)
		appInfo, err := bundle.Bundle.GetApp(id)
		if err != nil {
			log.Errorf("get app fialed, err:%+v", err)
		}
		ctx["app"] = appInfo.Name
		audit.AppID = id
	}
	if scopeInfo.Workspace != "" {
		ctx["workspace"] = strings.ToUpper(scopeInfo.Workspace)
	}
	if scopeInfo.ServiceName != "" {
		ctx["service"] = scopeInfo.ServiceName
	}
	if scopeInfo.RuntimeName != "" {
		ctx["runtime"] = scopeInfo.RuntimeName
	}
	if res != nil {
		if res.Success {
			audit.Result = apistructs.SuccessfulResult
		} else {
			//audit.Result = apistructs.FailureResult
			return nil
		}
		// if res.Err != nil {
		// 	audit.ErrorMsg = res.Err.Msg
		// }
	}
	audit.Context = ctx
	audit.EndTime = strconv.FormatInt(time.Now().Unix(), 10)
	return audit
}
