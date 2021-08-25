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

package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor"
	tasktypes "github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/executor/types"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/logic"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/pkg/task_uuid"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/strutil"
)

var Kind = types.Kind(spec.PipelineTaskExecutorKindScheduler)

const (
	OPTION_ADDR = "ADDR"

	notFoundError = "not found"
)

var (
	errMissingNamespace = errors.New("action missing namespace")
	errMissingUUID      = errors.New("action missing UUID")
)

func init() {
	types.MustRegister(Kind, func(name types.Name, options map[string]string) (types.ActionExecutor, error) {
		addr, ok := options[OPTION_ADDR]
		if !ok {
			return nil, errors.Errorf("not found some config of action executor, kind [%s], name [%s], field [ADDR]", Kind, name)
		}
		if discover.Scheduler() != "" {
			addr = discover.Scheduler()
			logrus.Infof("=> kind [%v], name [%v], option: %s=%s from env", Kind, name, OPTION_ADDR, addr)
		}

		mgr := executor.GetManager()
		if err := mgr.Initialize(); err != nil {
			return nil, err
		}
		return &Sched{
			name:        name,
			options:     options,
			addr:        addr,
			taskManager: mgr,
		}, nil
	})
}

type Sched struct {
	name        types.Name
	options     map[string]string
	addr        string
	taskManager *executor.Manager
}

func (s *Sched) Kind() types.Kind {
	return Kind
}

func (s *Sched) Name() types.Name {
	return s.name
}

// GetTaskExecutor return bool, task exectuor, error, bool means should it be dispatch to scheduler
func (s *Sched) GetTaskExecutor(executorType string, clusterName string, task *spec.PipelineTask) (bool, tasktypes.TaskExecutor, error) {
	var executorName string
	cluster, err := s.taskManager.GetCluster(clusterName)
	if err != nil {
		return false, nil, err
	}
	switch cluster.Type {
	case apistructs.DCOS:
		return true, nil, nil
	//case apistructs.EDAS:
	//	return true, nil, nil
	case apistructs.K8S, apistructs.EDAS:
		if executorType == "flink" || executorType == "spark" {
			return false, nil, errors.Errorf("k8s cluster don`t support executor type: %s", executorType)
		}
		executorName = "k8sjob"
		if value, ok := task.Extra.Action.Params["bigDataConf"]; ok {
			spec := apistructs.BigdataSpec{}
			if err := json.Unmarshal([]byte(value.(string)), &spec); err != nil {
				return false, nil, fmt.Errorf("failed to unmarshal task bigDataConf")
			}
			if spec.FlinkConf != nil {
				executorName = "k8sflink"
			}
			if spec.SparkConf != nil {
				executorName = "k8sspark"
			}
		}
		name := fmt.Sprintf("%sfor%s", clusterName, executorName)
		return s.taskManager.TryGetExecutor(tasktypes.Name(name), cluster)
	default:
		return false, nil, errors.Errorf("invalid cluster type: %s", cluster.Type)
	}
}

func validateAction(action *spec.PipelineTask) error {
	if action.Extra.Namespace == "" {
		return errMissingNamespace
	}
	if action.Extra.UUID == "" {
		return errMissingUUID
	}
	return nil
}

// Exist 返回 job 存在情况
// created: 调用 create 成功，job 在 etcd 中已创建
// started: 调用 start 成功，job 在 cluster 中已存在并开始执行
func (s *Sched) Exist(ctx context.Context, action *spec.PipelineTask) (created, started bool, err error) {
	statusDesc, err := s.Status(ctx, action)
	if err != nil {
		created = false
		started = false
		// 该 ErrMsg 表示记录在 etcd 中不存在，即未创建
		if strutil.Contains(err.Error(), "failed to inspect job, err: not found") {
			err = nil
			return
		}
		// 获取 job 状态失败
		return
	}
	// err 为空，说明在 etcd 中存在记录，即已经创建成功
	created = true

	// 根据状态判断是否实际 job(k8s job, DC/OS job) 是否已开始执行
	switch statusDesc.Status {
	// err
	case apistructs.PipelineStatusError, apistructs.PipelineStatusUnknown:
		err = errors.Errorf("failed to judge job exist or not, detail: %s", statusDesc)
	// not started
	case apistructs.PipelineStatusCreated, apistructs.PipelineStatusStartError:
		started = false
	// started
	case apistructs.PipelineStatusQueue, apistructs.PipelineStatusRunning,
		apistructs.PipelineStatusSuccess, apistructs.PipelineStatusFailed,
		apistructs.PipelineStatusStopByUser:
		started = true

	// default
	default:
		started = false
	}
	return
}

