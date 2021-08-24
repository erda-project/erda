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

package migration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/orchestrator/dbclient"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

// Migration migration 实例对象封装
type Migration struct {
	db  *dbclient.DBClient
	bdl *bundle.Bundle
}

// Option Migration 实例对象配置选项
type Option func(*Migration)

// New 新建 Migration service
func New(options ...Option) *Migration {
	var migration Migration
	for _, op := range options {
		op(&migration)
	}

	return &migration
}

// WithDBClient 配置 db client
func WithDBClient(db *dbclient.DBClient) Option {
	return func(a *Migration) {
		a.db = db
	}
}

// WithBundle 配置 bundle
func WithBundle(bdl *bundle.Bundle) Option {
	return func(a *Migration) {
		a.bdl = bdl
	}
}

func (m *Migration) Create(migrationLog *dbclient.MigrationLog, diceyml *diceyml.DiceYaml, Runtime *dbclient.Runtime, App *apistructs.ApplicationDTO) (data interface{}, err error) {

	job, err := m.transferToSchedulerJob(migrationLog, diceyml, Runtime, App)
	if err != nil {
		return nil, errors.Errorf("transfer to scheduler job err: %v", err)
	}
	bb, err := json.Marshal(job)
	logrus.Infof("Create schedule job body: %s", string(bb))
	var body bytes.Buffer
	resp, err := httpclient.New().Put(discover.Scheduler()).
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
	if result.Error != "" {
		return nil, errors.Errorf("statusCode: %d, result.error: %s", statusCode, result.Error)
	}

	return result.Job, nil
}

func (m *Migration) Start(migrationLog *dbclient.MigrationLog, diceyml *diceyml.DiceYaml, Runtime *dbclient.Runtime, App *apistructs.ApplicationDTO) (err error) {
	created, started, err := m.Exist(migrationLog)
	if err != nil {
		return err
	}
	if !created {
		logrus.Warnf("scheduler: migration job not create yet, try to create, migration info: %+v", *migrationLog)
		_, err = m.Create(migrationLog, diceyml, Runtime, App)
		if err != nil {
			return err
		}
		logrus.Warnf("scheduler: migration job created, continue to start, migration info: %+v", *migrationLog)
	}
	if started {
		logrus.Warnf("scheduler: migration job already started, migration info: %+v", *migrationLog)
		return nil
	}

	namespace, name := getNamespaceAndName(migrationLog)
	var body bytes.Buffer
	resp, err := httpclient.New().Post(discover.Scheduler()).
		Path(fmt.Sprintf("/v1/job/%s/%s/start", namespace, name)).
		Do().Body(&body)
	if err != nil {
		return errors.Errorf("http invoke err: %v", err)
	}

	statusCode := resp.StatusCode()
	respBody := body.String()

	var result apistructs.JobStartResponse
	err = json.Unmarshal([]byte(respBody), &result)
	if err != nil {
		return respBodyDecodeErr(statusCode, respBody, err)
	}
	logrus.Debugf("scheduler: invoke scheduler to start migration task, namespace: %s, statusCode: %d, respBody: %s",
		namespace, statusCode, respBody)
	if result.Error != "" {
		return errors.Errorf("statusCode: %d, result.error: %s", statusCode, result.Error)
	}

	return nil
}

// Exist 返回 job 存在情况
// created: 调用 create 成功，job 在 etcd 中已创建
// started: 调用 start 成功，job 在 cluster 中已存在并开始执行
func (m *Migration) Exist(migrationLog *dbclient.MigrationLog) (created, started bool, err error) {
	statusDesc, err := m.Status(migrationLog)
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
	case apistructs.StatusError, apistructs.StatusUnknown:
		err = errors.Errorf("failed to judge job exist or not, detail: %s", statusDesc)
	// not started
	case apistructs.StatusCreated, apistructs.StatusNotFoundInCluster:
		started = false
	// started
	case apistructs.StatusUnschedulable, apistructs.StatusRunning,
		apistructs.StatusStoppedOnOK, apistructs.StatusFinished,
		apistructs.StatusStoppedOnFailed, apistructs.StatusFailed,
		apistructs.StatusStoppedByKilled:
		started = true

	// default
	default:
		started = false
	}
	return
}

// Status 获取migration status信息
func (m *Migration) Status(migrationLog *dbclient.MigrationLog) (desc apistructs.MigrationStatusDesc, err error) {
	namespace, name := getNamespaceAndName(migrationLog)
	var body bytes.Buffer
	resp, err := httpclient.New().Get(discover.Scheduler(), httpclient.RetryErrResp).
		Path(fmt.Sprintf("/v1/job/%s/%s", namespace, name)).
		Do().Body(&body)
	if err != nil {
		return apistructs.MigrationStatusDesc{}, httpInvokeErr(err)
	}

	statusCode := resp.StatusCode()
	respBody := body.String()

	var result struct {
		Status      string `json:"status"`
		LastMessage string `json:"last_message"`
	}
	if err := json.NewDecoder(&body).Decode(&result); err != nil {
		return apistructs.MigrationStatusDesc{}, respBodyDecodeErr(statusCode, respBody, err)
	}
	if result.Status == "" {
		return apistructs.MigrationStatusDesc{}, errors.Errorf("get empty status from scheduler, respBody: %s", respBody)
	}
	transferredStatus := transferStatus(result.Status)
	logrus.Infof("migration namespace: %s, name: %s, schedulerStatus: %s, transferredStatus: %s, lastMessage: %s",
		namespace, name, result.Status, transferredStatus, result.LastMessage)
	return apistructs.MigrationStatusDesc{
		Status: transferredStatus,
		Desc:   result.LastMessage,
	}, nil
}

