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

package pkg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/plugins/scheduler/pkg/k8sjob"
	"github.com/erda-project/erda/modules/pipeline/pkg/task_uuid"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1"
	"github.com/erda-project/erda/pkg/strutil"
)

// use ation type and cluster name judge executor
func judgeExecutorKind(clusterName string) string {
	// TODO complete judge method
	if clusterName == "terminus-dev" {
		return "k8sjob"
	}
	return ""
}

func ForwardStatus(action *spec.PipelineTask) (desc apistructs.StatusDesc, err error) {
	executorName := judgeExecutorKind(action.Extra.ClusterName)
	switch executorName {
	case "k8sjob":
		var executor *k8sjob.K8sJob
		executor, err = k8sjob.New(action.Extra.ClusterName)
		if err != nil {
			return
		}
		return executor.Status(context.Background(), action)
	default:
		var body bytes.Buffer
		resp, err := httpclient.New().Get("scheduler.default.svc.cluster.local:9091", httpclient.RetryErrResp).
			Path(fmt.Sprintf("/v1/job/%s/%s", action.Extra.Namespace, task_uuid.MakeJobID(action))).
			Do().Body(&body)
		if err != nil {
			return desc, httpInvokeErr(err)
		}

		statusCode := resp.StatusCode()
		respBody := body.String()

		//var result struct {
		//	Status      string `json:"status"`
		//	LastMessage string `json:"last_message"`
		//}
		if err := json.NewDecoder(&body).Decode(&desc); err != nil {
			return desc, respBodyDecodeErr(statusCode, respBody, err)
		}
		if desc.Status == "" {
			return desc, errors.Errorf("get empty status from scheduler, respBody: %s", respBody)
		}
		return desc, nil
	}
}

func ForwardCreate(action *spec.PipelineTask) (data interface{}, err error) {
	var job apistructs.JobFromUser
	job, err = transferToSchedulerJob(action)
	if err != nil {
		return nil, err
	}
	executorName := judgeExecutorKind(job.ClusterName)
	switch executorName {
	case "k8sjob":
		return nil, nil
	default:
		var body bytes.Buffer
		resp, err := httpclient.New().Put("scheduler.default.svc.cluster.local:9091").
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
}

func ForwardStart(action *spec.PipelineTask) (data interface{}, err error) {
	executorName := judgeExecutorKind(action.Extra.ClusterName)
	switch executorName {
	case "k8sjob":
		var executor *k8sjob.K8sJob
		executor, err = k8sjob.New(action.Extra.ClusterName)
		if err != nil {
			return nil, err
		}
		return executor.Create(context.Background(), action)
	default:
		var body bytes.Buffer
		resp, err := httpclient.New().Post("scheduler.default.svc.cluster.local:9091").
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
}

func respBodyDecodeErr(statusCode int, respBody string, err error) error {
	return errors.Errorf("statusCode: %d, respBody: %s, err: %v", statusCode, respBody, err)
}

func httpInvokeErr(err error) error {
	return errors.Errorf("http invoke err: %v", err)
}

func transferToSchedulerJob(task *spec.PipelineTask) (job apistructs.JobFromUser, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("%v", r)
		}
	}()

	return apistructs.JobFromUser{
		Name: task_uuid.MakeJobID(task),
		Kind: func() string {
			switch task.Type {
			case string(pipelineymlv1.RES_TYPE_FLINK):
				return string(apistructs.Flink)
			case string(pipelineymlv1.RES_TYPE_SPARK):
				return string(apistructs.Spark)
			default:
				return ""
			}
		}(),
		Namespace: task.Extra.Namespace,
		ClusterName: func() string {
			if len(task.Extra.ClusterName) == 0 {
				panic(errors.New("missing cluster name in pipeline task"))
			}
			return task.Extra.ClusterName
		}(),
		Image:      task.Extra.Image,
		Cmd:        strings.Join(append([]string{task.Extra.Cmd}, task.Extra.CmdArgs...), " "),
		CPU:        task.Extra.RuntimeResource.CPU,
		Memory:     task.Extra.RuntimeResource.Memory,
		Binds:      task.Extra.Binds,
		Volumes:    makeVolume(task),
		PreFetcher: task.Extra.PreFetcher,
		Env:        task.Extra.PublicEnvs,
		Labels:     task.Extra.Labels,
		// flink/spark
		Resource:  task.Extra.FlinkSparkConf.JarResource,
		MainClass: task.Extra.FlinkSparkConf.MainClass,
		MainArgs:  task.Extra.FlinkSparkConf.MainArgs,
		// 重试不依赖 scheduler，由 pipeline engine 自己实现，保证所有 action executor 均适用
		Params: task.Extra.Action.Params,
	}, nil
}

func makeVolume(task *spec.PipelineTask) []diceyml.Volume {
	diceVolumes := make([]diceyml.Volume, 0)
	for _, vo := range task.Extra.Volumes {
		if vo.Type == string(spec.StoreTypeDiceVolumeFake) || vo.Type == string(spec.StoreTypeDiceCacheNFS) {
			// fake volume,没有实际挂载行为,不传给scheduler
			continue
		}
		diceVolume := diceyml.Volume{
			Path: vo.Value,
			Storage: func() string {
				switch vo.Type {
				case string(spec.StoreTypeDiceVolumeNFS):
					return "nfs"
				case string(spec.StoreTypeDiceVolumeLocal):
					return "local"
				default:
					panic(errors.Errorf("%q has not supported volume type: %s", vo.Name, vo.Type))
				}
			}(),
		}
		if vo.Labels != nil {
			if id, ok := vo.Labels["ID"]; ok {
				diceVolume.ID = &id
				goto AppendDiceVolume
			}
		}
		// labels == nil or labels["ID"] not exist
		// 如果 id 不存在，说明上一次没有生成 volume，并且是 optional 的，则不创建 diceVolume
		if vo.Optional {
			continue
		}
	AppendDiceVolume:
		diceVolumes = append(diceVolumes, diceVolume)
	}
	return diceVolumes
}

func printActionInfo(action *spec.PipelineTask) string {
	return fmt.Sprintf("pipelineID: %d, id: %d, name: %s, namespace: %s, schedulerJobID: %s",
		action.PipelineID, action.ID, action.Name, action.Extra.Namespace, task_uuid.MakeJobID(action))
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