func (s *Sched) Create(ctx context.Context, action *spec.PipelineTask) (data interface{}, err error) {
	defer wrapError(&err, "create job", action)

	if err = validateAction(action); err != nil {
		return nil, err
	}

	created, _, err := s.Exist(ctx, action)
	if err != nil {
		return nil, err
	}
	if created {
		logrus.Warnf("scheduler: action already created, actionInfo: %s", printActionInfo(action))
		return nil, nil
	}

	var taskExecutor tasktypes.TaskExecutor
	var shouldDispatch bool
	shouldDispatch, taskExecutor, err = s.GetTaskExecutor(action.Type, action.Extra.ClusterName, action)
	if err != nil {
		return nil, err
	}
	if !shouldDispatch {
		logrus.Infof("task executor %s execute create", taskExecutor.Name())
		return nil, nil
	}

	job, err := logic.TransferToSchedulerJob(action)
	if err != nil {
		return nil, errors.Errorf("transfer to scheduler job err: %v", err)
	}

	var body bytes.Buffer
	resp, err := httpclient.New().Put(s.addr).
		Path("/v1/job/create").JSONBody(apistructs.JobCreateRequest(job)).
		Do().Body(&body)
	if err != nil {
		return nil, httpInvokeErr(err)
	}

	statusCode := resp.StatusCode()
	respBody := body.String()

	var result apistructs.JobCreateResponse
	err = json.Unmarshal([]byte(respBody), &result)
	if err != nil {
		return nil, respBodyDecodeErr(statusCode, respBody, err)
	}
	logrus.Debugf("scheduler: invoke scheduler to create task, pipelineID: %d, actionInfo: %s, statusCode: %d, respBody: %s",
		action.PipelineID, printActionInfo(action), statusCode, respBody)
	if result.Error != "" {
		// 幂等
		if isJobIdempotentErrMsg(result.Error) {
			logrus.Warnf("scheduler: action already created, pipelineID: %d, actionInfo: %s, err: %v",
				action.PipelineID, printActionInfo(action), result.Error)
			return nil, nil
		}
		return nil, errors.Errorf("statusCode: %d, result.error: %s", statusCode, result.Error)
	}

	return result.Job, nil
}

func (s *Sched) Start(ctx context.Context, action *spec.PipelineTask) (data interface{}, err error) {
	defer wrapError(&err, "start job", action)

	if err = validateAction(action); err != nil {
		return nil, err
	}

	created, started, err := s.Exist(ctx, action)
	if err != nil {
		return nil, err
	}
	if !created {
		logrus.Warnf("scheduler: action not create yet, try to create, actionInfo: %s", printActionInfo(action))
		_, err = s.Create(ctx, action)
		if err != nil {
			return nil, err
		}
		logrus.Warnf("scheduler: action created, continue to start, actionInfo: %s", printActionInfo(action))
	}
	if started {
		logrus.Warnf("scheduler: action already started, actionInfo: %s", printActionInfo(action))
		return nil, nil
	}

	var taskExecutor tasktypes.TaskExecutor
	var shouldDispatch bool
	shouldDispatch, taskExecutor, err = s.GetTaskExecutor(action.Type, action.Extra.ClusterName, action)
	if err != nil {
		return nil, err
	}
	if !shouldDispatch {
		logrus.Infof("task executor %s execute start", taskExecutor.Name())
		return taskExecutor.Create(ctx, action)
	}

	var body bytes.Buffer
	resp, err := httpclient.New().Post(s.addr).
		Path(fmt.Sprintf("/v1/job/%s/%s/start", action.Extra.Namespace, task_uuid.MakeJobID(action))).
		Do().Body(&body)
	if err != nil {
		return nil, errors.Errorf("http invoke err: %v", err)
	}

	statusCode := resp.StatusCode()
	respBody := body.String()

	var result apistructs.JobStartResponse
	err = json.Unmarshal([]byte(respBody), &result)
	if err != nil {
		return nil, respBodyDecodeErr(statusCode, respBody, err)
	}
	logrus.Debugf("scheduler: invoke scheduler to start task, pipelineID: %d, actionInfo: %s, statusCode: %d, respBody: %s",
		action.PipelineID, printActionInfo(action), statusCode, respBody)
	if result.Error != "" {
		// 幂等
		if isJobIdempotentErrMsg(result.Error) {
			logrus.Warnf("scheduler: action already started, pipelineID: %d, actionInfo: %s, result.error: %s",
				action.PipelineID, printActionInfo(action), result.Error)
			return nil, nil
		}
		return nil, errors.Errorf("statusCode: %d, result.error: %s", statusCode, result.Error)
	}

	return result.Job, nil
}

