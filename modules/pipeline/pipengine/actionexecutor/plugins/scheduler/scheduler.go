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
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml/pipelineymlv1"
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
		if conf.SchedulerAddr() != "" {
			addr = conf.SchedulerAddr()
			logrus.Infof("=> kind [%v], name [%v], option: %s=%s from env", Kind, name, OPTION_ADDR, addr)
		}
		return &Sched{
			name:    name,
			options: options,
			addr:    addr,
		}, nil
	})
}

type Sched struct {
	name    types.Name
	options map[string]string
	addr    string
}

func (s *Sched) Kind() types.Kind {
	return Kind
}

func (s *Sched) Name() types.Name {
	return s.name
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

	job, err := transferToSchedulerJob(action)
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

	var body bytes.Buffer
	resp, err := httpclient.New().Post(s.addr).
		Path(fmt.Sprintf("/v1/job/%s/%s/start", action.Extra.Namespace, makeJobID(action))).
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

	var body bytes.Buffer
	resp, err := httpclient.New().Get(s.addr, httpclient.RetryErrResp).
		Path(fmt.Sprintf("/v1/job/%s/%s", action.Extra.Namespace, makeJobID(action))).
		Do().Body(&body)
	if err != nil {
		return apistructs.PipelineStatusDesc{}, httpInvokeErr(err)
	}

	statusCode := resp.StatusCode()
	respBody := body.String()

	var result struct {
		Status      string `json:"status"`
		LastMessage string `json:"last_message"`
	}
	if err := json.NewDecoder(&body).Decode(&result); err != nil {
		return apistructs.PipelineStatusDesc{}, respBodyDecodeErr(statusCode, respBody, err)
	}
	if result.Status == "" {
		return apistructs.PipelineStatusDesc{}, errors.Errorf("get empty status from scheduler, respBody: %s", respBody)
	}
	transferredStatus := transferStatus(result.Status)
	logrus.Debugf("pipelineID: %d, taskID: %d, schedulerStatus: %s, transferredStatus: %s, lastMessage: %s",
		action.PipelineID, action.ID, result.Status, transferredStatus, result.LastMessage)
	return apistructs.PipelineStatusDesc{
		Status: transferredStatus,
		Desc:   result.LastMessage,
	}, nil
}

func (s *Sched) Inspect(ctx context.Context, action *spec.PipelineTask) (interface{}, error) {
	return nil, errors.New("scheduler(job) not support inspect operation")
}

func (s *Sched) Cancel(ctx context.Context, action *spec.PipelineTask) (data interface{}, err error) {
	defer wrapError(&err, "cancel job", action)

	if err = validateAction(action); err != nil {
		return nil, err
	}

	var body bytes.Buffer
	resp, err := httpclient.New().Post(s.addr).
		Path(fmt.Sprintf("/v1/job/%s/%s/stop", action.Extra.Namespace, makeJobID(action))).
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

	var body bytes.Buffer
	resp, err := httpclient.New().Delete(s.addr).
		Path(fmt.Sprintf("/v1/job/%s/%s/delete", action.Extra.Namespace, makeJobID(action))).
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

	var req []apistructs.JobFromUser
	for _, action := range actions {
		if len(action.Extra.UUID) <= 0 {
			continue
		}
		req = append(req, apistructs.JobFromUser{Name: action.Extra.UUID, Namespace: action.Extra.Namespace})
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

func transferToSchedulerJob(task *spec.PipelineTask) (job apistructs.JobFromUser, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("%v", r)
		}
	}()

	return apistructs.JobFromUser{
		Name: makeJobID(task),
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
		Image:  task.Extra.Image,
		Cmd:    strings.Join(append([]string{task.Extra.Cmd}, task.Extra.CmdArgs...), " "),
		CPU:    task.Extra.RuntimeResource.CPU,
		Memory: task.Extra.RuntimeResource.Memory,
		Binds:  task.Extra.Binds,
		Volumes: func() []diceyml.Volume {
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
		}(),
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
	return fmt.Sprintf("pipelineID: %d, id: %d, name: %s, namespace: %s, schedulerJobID: %s",
		action.PipelineID, action.ID, action.Name, action.Extra.Namespace, makeJobID(action))
}

// makeJobID 返回 job id。若需要循环，则在 uuid 后追加当前是第几次执行。
func makeJobID(action *spec.PipelineTask) string {
	if action.Extra.LoopOptions != nil && action.Extra.LoopOptions.CalculatedLoop != nil && action.Extra.LoopOptions.CalculatedLoop.Strategy.MaxTimes > 0 {
		return fmt.Sprintf("%s-loop-%d", action.Extra.UUID, action.Extra.LoopOptions.LoopedTimes)
	}
	return action.Extra.UUID
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