func transferStatus(status string) apistructs.StatusCode {
	switch status {

	case string(apistructs.StatusError):
		return apistructs.StatusError

	case string(apistructs.StatusUnknown):
		return apistructs.StatusUnknown

	case string(apistructs.StatusCreated):
		return apistructs.StatusCreated

	case string(apistructs.StatusUnschedulable), "INITIAL":
		return apistructs.StatusUnschedulable

	case string(apistructs.StatusRunning), "ACTIVE":
		return apistructs.StatusRunning

	case string(apistructs.StatusStoppedOnOK), string(apistructs.StatusFinished):
		return apistructs.StatusStoppedOnOK

	case string(apistructs.StatusStoppedOnFailed), string(apistructs.StatusFailed):
		return apistructs.StatusStoppedOnFailed

	case string(apistructs.StatusStoppedByKilled):
		return apistructs.StatusStoppedByKilled

	case string(apistructs.StatusNotFoundInCluster):
		// scheduler 返回 job 在 cluster 中不存在 (在 etcd 中存在)，对应为 启动错误
		// 典型场景：created 成功，env key 为数字，导致 start job 时真正去创建 k8s job 时失败，即启动失败
		return apistructs.StatusNotFoundInCluster
	}

	return apistructs.StatusUnknown
}

// transferToSchedulerJob 转换为job
func (m *Migration) transferToSchedulerJob(migrationLog *dbclient.MigrationLog, diceyml *diceyml.DiceYaml, Runtime *dbclient.Runtime, App *apistructs.ApplicationDTO) (job apistructs.JobFromUser, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("%v", r)
		}
	}()

	// 获取schedule的namespace信息
	if migrationLog.ID <= 0 {
		return apistructs.JobFromUser{}, errors.Errorf("not found migration log id")
	}
	namespace, name := getNamespaceAndName(migrationLog)
	// env环境变量填充
	var config map[string]interface{}
	if err = json.Unmarshal([]byte(migrationLog.AddonInstanceConfig), &config); err != nil {
		return apistructs.JobFromUser{}, err
	}
	env := map[string]string{}
	for k, v := range config {
		switch t := v.(type) {
		case string:
			env[k] = t
		default:
			env[k] = fmt.Sprintf("%v", t)
		}
	}
	for k, v := range diceyml.Obj().Envs {
		env[k] = v
	}
	env["TERMINUS_DEFINE_TAG"] = "migration-task-" + strconv.FormatUint(migrationLog.ID, 10)
	return apistructs.JobFromUser{
		Name:        name,
		Kind:        "",
		Namespace:   namespace,
		ClusterName: Runtime.ClusterName,
		Image:       diceyml.Obj().Jobs["migration"].Image,
		Cmd:         "",
		CPU:         diceyml.Obj().Jobs["migration"].Resources.CPU,
		Memory:      float64(diceyml.Obj().Jobs["migration"].Resources.Mem),
		Env:         env,
		Labels: map[string]string{
			"DICE_PROJECT_ID":     strconv.FormatUint(migrationLog.ProjectID, 10),
			"DICE_WORKSPACE":      Runtime.Workspace,
			"DICE_ORG_NAME":       App.OrgName,
			"TERMINUS_DEFINE_TAG": "migration-task-" + strconv.FormatUint(migrationLog.ID, 10),
		},
	}, nil
}

func (m *Migration) CleanUnusedMigrationNs() (bool, error) {
	migrationLogs, err := m.db.GetMigrationLogExpiredThreeDays()
	if err != nil {
		return false, err
	}
	if len(*migrationLogs) == 0 {
		logrus.Info("no migration record found.")
		return false, err
	}
	for _, v := range *migrationLogs {
		namespace, _ := getNamespaceAndName(&v)
		if err := m.Remove(namespace, &v); err != nil {
			logrus.Errorf("remove migration record fail, namespace: %s, resp.Error: %s",
				namespace, err.Error())
			continue
		}
		v.Status = apistructs.MigrationStatusDeleted
		if err := m.db.UpdateMigrationLog(&v); err != nil {
			logrus.Errorf("update migration record status fail, namespace: %s, resp.Error: %s",
				namespace, err.Error())
		}
	}
	return false, nil
}

func (m *Migration) Remove(namespace string, migrationLog *dbclient.MigrationLog) (err error) {
	var body bytes.Buffer
	resp, err := httpclient.New().Delete(discover.Scheduler()).
		Path(fmt.Sprintf("/v1/job/%s/deletealljobs", namespace)).
		Do().Body(&body)
	if err != nil {
		return httpInvokeErr(err)
	}

	var result apistructs.JobDeleteResponse
	if err := json.NewDecoder(&body).Decode(&result); err != nil {
		return err
	}
	if result.Error != "" {
		if strings.Contains(result.Error, "not found") {
			logrus.Warnf("skip resp.Error(not found) when invoke scheduler.remove, namespace: %s, resp.Error: %s",
				namespace, result.Error)
			return nil
		}
		return errors.Errorf("statusCode: %d, resp.error: %s", resp.StatusCode(), result.Error)
	}
	return nil
}

// respBodyDecodeErr err response封装
func respBodyDecodeErr(statusCode int, respBody string, err error) error {
	return errors.Errorf("statusCode: %d, respBody: %s, err: %v", statusCode, respBody, err)
}

// httpInvokeErr http err封装
func httpInvokeErr(err error) error {
	return errors.Errorf("http invoke err: %v", err)
}

// 获取schedule namespace和name
func getNamespaceAndName(migrationLog *dbclient.MigrationLog) (string, string) {
	taskId := strconv.FormatUint(migrationLog.ID, 10)
	return "migration-" + taskId, "migration-task-" + taskId
}