func (s *Sched) Update(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	return nil, errors.New("scheduler(job) not support update operation")
}

func (s *Sched) Status(ctx context.Context, action *spec.PipelineTask) (desc apistructs.PipelineStatusDesc, err error) {
	defer wrapError(&err, "status job", action)

	if err = validateAction(action); err != nil {
		return apistructs.PipelineStatusDesc{}, err
	}

	var result apistructs.StatusDesc
	var taskExecutor tasktypes.TaskExecutor
	var shouldDispatch bool
	shouldDispatch, taskExecutor, err = s.GetTaskExecutor(action.Type, action.Extra.ClusterName, action)
	if err != nil {
		return apistructs.PipelineStatusDesc{}, err
	}
	if !shouldDispatch {
		logrus.Infof("task executor %s execute status", taskExecutor.Name())
		result, err = taskExecutor.Status(ctx, action)
		if err != nil {
			return apistructs.PipelineStatusDesc{}, err
		}
	} else {
		var body bytes.Buffer
		resp, err := httpclient.New().Get(s.addr, httpclient.RetryErrResp).
			Path(fmt.Sprintf("/v1/job/%s/%s", action.Extra.Namespace, task_uuid.MakeJobID(action))).
			Do().Body(&body)
		if err != nil {
			return apistructs.PipelineStatusDesc{}, httpInvokeErr(err)
		}

		statusCode := resp.StatusCode()
		respBody := body.String()

		if err := json.NewDecoder(&body).Decode(&result); err != nil {
			return apistructs.PipelineStatusDesc{}, respBodyDecodeErr(statusCode, respBody, err)
		}
	}
	if result.Status == "" {
		return apistructs.PipelineStatusDesc{}, errors.Errorf("get empty status from scheduler, statusCode: %s, lastMsg: %s", result.Status, result.LastMessage)
	}
	transferredStatus := transferStatus(string(result.Status))
	logrus.Debugf("pipelineID: %d, taskID: %d, schedulerStatus: %s, transferredStatus: %s, lastMessage: %s",
		action.PipelineID, action.ID, result.Status, transferredStatus, result.LastMessage)
	return apistructs.PipelineStatusDesc{
		Status: transferredStatus,
		Desc:   result.LastMessage,
	}, nil
}

func (s *Sched) Inspect(ctx context.Context, action *spec.PipelineTask) (apistructs.TaskInspect, error) {
	var (
		taskExecutor   tasktypes.TaskExecutor
		shouldDispatch bool
		err            error
	)
	shouldDispatch, taskExecutor, err = s.GetTaskExecutor(action.Type, action.Extra.ClusterName, action)
	if err != nil {
		return apistructs.TaskInspect{}, err
	}
	if !shouldDispatch {
		logrus.Infof("task executor %s execute inspect", taskExecutor.Name())
		return taskExecutor.Inspect(ctx, action)
	}
	return apistructs.TaskInspect{}, errors.New("scheduler(job) not support inspect operation")
}

