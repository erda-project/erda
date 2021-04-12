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
	"fmt"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
)

// checkPipelineCronPermission 包装校验 cron 权限
func (e *Endpoints) checkPipelineCronPermission(r *http.Request, cronID uint64, permissionAction string) error {
	// 获取 cron 信息
	pc, err := e.dbClient.GetPipelineCron(cronID)
	if err != nil {
		return apierrors.ErrGetPipelineCron.InvalidParameter(err)
	}

	return e.checkBranchPermission(r, strconv.FormatUint(pc.GetAppID(), 10), pc.GetBranch(), permissionAction)
}

// checkPipelineCronPermission 包装校验 app 权限
func (e *Endpoints) checkAppPermission(r *http.Request, appID uint64, permissionAction string) error {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return err
	}
	return e.permissionSvc.CheckApp(identityInfo, appID, permissionAction)
}

// CheckBranch 方便用户直接使用分支进行鉴权
func (e *Endpoints) checkBranchPermission(r *http.Request, appIDStr, branch string, permissionAction string) error {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return err
	}
	if identityInfo.IsInternalClient() {
		return nil
	}
	// TODO adaptor
	return nil
}

// checkPipelineOperatePermission 包装校验 pipeline 编辑操作权限
func (e *Endpoints) checkPipelineOperatePermission(r *http.Request, pipelineID uint64,
	operateRequest apistructs.PipelineOperateRequest, permissionAction string) error {

	// 校验 pipeline 是否合法
	p, err := e.dbClient.GetPipeline(pipelineID)
	if err != nil {
		return apierrors.ErrGetPipeline.InvalidParameter(err)
	}
	// 校验 task 是否属于当前 pipeline
	tasks, err := e.dbClient.ListPipelineTasksByPipelineID(pipelineID)
	if err != nil {
		return apierrors.ErrListPipelineTasks.InvalidParameter(err)
	}
	validTaskMap := make(map[uint64]struct{}, 0)
	for _, task := range tasks {
		validTaskMap[task.ID] = struct{}{}
	}
	for _, taskReq := range operateRequest.TaskOperates {
		if _, ok := validTaskMap[taskReq.TaskID]; !ok {
			return apierrors.ErrCheckPermission.InvalidParameter(
				fmt.Sprintf("task not belong to pipeline, taskID: %d, pipelineID: %d", taskReq.TaskID, pipelineID))
		}
	}

	// 校验分支权限
	return e.checkBranchPermission(r, p.Labels[apistructs.LabelAppID], p.Labels[apistructs.LabelBranch], permissionAction)
}
