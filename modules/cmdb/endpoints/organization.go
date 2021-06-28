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
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	// job/deployment列表的任务存在时间，默认7天
	TaskCleanDurationTimestamp int64 = 7 * 24 * 60 * 60
)

// ListOrgRunningTasks  指定企业获取服务或者job列表
func (e *Endpoints) ListOrgRunningTasks(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	org, err := e.getOrgByRequest(r)
	if err != nil {
		return apierrors.ErrListClusterAbnormalHosts.InvalidParameter("org id header").ToResp(), nil
	}

	reqParam, err := e.getRunningTasksListParam(r)
	if err != nil {
		return apierrors.ErrListClusterAbnormalHosts.InvalidParameter(err).ToResp(), nil
	}

	total, tasksResults, err := e.org.ListOrgRunningTasks(reqParam, int64(org.ID))
	if err != nil {
		return apierrors.ErrListOrgRunningTasks.InternalError(err).ToResp(), nil
	}

	// insert userID
	userIDs := make([]string, 0, len(tasksResults))
	for _, task := range tasksResults {
		userIDs = append(userIDs, task.UserID)
	}

	return httpserver.OkResp(apistructs.OrgRunningTasksData{Total: total, Tasks: tasksResults},
		strutil.DedupSlice(userIDs, true))
}

func (e *Endpoints) getRunningTasksListParam(r *http.Request) (*apistructs.OrgRunningTasksListRequest, error) {
	// 获取type参数
	taskType := r.URL.Query().Get("type")
	if taskType == "" {
		return nil, errors.Errorf("type")
	}

	if taskType != "job" && taskType != "deployment" {
		return nil, errors.Errorf("type")
	}

	cluster := r.URL.Query().Get("cluster")
	projectName := r.URL.Query().Get("projectName")
	appName := r.URL.Query().Get("appName")
	status := r.URL.Query().Get("status")
	userID := r.URL.Query().Get("userID")
	env := r.URL.Query().Get("env")

	var (
		startTime int64
		endTime   int64
		pipeline  uint64
		err       error
	)
	pipelineID := r.URL.Query().Get("pipelineID")
	if pipelineID != "" {
		pipeline, err = strconv.ParseUint(pipelineID, 10, 64)
		if err != nil {
			return nil, errors.Errorf("convert pipelineID, (%+v)", err)
		}
	}

	// 获取时间范围
	startTimeStr := r.URL.Query().Get("startTime")
	if startTimeStr != "" {
		startTime, err = strutil.Atoi64(startTimeStr)
		if err != nil {
			return nil, err
		}
	}

	endTimeStr := r.URL.Query().Get("endTime")
	if endTimeStr != "" {
		endTime, err = strutil.Atoi64(endTimeStr)
		if err != nil {
			return nil, err
		}
	}

	// 获取pageNo参数
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		return nil, errors.Errorf("invalid param, pageNo: %v", pageNoStr)
	}

	// 获取pageSize参数
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return nil, errors.Errorf("invalid param, pageSize: %v", pageSizeStr)
	}

	return &apistructs.OrgRunningTasksListRequest{
		Cluster:     cluster,
		ProjectName: projectName,
		AppName:     appName,
		PipelineID:  pipeline,
		Status:      status,
		UserID:      userID,
		Env:         env,
		StartTime:   startTime,
		EndTime:     endTime,
		PageNo:      pageNo,
		PageSize:    pageSize,
		Type:        taskType,
	}, nil
}

// DealTaskEvent 接收任务的事件
func (e *Endpoints) DealTaskEvent(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		req           apistructs.PipelineTaskEvent
		runningTaskID int64
		err           error
	)
	if r.Body == nil {
		return apierrors.ErrDealTaskEvents.MissingParameter("body").ToResp(), nil
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrDealTaskEvents.InvalidParameter(err).ToResp(), nil
	}
	logrus.Debugf("ReceiveTaskEvents: request body: %+v", req)

	if req.Event == "pipeline_task" {
		if runningTaskID, err = e.org.DealReceiveTaskEvent(&req); err != nil {
			return apierrors.ErrDealTaskEvents.InvalidParameter(err).ToResp(), nil
		}
	} else if req.Event == "pipeline_task_runtime" {
		if runningTaskID, err = e.org.DealReceiveTaskRuntimeEvent(&req); err != nil {
			return apierrors.ErrDealTaskEvents.InvalidParameter(err).ToResp(), nil
		}
	}

	return httpserver.OkResp(runningTaskID)
}

