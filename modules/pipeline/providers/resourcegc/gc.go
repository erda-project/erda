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

package resourcegc

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	v3 "github.com/coreos/etcd/clientv3"

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

func (p *provider) listenGC(ctx context.Context) {
	p.infof("start watching gc pipelines")
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// watch: isPrefix[true], filterDelete[false], keyOnly[true]
			err := p.js.IncludeWatch().Watch(ctx, etcdReconcilerGCWatchPrefix, true, false, true, nil,
				func(key string, _ interface{}, t storetypes.ChangeType) error {

					// async handle, non-blocking, so we can watch subsequent incoming pipelines
					go func() {
						p.infof("gc watched a key change, key: %s, changeType: %s", key, t.String())
						// only listen del op
						if t != storetypes.Del {
							return
						}

						// {ns}
						namespace := getPipelineNamespaceFromGCWatchedKey(key)

						var err error
						defer func() {
							if err != nil {
								p.errorf("gc handle failed, key: %s, changeType: %s, err: %v",
									key, t.String(), err)

								// put to gc again, retry gc in 60s
								p.WaitGC(namespace, 0, 60)
							}
						}()

						// new data ignore namespace's subKey
						//
						// new data:
						// {prefix}/{ns}
						// {prefix}/{ns}/{pipelineID-1}
						// {prefix}/{ns}/{pipelineID-2}
						var oldFormatSubKeys []string
						if strutil.Contains(namespace, "/") {
							// list key，如果没找到 key: {prefix}/{ns}，则为老格式
							namespace = strutil.Split(namespace, "/")[0]
							keys, err := p.js.ListKeys(ctx, makePipelineGCKey(namespace))
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
						lock, err := p.etcd.GetClient().Txn(context.Background()).
							If(v3.Compare(v3.Version(gcNamespaceLockKey), "=", 0)).
							Then(v3.OpPut(gcNamespaceLockKey, "")).
							Commit()
						defer func() {
							_, _ = p.etcd.GetClient().Txn(context.Background()).Then(v3.OpDelete(gcNamespaceLockKey)).Commit()
						}()
						if err != nil {
							return
						}
						if lock != nil && !lock.Succeeded {
							return
						}

						// gc logic
						if err = p.gcNamespace(namespace, oldFormatSubKeys...); err != nil {
							return
						}
					}()
					return nil
				},
			)
			if err != nil {
				p.errorf("gc watch failed, err: %v", err)
			}
		}
	}
}

// DelayGC delay GC
// 1) if not wait gc，return directly
// 2) if waiting gc，lease set a longer time (two day)，prevent GC
//
// example：retry failed-task，namespace unchanging，if namespace be gc，will cause task gc, pipeline failed
func (p *provider) DelayGC(namespace string, pipelineID uint64) {
	notFound, err := p.js.Notfound(context.Background(), makePipelineGCKey(namespace))
	if err != nil {
		p.errorf("delay gc failed, namespace: %s, cause pipelineID: %d, err: %v",
			namespace, pipelineID, err)
		return
	}
	if notFound {
		return
	}
	p.infof("delay gc begin, namespace: %s, cause pipelineID: %d", namespace, pipelineID)
	p.WaitGC(namespace, pipelineID, defaultGCTime)
}

// WaitGC after ttl, gc namespace
func (p *provider) WaitGC(namespace string, pipelineID uint64, ttl uint64) {
	var err error
	defer func() {
		if err != nil {
			p.errorf("gc wait failed, namespace: %s, pipelineID: %d, err: %v",
				namespace, pipelineID, err)
		} else {
			p.debugf("gc namespace: %s in the future (%s) (TTL: %ds)",
				namespace, time.Now().Add(time.Duration(int64(time.Second)*int64(ttl))).Format(time.RFC3339), ttl)
		}
	}()

	// set gc wait time
	lease := v3.NewLease(p.etcd.GetClient())
	grant, err := lease.Grant(context.Background(), int64(ttl))
	if err != nil {
		return
	}
	// Lease in etcd is hexadecimal
	leaseID := strconv.FormatInt(int64(grant.ID), 16)
	p.debugf("gc grant lease, pipelineID: %d, leaseID: %s", pipelineID, leaseID)

	// insert or update namespace key
	gcInfo := apistructs.MakePipelineGCInfo(ttl, leaseID, nil)
	_, err = p.js.PutWithOption(context.Background(),
		makePipelineGCKey(namespace), gcInfo,
		[]interface{}{v3.WithLease(grant.ID)})
	if err != nil {
		return
	}

	// insert subKey
	if pipelineID > 0 {
		if err := p.js.Put(context.Background(), makePipelineGCSubKey(namespace, pipelineID), nil); err != nil {
			return
		}
	}
}

// gcNamespace
//
// namespace struct stored in etcd:
// /devops/pipeline/gc/reconciler/{ns}/{affectedPipelineID}
//
// /devops/pipeline/gc/reconciler/pipeline-1
// /devops/pipeline/gc/reconciler/pipeline-1/1
// /devops/pipeline/gc/reconciler/pipeline-1/2 # retry failed with same ns
func (p *provider) gcNamespace(namespace string, subKeys ...string) error {

	// 1) 遍历 pipelineID 标记为 gc completed
	// 2) executor.DeleteNamespace
	// 3) 根据 prefix 删除 key

	gcPrefixKey := makePipelineGCKeyWithSlash(namespace)
	if len(subKeys) == 0 {
		eSubKeys, err := p.js.ListKeys(context.Background(), gcPrefixKey)
		if err != nil {
			return err
		}
		subKeys = eSubKeys
	}
	p.infof("gc triggered, namespace: %s, subKeys: %s",
		namespace, strutil.Join(subKeys, ", ", true))

	affectedPipelineIDs := make([]uint64, 0, len(subKeys))

	for _, subKey := range subKeys {
		pipelineIDStr := strutil.TrimPrefixes(subKey, gcPrefixKey)
		pipelineID, err := strconv.ParseUint(pipelineIDStr, 10, 64)
		if err != nil {
			return err
		}
		// clean up pipelines that have been archived
		pt, found, findFromArchive, err := p.DBGC.GetPipelineIncludeArchived(context.Background(), pipelineID)
		if !found {
			p.errorf("gc triggered but ignored, pipeline already not exists, pipelineID: %d", pipelineID)
		}
		if !findFromArchive {
			// mark it as GC complete first, then do the cleanup.
			// even if the cleanup fails, this pipeline is guaranteed not to be used again (retry).
			pt.Extra.CompleteReconcilerGC = true
			if err := p.dbClient.UpdatePipelineExtraExtraInfoByPipelineID(pipelineID, pt.Extra); err != nil {
				return err
			}
		}
		affectedPipelineIDs = append(affectedPipelineIDs, pipelineID)
	}

	// group tasks by executorName
	groupedTasks := make(map[spec.PipelineTaskExecutorName][]*spec.PipelineTask) // key: executorName
	for _, affectedPipelineID := range affectedPipelineIDs {
		dbTasks, _, err := p.DBGC.GetPipelineTasksIncludeArchived(context.Background(), affectedPipelineID)
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
				executorName := loopTask.GetExecutorName()
				groupedTasks[executorName] = append(groupedTasks[executorName], &loopTask)
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

	if _, err := p.js.PrefixRemove(context.Background(), gcPrefixKey); err != nil {
		return err
	}

	p.infof("gc success, namespace: %s, affected pipelineIDs: %v", namespace, affectedPipelineIDs)

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