func (s *Sched) Cancel(ctx context.Context, action *spec.PipelineTask) (data interface{}, err error) {
	defer wrapError(&err, "cancel job", action)

	if err = validateAction(action); err != nil {
		return nil, err
	}

	var taskExecutor tasktypes.TaskExecutor
	var shouldDispatch bool
	shouldDispatch, taskExecutor, err = s.GetTaskExecutor(action.Type, action.Extra.ClusterName, action)
	if err != nil {
		return nil, err
	}
	if !shouldDispatch {
		logrus.Infof("task executor %s execute cancel", taskExecutor.Name())
		// TODO move all makeJobID to framework
		// now move makeJobID to framework may change task uuid in database
		// Restore the task uuid after remove, because gc will make the job id, but cancel don't make the job id
		oldUUID := action.Extra.UUID
		action.Extra.UUID = task_uuid.MakeJobID(action)
		d, err := taskExecutor.Remove(ctx, action)
		action.Extra.UUID = oldUUID
		return d, err
	}
	var body bytes.Buffer
	resp, err := httpclient.New().Post(s.addr).
		Path(fmt.Sprintf("/v1/job/%s/%s/stop", action.Extra.Namespace, task_uuid.MakeJobID(action))).
		Do().Body(&body)
	if err != nil {
		return nil, httpInvokeErr(err)
	}

	statusCode := resp.StatusCode()
	respBody := body.String()

	var result apistructs.JobStopResponse
	if err := json.NewDecoder(&body).Decode(&result); err != nil {
		return nil, respBodyDecodeErr(statusCode, respBody, err)
	}
	if result.Error != "" {
		return nil, errors.New(result.Error)
	}
	return result.Name, nil
}

