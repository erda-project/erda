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

package reconciler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/pkg/task_uuid"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	etcdReconcilerGCWatchPrefix            = "/devops/pipeline/gc/reconciler/"
	etcdReconcilerGCNamespaceLockKeyPrefix = "/devops/pipeline/gc/dlock/"
	defaultGCTime                          = 3600 * 24 * 2
)

// ListenGC 监听需要 GC 的 pipeline.
func (r *Reconciler) ListenGC() {
	logrus.Info("reconciler: start watching gc pipelines")
	for {
		ctx := context.Background()

		// watch: isPrefix[true], filterDelete[false], keyOnly[true]
		err := r.js.IncludeWatch().Watch(ctx, etcdReconcilerGCWatchPrefix, true, false, true, nil,
			func(key string, _ interface{}, t storetypes.ChangeType) error {

				// async handle, non-blocking, so we can watch subsequent incoming pipelines
				go func() {
					logrus.Infof("reconciler: gc watched a key change, key: %s, changeType: %s", key, t.String())
					// only listen del op
					if t != storetypes.Del {
						return
					}

					// {ns}
					namespace := getPipelineNamespaceFromGCWatchedKey(key)

					var err error
					defer func() {
						if err != nil {
							logrus.Errorf("[alert] reconciler: gc handle failed, key: %s, changeType: %s, err: %v",
								key, t.String(), err)

							// put to gc again, retry gc in 60s
							r.waitGC(namespace, 0, 60)
						}
					}()

					// 新数据 忽略 namespace 下的 subKey
					//
					// 新数据格式为多条:
					// {prefix}/{ns}
					// {prefix}/{ns}/{pipelineID-1}
					// {prefix}/{ns}/{pipelineID-2}
					var oldFormatSubKeys []string
					if strutil.Contains(namespace, "/") {
						// list key，如果没找到 key: {prefix}/{ns}，则为老格式
						namespace = strutil.Split(namespace, "/")[0]
						keys, err := r.js.ListKeys(ctx, makePipelineGCKey(namespace))
						if err != nil {
							return
						}
						isOldFormat := true
						for _, key := range keys {
							if key == makePipelineGCKey(namespace) {
								isOldFormat = false
								break
							}
						}

						if !isOldFormat {
							return
						}
						oldFormatSubKeys = append(oldFormatSubKeys, key)
					}

					// acquire a dlock
					gcNamespaceLockKey := etcdReconcilerGCNamespaceLockKeyPrefix + namespace
					lock, err := r.etcd.GetClient().Txn(context.Background()).
						If(v3.Compare(v3.Version(gcNamespaceLockKey), "=", 0)).
						Then(v3.OpPut(gcNamespaceLockKey, "")).
						Commit()
					defer func() {
						_, _ = r.etcd.GetClient().Txn(context.Background()).Then(v3.OpDelete(gcNamespaceLockKey)).Commit()
					}()
					if err != nil {
						return
					}
					if lock != nil && !lock.Succeeded {
						return
					}

					// gc logic
					if err = r.gcNamespace(namespace, oldFormatSubKeys...); err != nil {
						return
					}
				}()
				return nil
			},
		)
		if err != nil {
			logrus.Errorf("[alert] reconciler: gc watch failed, err: %v", err)
		}
	}
}

// delayGC 延迟 GC
// 1) 若暂未 wait gc，则直接返回
// 2) 若正在 wait gc，则 lease 设置一个较大的值 (两天)，防止被 GC
//
// 典型场景：重试失败节点时，namespace 不变，若失败 pipeline 被 GC，会导致新的任务也被 GC，重试失败
func (r *Reconciler) delayGC(namespace string, pipelineID uint64) {
	notFound, err := r.js.Notfound(context.Background(), makePipelineGCKey(namespace))
	if err != nil {
		logrus.Errorf("[alert] reconciler: delay gc failed, namespace: %s, cause pipelineID: %d, err: %v",
			namespace, pipelineID, err)
		return
	}
	if notFound {
		return
	}
	logrus.Errorf("reconciler: delay gc begin, namespace: %s, cause pipelineID: %d", namespace, pipelineID)
	r.waitGC(namespace, pipelineID, defaultGCTime)
}

// waitGC 等待 GC，在 TTL 后执行
func (r *Reconciler) waitGC(namespace string, pipelineID uint64, ttl uint64) {
	var err error
	defer func() {
		if err != nil {
			logrus.Errorf("[alert] reconciler: gc wait failed, namespace: %s, pipelineID: %d, err: %v",
				namespace, pipelineID, err)
		} else {
			logrus.Debugf("reconciler: gc namespace: %s in the future (%s) (TTL: %ds)",
				namespace, time.Now().Add(time.Duration(int64(time.Second)*int64(ttl))).Format(time.RFC3339), ttl)
		}
	}()

	// 设置 gc 等待时间
	lease := v3.NewLease(r.etcd.GetClient())
	grant, err := lease.Grant(context.Background(), int64(ttl))
	if err != nil {
		return
	}
	// etcd 中 lease 为 16 进制
	leaseID := strconv.FormatInt(int64(grant.ID), 16)
	logrus.Debugf("reconciler: gc grant lease, pipelineID: %d, leaseID: %s", pipelineID, leaseID)

	// 插入或更新 namespace key
	gcInfo := apistructs.MakePipelineGCInfo(ttl, leaseID, nil)
	_, err = r.js.PutWithOption(context.Background(),
		makePipelineGCKey(namespace), gcInfo,
		[]interface{}{v3.WithLease(grant.ID)})
	if err != nil {
		return
	}

	// 插入 subKey
	if pipelineID > 0 {
		if err := r.js.Put(context.Background(), makePipelineGCSubKey(namespace, pipelineID), nil); err != nil {
			return
		}
	}
}

