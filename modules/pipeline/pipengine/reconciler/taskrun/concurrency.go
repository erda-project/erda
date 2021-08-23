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

package taskrun

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/jsonstore"
	"github.com/erda-project/erda/pkg/loop"
)

const (
	etcdTaskConcurrencyKeyPrefix     = "/devops/pipeline/task-concurrency/v1"
	etcdTaskConcurrencyLockKeyPrefix = "/devops/pipeline/dlock/task-concurrency/v1"
)

// {prefix}/{actionType}/cluster/{cluster-name}/count
func (tr *TaskRun) makeTaskConcurrencyCountKey() string {
	return fmt.Sprintf("%s/%s/cluster/%s/count", etcdTaskConcurrencyKeyPrefix, tr.Task.Type, tr.P.ClusterName)
}

// {prefix}/{actionType}/cluster/{cluster-name}
func (tr *TaskRun) makeTaskConcurrencyCountLockKey() string {
	return fmt.Sprintf("%s/%s/cluster/%s", etcdTaskConcurrencyLockKeyPrefix, tr.Task.Type, tr.P.ClusterName)
}

// CalibrateConcurrencyCountFromDB 从数据库校准并发度
func (tr *TaskRun) CalibrateConcurrencyCountFromDB() {
	var tasks []spec.PipelineTask
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
		_tasks, err := tr.DBClient.ListPipelineTasksByTypeStatuses(tr.Task.Type, apistructs.PipelineStatusQueue, apistructs.PipelineStatusRunning)
		if err != nil {
			return false, err
		}
		// filter by clusterName
		for _, t := range _tasks {
			if tr.Task.Extra.ClusterName != t.Extra.ClusterName {
				continue
			}
			tasks = append(tasks, t)
		}
		return true, nil
	})
	calibratedValue := len(tasks)
	tr.AddTaskConcurrencyCount(calibratedValue, true)
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "task concurrency: calibrate from db success, actionType: %s, count: %d", tr.Task.Type, calibratedValue)
}

func (tr *TaskRun) GetTaskConcurrencyCount() int {
	ctx := context.Background()
	countKey := tr.makeTaskConcurrencyCountKey()
	var count int
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
		if err := tr.Js.Get(ctx, countKey, &count); err != nil {
			if err == jsonstore.NotFoundErr {
				count = 0
				return true, nil
			}
			return false, err
		}
		return true, nil
	})
	return count
}

func (tr *TaskRun) AddTaskConcurrencyCount(add int, overwriteOpt ...bool) {
	overwrite := false
	if len(overwriteOpt) > 0 {
		overwrite = overwriteOpt[0]
	}
	//// 操作前先获取分布式锁
	//var l *dlock.DLock
	//defer func() {
	//	if l != nil {
	//		_ = l.UnlockAndClose()
	//	}
	//}()
	ctx := context.Background()
	//_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
	//	lockKey := tr.makeTaskConcurrencyCountLockKey()
	//	_l, err := dlock.New(lockKey, nil)
	//	if err != nil {
	//		return false, err
	//	}
	//	l = _l
	//	if err := l.Lock(ctx); err != nil {
	//		return false, err
	//	}
	//	return true, nil
	//})
	var newCount int
	if overwrite {
		newCount = add
	} else {
		// 查询最新值
		current := tr.GetTaskConcurrencyCount()
		// 写入最新值
		newCount = current + add
	}
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
		if err := tr.Js.Put(ctx, tr.makeTaskConcurrencyCountKey(), &newCount); err != nil {
			return false, err
		}
		return true, nil
	})
}

func (tr *TaskRun) GetActionSpec() apistructs.ActionSpec {
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "begin get action spec")
	var actionSpec apistructs.ActionSpec
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
		_, actionSpecYmlJobMap, err := tr.ExtMarketSvc.SearchActions([]string{extmarketsvc.MakeActionTypeVersion(&tr.Task.Extra.Action)})
		if err != nil {
			return false, err
		}
		_spec, ok := actionSpecYmlJobMap[extmarketsvc.MakeActionTypeVersion(&tr.Task.Extra.Action)]
		if !ok {
			rlog.TErrorf(tr.P.ID, tr.Task.ID, "not found action spec, actionType: %s", tr.Task.Type)
			return false, fmt.Errorf("err for decline ratio")
		}
		actionSpec = *_spec
		return true, nil
	})
	b, _ := json.Marshal(actionSpec)
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "end get action spec, %s", string(b))
	return actionSpec
}

// -1 代表无限制
const NoConcurrencyLimit = -1

func (tr *TaskRun) GetConcurrencyLimit(concurrency *apistructs.ActionConcurrency) int {
	// 无限制
	if noConcurrencyLimit(concurrency) {
		return NoConcurrencyLimit
	}

	// 不同版本的处理逻辑不同
	// only v1 now
	if v1 := concurrency.V1; v1 != nil {
		return tr.getConcurrencyLimitV1(*v1)
	}

	// 默认无限制
	return NoConcurrencyLimit
}

func noConcurrencyLimit(concurrency *apistructs.ActionConcurrency) bool {
	if concurrency == nil {
		return true
	}
	if !concurrency.Enable {
		return true
	}
	if concurrency.V1 == nil {
		return true
	}
	return false
}

func (tr *TaskRun) getConcurrencyLimitV1(v1 apistructs.ActionConcurrencyV1) int {
	// 先使用对应集群的配置
	if v1.Clusters != nil {
		config, ok := v1.Clusters[tr.P.ClusterName]
		if ok {
			return config.Max
		}
	}

	// 无对应集群配置，使用默认配置
	return v1.Default.Max
}