// SyncTaskStatus 定时同步主机实际使用资源
func (e *Endpoints) SyncTaskStatus(interval time.Duration) {
	l := loop.New(loop.WithInterval(interval))
	l.Do(func() (bool, error) {
		// deal job resource
		jobs := e.db.ListRunningJobs()

		for _, job := range jobs {
			// 根据pipelineID获取task列表信息
			bdl := bundle.New(bundle.WithPipeline())
			pipelineInfo, err := bdl.GetPipeline(job.PipelineID)
			if err != nil {
				logrus.Errorf("failed to get pipeline info by pipelineID, pipelineID:%d, (%+v)", job.PipelineID, err)
				continue
			}

			for _, stage := range pipelineInfo.PipelineStages {
				for _, task := range stage.PipelineTasks {
					if task.ID == job.TaskID {
						if string(task.Status) != job.Status {
							job.Status = string(task.Status)

							// 更新数据库状态
							e.db.UpdateJobStatus(&job)
							logrus.Debugf("update job status, jobID:%d, status:%s", job.ID, job.Status)
						}
					}
				}
			}

		}

		// deal deployment resource
		deployments := e.db.ListRunningDeployments()

		for _, deployment := range deployments {
			// 根据pipelineID获取task列表信息
			bdl := bundle.New(bundle.WithPipeline())
			pipelineInfo, err := bdl.GetPipeline(deployment.PipelineID)
			if err != nil {
				logrus.Errorf("failed to get pipeline info by pipelineID, pipelineID:%d, (%+v)", deployment.PipelineID, err)
				continue
			}

			for _, stage := range pipelineInfo.PipelineStages {
				for _, task := range stage.PipelineTasks {
					if task.ID == deployment.TaskID {
						if string(task.Status) != deployment.Status {
							deployment.Status = string(task.Status)

							// 更新数据库状态
							e.db.UpdateDeploymentStatus(&deployment)
						}
					}
				}
			}

		}

		return false, nil
	})
}

// TaskClean 定时清理任务(job/deployment)资源
func (e *Endpoints) TaskClean(interval time.Duration) {
	l := loop.New(loop.WithInterval(interval))
	l.Do(func() (bool, error) {
		timeUnix := time.Now().Unix()
		fmt.Println(timeUnix)

		startTimestamp := timeUnix - TaskCleanDurationTimestamp

		startTime := time.Unix(startTimestamp, 0).Format("2006-01-02 15:04:05")

		// clean job resource
		jobs := e.db.ListExpiredJobs(startTime)

		for _, job := range jobs {
			err := e.db.DeleteJob(strconv.FormatUint(job.OrgID, 10), job.TaskID)
			if err != nil {
				err = e.db.DeleteJob(strconv.FormatUint(job.OrgID, 10), job.TaskID)
				if err != nil {
					logrus.Errorf("failed to delete job, job: %+v, (%+v)", job, err)
				}
			}
			logrus.Debugf("[clean] expired job: %+v", job)
		}

		// clean deployment resource
		deployments := e.db.ListExpiredDeployments(startTime)

		for _, deployment := range deployments {
			err := e.db.DeleteDeployment(strconv.FormatUint(deployment.OrgID, 10), deployment.TaskID)
			if err != nil {
				err = e.db.DeleteDeployment(strconv.FormatUint(deployment.OrgID, 10), deployment.TaskID)
				if err != nil {
					logrus.Errorf("failed to delete deployment, deployment: %+v, (%+v)", deployment, err)
				}
			}

			logrus.Debugf("[clean] expired deployment: %+v", deployment)
		}

		return false, nil
	})
}