func (s *Sched) Remove(ctx context.Context, action *spec.PipelineTask) (data interface{}, err error) {
	defer wrapError(&err, "remove job", action)

	if err = validateAction(action); err != nil {
		return nil, err
	}

	var taskExecutor tasktypes.TaskExecutor
	var shouldDispatch bool
	shouldDispatch, taskExecutor, err = s.GetTaskExecutor(action.Type, action.Extra.ClusterName, action)
	if err != nil {
		return nil, err
	}
	if !shouldDispatch {
		// TODO move all makeJobID to framework
		// now move makeJobID to framework may change task uuid in database
		action.Extra.UUID = task_uuid.MakeJobID(action)
		logrus.Infof("task executor %s execute remove", taskExecutor.Name())
		return taskExecutor.Remove(ctx, action)
	}

	var body bytes.Buffer
	resp, err := httpclient.New().Delete(s.addr).
		Path(fmt.Sprintf("/v1/job/%s/%s/delete", action.Extra.Namespace, task_uuid.MakeJobID(action))).
		Do().Body(&body)
	if err != nil {
		return nil, httpInvokeErr(err)
	}

	var result apistructs.JobDeleteResponse
	if err := json.NewDecoder(&body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Error != "" {
		if strings.Contains(result.Error, notFoundError) {
			logrus.Warnf("skip resp.Error(not found) when invoke scheduler.remove, taskID: %d, pipelineID: %d, resp.Error: %s",
				action.ID, action.PipelineID, result.Error)
			return result.Name, nil
		}
		return nil, errors.Errorf("statusCode: %d, resp.error: %s", resp.StatusCode(), result.Error)
	}
	return result.Name, nil
}

func (s *Sched) BatchDelete(ctx context.Context, actions []*spec.PipelineTask) (data interface{}, err error) {
	if len(actions) == 0 {
		return nil, nil
	}

	action := actions[0]

	defer wrapError(&err, "batch delete jobs", action)

	if err = validateAction(action); err != nil {
		return nil, err
	}

	var taskExecutor tasktypes.TaskExecutor
	var shouldDispatch bool
	shouldDispatch, taskExecutor, err = s.GetTaskExecutor(action.Type, action.Extra.ClusterName, action)
	if err != nil {
		return nil, err
	}
	if !shouldDispatch {
		logrus.Infof("task executor %s execute batch delete", taskExecutor.Name())
		return taskExecutor.BatchDelete(ctx, actions)
	}

	var req []apistructs.JobFromUser
	for _, action := range actions {
		if len(action.Extra.UUID) <= 0 {
			continue
		}
		req = append(req, apistructs.JobFromUser{
			Name:        action.Extra.UUID,
			Namespace:   action.Extra.Namespace,
			ClusterName: action.Extra.ClusterName,
			Volumes:     logic.MakeVolume(action),
		})
	}
	var body bytes.Buffer
	resp, err := httpclient.New().Delete(s.addr).
		Path("/v1/jobs").
		JSONBody(&req).
		Do().Body(&body)
	if err != nil {
		return nil, httpInvokeErr(err)
	}

	statusCode := resp.StatusCode()
	respBody := body.String()

	var results apistructs.JobsDeleteResponse
	if err := json.NewDecoder(&body).Decode(&results); err != nil {
		return nil, respBodyDecodeErr(statusCode, respBody, err)
	}
	var filteredErrResults apistructs.JobsDeleteResponse
	for i := range results {
		result := results[i]
		if result.Error == "" {
			continue
		}
		if strings.Contains(result.Error, notFoundError) {
			logrus.Infof("skip resp.Error(not found) when invoke scheduler.batchDelete, pipelineID: %d, namespace: %s, taskName: %v, resp.Error: %s",
				action.PipelineID, result.Namespace, result.Name, result.Error)
			continue
		}
		filteredErrResults = append(filteredErrResults, result)
	}
	if len(filteredErrResults) > 0 {
		return nil, fmt.Errorf("statusCode: %d, results: %+v", resp.StatusCode(), filteredErrResults)
	}
	return "", nil
}

func transferStatus(status string) apistructs.PipelineStatus {
	switch status {

	case string(apistructs.StatusError):
		return apistructs.PipelineStatusError

	case string(apistructs.StatusUnknown):
		return apistructs.PipelineStatusUnknown

	case string(apistructs.StatusCreated):
		return apistructs.PipelineStatusCreated

	case string(apistructs.StatusUnschedulable), "INITIAL":
		return apistructs.PipelineStatusQueue

	case string(apistructs.StatusRunning), "ACTIVE":
		return apistructs.PipelineStatusRunning

	case string(apistructs.StatusStoppedOnOK), string(apistructs.StatusFinished):
		return apistructs.PipelineStatusSuccess

	case string(apistructs.StatusStoppedOnFailed), string(apistructs.StatusFailed):
		return apistructs.PipelineStatusFailed

	case string(apistructs.StatusStoppedByKilled):
		return apistructs.PipelineStatusStopByUser

	case string(apistructs.StatusNotFoundInCluster):
		// scheduler 返回 job 在 cluster 中不存在 (在 etcd 中存在)，对应为 启动错误
		// 典型场景：created 成功，env key 为数字，导致 start job 时真正去创建 k8s job 时失败，即启动失败
		return apistructs.PipelineStatusStartError
	}

	return apistructs.PipelineStatusUnknown
}

func wrapError(err *error, op string, action *spec.PipelineTask) {
	if err == nil || *err == nil {
		return
	}
	*err = errors.Errorf("failed to invoke scheduler to %s, actionInfo: %s, err: %v", op, printActionInfo(action), *err)
}

func httpInvokeErr(err error) error {
	return errors.Errorf("http invoke err: %v", err)
}

func respBodyDecodeErr(statusCode int, respBody string, err error) error {
	return errors.Errorf("statusCode: %d, respBody: %s, err: %v", statusCode, respBody, err)
}

func printActionInfo(action *spec.PipelineTask) string {
	return fmt.Sprintf("pipelineID: %d, id: %d, name: %s, namespace: %s, schedulerJobID: %s, clusterName: %s",
		action.PipelineID, action.ID, action.Name, action.Extra.Namespace, task_uuid.MakeJobID(action), action.Extra.ClusterName)
}

// isJobIdempotent
func isJobIdempotentErrMsg(errMsg string) bool {
	// polish errMsg
	errMsg = strings.NewReplacer(`\\`, `\`, `\"`, `"`, `\'`, `'`).Replace(errMsg)

	// "code":409,"reason":"AlreadyExists"
	if strutil.Contains(errMsg, `"code":409`) {
		// vendor/k8s.io/apimachinery/pkg/apis/meta/v1/types.go:726 StatusReasonAlreadyExists
		if strutil.Contains(errMsg, `"reason":"AlreadyExists"`) {
			return true
		}
	}

	// job is running
	if strutil.Contains(strutil.ToLower(errMsg), apistructs.ErrJobIsRunning.Error()) {
		return true
	}

	return false
}