// gcNamespace
//
// etcd 中 namespace 数据结构:
// /devops/pipeline/gc/reconciler/{ns}/{affectedPipelineID}
//
// /devops/pipeline/gc/reconciler/pipeline-1
// /devops/pipeline/gc/reconciler/pipeline-1/1
// /devops/pipeline/gc/reconciler/pipeline-1/2 # 重试失败节点，共用一个 ns
func (r *Reconciler) gcNamespace(namespace string, subKeys ...string) error {

	// 1) 遍历 pipelineID 标记为 gc completed
	// 2) executor.DeleteNamespace
	// 3) 根据 prefix 删除 key

	gcPrefixKey := makePipelineGCKeyWithSlash(namespace)
	if len(subKeys) == 0 {
		eSubKeys, err := r.js.ListKeys(context.Background(), gcPrefixKey)
		if err != nil {
			return err
		}
		subKeys = eSubKeys
	}
	logrus.Infof("reconciler: gc triggered, namespace: %s, subKeys: %s",
		namespace, strutil.Join(subKeys, ", ", true))

	affectedPipelineIDs := make([]uint64, 0, len(subKeys))

	for _, subKey := range subKeys {
		pipelineIDStr := strutil.TrimPrefixes(subKey, gcPrefixKey)
		pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
		if err != nil {
			return err
		}
		// 为了清理已经被归档的流水线
		p, found, findFromArchive, err := r.dbClient.GetPipelineIncludeArchive(pipelineID)
		if !found {
			logrus.Errorf("[alert] reconciler: gc triggered but ignored, pipeline already not exists, pipelineID: %d", pipelineID)
		}
		if !findFromArchive {
			// 先标记为 GC 完成，再做清理。即使清理失败，也保证这条 pipeline 不会再被使用(重试).
			p.Extra.CompleteReconcilerGC = true
			if err := r.dbClient.UpdatePipelineExtraExtraInfoByPipelineID(pipelineID, p.Extra); err != nil {
				return err
			}
		}
		affectedPipelineIDs = append(affectedPipelineIDs, pipelineID)
	}

	// group tasks by executorName
	groupedTasks := make(map[spec.PipelineTaskExecutorName][]*spec.PipelineTask) // key: executorName
	for _, affectedPipelineID := range affectedPipelineIDs {
		dbTasks, _, err := r.dbClient.GetPipelineTasksIncludeArchive(affectedPipelineID)
		if err != nil {
			return err
		}
		for i := range dbTasks {
			task := dbTasks[i]
			// snippet task has no entity to delete
			if task.IsSnippet {
				continue
			}
			// no executor info
			if task.ExecutorKind == "" || task.Extra.ExecutorName == "" {
				continue
			}
			// not begin reconcile prepare
			if task.Extra.UUID == "" {
				continue
			}

			for _, uuid := range task_uuid.MakeJobIDSliceWithLoopedTimes(&task) {
				var loopTask = task
				loopTask.Extra.UUID = uuid
				groupedTasks[loopTask.Extra.ExecutorName] = append(groupedTasks[loopTask.Extra.ExecutorName], &loopTask)
			}
		}
	}

	// iterate groupedTasks by executorName and batchDelete tasks
	var batchDeleteErrs []string
	for executorName, tasks := range groupedTasks {
		executor, err := actionexecutor.GetManager().Get(types.Name(executorName))
		if err != nil {
			batchDeleteErrs = append(batchDeleteErrs, err.Error())
			continue
		}
		if _, err := executor.BatchDelete(context.Background(), tasks); err != nil {
			batchDeleteErrs = append(batchDeleteErrs, err.Error())
			continue
		}
	}
	if len(batchDeleteErrs) > 0 {
		return fmt.Errorf("failed to gc namespace: %s, errs: %s", namespace, strings.Join(batchDeleteErrs, ", "))
	}

	if _, err := r.js.PrefixRemove(context.Background(), gcPrefixKey); err != nil {
		return err
	}

	logrus.Infof("reconciler: gc success, namespace: %s, affected pipelineIDs: %v", namespace, affectedPipelineIDs)

	return nil
}

func getPipelineNamespaceFromGCWatchedKey(key string) string {
	return strutil.TrimPrefixes(key, etcdReconcilerGCWatchPrefix)
}

// ex: /devops/pipeline/gc/reconciler/pipeline-1
func makePipelineGCKey(namespace string) string {
	return strutil.Concat(etcdReconcilerGCWatchPrefix, namespace)
}

// 加上 `/`，防止 prefixGet 到别的 namespace 导致误删，例如: PrefixGet(pipeline-1) => pipeline-12, pipeline-13
// ex: /devops/pipeline/gc/reconciler/pipeline-1/
func makePipelineGCKeyWithSlash(namespace string) string {
	return strutil.Concat(makePipelineGCKey(namespace), "/")
}

func makePipelineGCSubKey(namespace string, pipelineID uint64) string {
	return fmt.Sprintf("%s/%d", makePipelineGCKey(namespace), pipelineID)
}
