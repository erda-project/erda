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
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	v3 "github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	etcdDBGCWatchPrefix    = "/devops/pipeline/dbgc/pipeline/"
	etcdDBGCDLockKeyPrefix = "/devops/pipeline/dbgc/dlock/"
)

// remove ListenDatabaseGC and EnsureDatabaseGC these two methods，
// these two methods will create a lot of etcd ttl, will cause high load on etcd
// use fixed gc time, traverse the data in the database every day
func (r *Reconciler) PipelineDatabaseGC() {

	r.doAnalyzedPipelineDatabaseGC(true)
	r.doAnalyzedPipelineDatabaseGC(false)

	r.doNotAnalyzedPipelineDatabaseGC(true)
	r.doNotAnalyzedPipelineDatabaseGC(false)

	time.AfterFunc(24*time.Hour, func() {
		r.PipelineDatabaseGC()
	})
}

// query the data in the database according to req paging to perform gc
func (r *Reconciler) doPipelineDatabaseGC(req apistructs.PipelinePageListRequest) {
	var pageNum = req.PageNum

	for {
		req.PageNum = pageNum

		pipelineResults, _, _, _, err := r.dbClient.PageListPipelines(req)

		if err != nil {
			logrus.Errorf("doPipelineDatabaseGC failed to compensate pipeline req %v err: %v", req, err)
			return
		}

		if len(pipelineResults) <= 0 {
			return
		}

		for _, p := range pipelineResults {
			var needArchive = false
			if p.Status == apistructs.PipelineStatusAnalyzed {
				if p.Extra.GC.DatabaseGC.Analyzed.NeedArchive != nil {
					needArchive = *p.Extra.GC.DatabaseGC.Analyzed.NeedArchive
				}
			} else {
				if p.Extra.GC.DatabaseGC.Finished.NeedArchive != nil {
					needArchive = *p.Extra.GC.DatabaseGC.Finished.NeedArchive
				}
			}

			// gc logic
			if err := r.DoDBGC(p.PipelineID, apistructs.PipelineGCDBOption{NeedArchive: needArchive}); err != nil {
				logrus.Errorf("[alert] dbgc: failed to do gc logic, pipelineID: %d, err: %v", p.PipelineID, err)
				return
			}
		}

		pageNum += 1
		time.Sleep(time.Second * 10)
	}
}

// gc Analyzed status pipeline
func (r *Reconciler) doAnalyzedPipelineDatabaseGC(isSnippetPipeline bool) {
	var req apistructs.PipelinePageListRequest
	req.Statuses = []string{apistructs.PipelineStatusAnalyzed.String()}
	req.IncludeSnippet = isSnippetPipeline
	req.DescCols = []string{"id"}
	req.EndTimeCreated = time.Now().Add(-time.Second * time.Duration(conf.AnalyzedPipelineDefaultDatabaseGCTTLSec()))
	req.PageSize = 1000
	req.LargePageSize = true
	req.PageNum = 1
	req.AllSources = true

	r.doPipelineDatabaseGC(req)
}

// gc other status pipeline
func (r *Reconciler) doNotAnalyzedPipelineDatabaseGC(isSnippetPipeline bool) {
	var req apistructs.PipelinePageListRequest
	req.NotStatuses = []string{apistructs.PipelineStatusAnalyzed.String()}
	req.IncludeSnippet = isSnippetPipeline
	req.DescCols = []string{"id"}
	req.EndTimeCreated = time.Now().Add(-time.Second * time.Duration(conf.FinishedPipelineDefaultDatabaseGCTTLSec()))
	req.PageSize = 1000
	req.LargePageSize = true
	req.PageNum = 1
	req.AllSources = true

	r.doPipelineDatabaseGC(req)
}

// ListenDatabaseGC 监听需要 GC 的 pipeline database record.
func (r *Reconciler) ListenDatabaseGC() {
	logrus.Info("dbgc: start watching gc pipelines")
	for {
		ctx := context.Background()

		err := r.js.IncludeWatch().Watch(ctx, etcdDBGCWatchPrefix, true, false, false, apistructs.PipelineGCInfo{},
			func(key string, value interface{}, t storetypes.ChangeType) error {

				// async handle, non-blocking, so we can watch subsequent incoming pipelines
				go func() {

					logrus.Infof("dbgc: gc watched a key change, key: %s, changeType: %s", key, t.String())
					// only listen del op
					if t != storetypes.Del {
						return
					}

					pipelineID, parseErr := getPipelineIDFromDBGCWatchedKey(key)
					if parseErr != nil {
						logrus.Errorf("[alert] dbgc: invalid key: %s, failed to get pipelineID from key, err: %v", key, parseErr)
						return
					}

					// acquire a dlock
					gcDBLockKey := makeDBGCDLockKey(pipelineID)
					lock, err := r.etcd.GetClient().Txn(context.Background()).
						If(v3.Compare(v3.Version(gcDBLockKey), "=", 0)).
						Then(v3.OpPut(gcDBLockKey, "")).
						Commit()
					defer func() {
						_, _ = r.etcd.GetClient().Txn(context.Background()).Then(v3.OpDelete(gcDBLockKey)).Commit()
					}()
					if err != nil {
						return
					}
					if lock != nil && !lock.Succeeded {
						return
					}

					gcOption, err := getGCOptionFromValue(value.(*apistructs.PipelineGCInfo).Data)
					if err != nil {
						logrus.Errorf("[alert] dbgc: failed to get gc option from value, pipelineID: %d, err: %v", pipelineID, err)
						return
					}

					// gc logic
					if err := r.DoDBGC(pipelineID, gcOption); err != nil {
						logrus.Errorf("[alert] dbgc: failed to do gc logic, pipelineID: %d, err: %v", pipelineID, err)
						return
					}
				}()
				return nil
			},
		)
		if err != nil {
			logrus.Errorf("[alert] dbgc: gc watch failed, err: %v", err)
		}
	}
}

func (r *Reconciler) WaitDBGC(pipelineID uint64, ttl uint64, needArchive bool) {
	var err error
	defer func() {
		if err != nil {
			logrus.Errorf("[alert] dbgc: gc wait failed, pipelineID: %d, err: %v",
				pipelineID, err)
		} else {
			logrus.Debugf("dbgc: gc pipeline: %d in the future (%s) (TTL: %ds)",
				pipelineID, time.Now().Add(time.Duration(int64(time.Second)*int64(ttl))).Format(time.RFC3339), ttl)
		}
	}()

	// 设置 gc 等待时间
	lease := v3.NewLease(r.etcd.GetClient())
	grant, err := lease.Grant(context.Background(), int64(ttl))
	if err != nil {
		return
	}
	leaseID := strconv.FormatInt(int64(grant.ID), 16)
	logrus.Debugf("dbgc: grant lease, pipelineID: %d, leaseID: %s", pipelineID, leaseID)

	// 插入或更新 key
	gcOptionByte, err := generateGCOption(needArchive)
	if err != nil {
		return
	}
	gcInfo := apistructs.MakePipelineGCInfo(ttl, leaseID, gcOptionByte)
	_, err = r.js.PutWithOption(context.Background(),
		makeDBGCKey(pipelineID), gcInfo,
		[]interface{}{v3.WithLease(grant.ID)})
	if err != nil {
		return
	}
}

func (r *Reconciler) DoDBGC(pipelineID uint64, gcOption apistructs.PipelineGCDBOption) error {
	p, exist, err := r.dbClient.GetPipelineWithExistInfo(pipelineID)
	if err != nil {
		return err
	}
	if !exist {
		logrus.Infof("dbgc: no need to gc db, pipeline already not exist in db, pipelineID: %d", pipelineID)
		return nil
	}

	if gcOption.NeedArchive {
		// 归档
		_, err := r.dbClient.ArchivePipeline(p.ID)
		if err != nil {
			logrus.Errorf("[alert] dbgc: failed to archive pipeline, id: %d, err: %v", p.ID, err)
		}
		logrus.Debugf("dbgc: archive pipeline success, id: %d", p.ID)
	} else {
		// 删除
		if err := r.dbClient.DeletePipelineRelated(p.ID); err != nil {
			logrus.Errorf("[alert] dbgc: failed to delete pipeline, id: %d, err: %v", p.ID, err)
		}
		logrus.Debugf("dbgc: delete pipeline success, id: %d", p.ID)
	}
	return nil
}

// ex: /devops/pipeline/dbgc/pipeline/10000001
func makeDBGCKey(pipelineID uint64) string {
	return fmt.Sprintf("%s%d", etcdDBGCWatchPrefix, pipelineID)
}

// ex: /devops/pipeline/dbgc/dlock/10000001
func makeDBGCDLockKey(pipelineID uint64) string {
	return fmt.Sprintf("%s%d", etcdDBGCDLockKeyPrefix, pipelineID)
}

func generateGCOption(needArchive bool) ([]byte, error) {
	return json.Marshal(&apistructs.PipelineGCDBOption{NeedArchive: needArchive})
}

func getGCOptionFromValue(data []byte) (op apistructs.PipelineGCDBOption, err error) {
	err = json.Unmarshal(data, &op)
	return
}

// makeDBGCEnsureKey 生成用于 dbgc ensure 的分布式锁 key
func makeDBGCEnsureKey() string {
	return makeDBGCDLockKey(0)
}

// EnsureDatabaseGC etcd lease ttl reset 存在问题，因此要定期巡检，主动 delete 那些已经到了 gcAt 时间仍然存在的 etcd key 来触发 dbgc
// github issue: https://github.com/etcd-io/etcd/issues/9395
func (r *Reconciler) EnsureDatabaseGC() {
	logrus.Info("dbgc ensure: start ensure dbgc pipelines")

	// 防止多实例启动时同时申请布式锁，先等待随机时间
	rand.Seed(time.Now().UnixNano())
	randN := rand.Intn(60)
	logrus.Debugf("dbgc ensure: random sleep %d seconds...", randN)
	time.Sleep(time.Duration(randN) * time.Second)

	for {
		ctx := context.Background()

		done := make(chan struct{})
		errDone := make(chan error)

		go func() {

			// 先获取分布式锁
			ensureLockKey := makeDBGCEnsureKey()
			lock, err := r.etcd.GetClient().Txn(context.Background()).
				If(v3.Compare(v3.Version(ensureLockKey), "=", 0)).
				Then(v3.OpPut(ensureLockKey, "")).
				Commit()
			defer func() {
				_, _ = r.etcd.GetClient().Txn(context.Background()).Then(v3.OpDelete(ensureLockKey)).Commit()
			}()
			if err != nil {
				errDone <- fmt.Errorf("failed to get dlock: %s, err: %v", ensureLockKey, err)
				return
			}
			if lock != nil && !lock.Succeeded {
				done <- struct{}{}
				return
			}

			// 获取所有 key 列表
			keys, err := r.js.ListKeys(ctx, etcdDBGCWatchPrefix)
			if err != nil {
				errDone <- fmt.Errorf("failed to list dbgc keys, err: %v", err)
				return
			}

			// 没有 etcd dbgc 的老数据，生成 etcd dbgc key 后 delete key 走相同的逻辑（归档或删除），而不是直接删除
			if len(keys) > 0 {
				// 检查点，在检查点之前的 pipeline id，没有 etcd dbgc key
				checkPointKey := keys[0]
				r.handleOldNonDBGCPipelines(checkPointKey)
			}

			for _, key := range keys {
				var gcInfo apistructs.PipelineGCInfo
				if err := r.js.Get(ctx, key, &gcInfo); err != nil {
					logrus.Errorf("dbgc ensure: failed to get dbgc key: %s, continue, err: %v", key, err)
					continue
				}
				now := time.Now().Round(0)
				// already expired
				if gcInfo.GCAt.Before(now) {
					if err := r.js.Remove(ctx, key, nil); err != nil {
						logrus.Errorf("dbgc ensure: failed to delete already expired dbgc key: %s, continue, err: %v", key, err)
						continue
					}
					logrus.Infof("dbgc ensure: remove already expired key: %s, gcAt: %s", key, gcInfo.GCAt.Format(time.RFC3339))
				}
			}

			done <- struct{}{}
		}()

		select {
		// 正常结束
		case <-done:
			// 完成本次 ensure 后等待 2h 开始下一次处理
			logrus.Infof("dbgc ensure: sleep 2 hours for next ensure...")
			time.Sleep(time.Hour * 2)

		// 异常结束
		case err := <-errDone:
			logrus.Errorf("dbgc ensure: failed to ensure, wait 5 mins for next ensure, err: %v", err)
			time.Sleep(time.Minute * 5)
		}
	}
}

func getPipelineIDFromDBGCWatchedKey(key string) (uint64, error) {
	s := strutil.TrimPrefixes(key, etcdDBGCWatchPrefix)
	id, err := strconv.ParseUint(s, 10, 64)
	if err == nil {
		return id, nil
	}
	return 0, fmt.Errorf("invalid key: %s", key)
}

func (r *Reconciler) handleOldNonDBGCPipelines(checkPointDBGCKey string) {
	checkPointPID, err := getPipelineIDFromDBGCWatchedKey(checkPointDBGCKey)
	if err != nil {
		logrus.Errorf("dbgc ensure: failed to get check point pid for handle old non-dbgc pipelines, err: %v", err)
		return
	}
	// fetch only id and status is enough
	var oldPipelineBases []spec.PipelineBase
	if err := r.dbClient.Cols(`id`, `status`).Where("id < ?", checkPointPID).Find(&oldPipelineBases); err != nil {
		logrus.Errorf("dbgc ensure: failed to query pipelines before check point, err: %v", err)
		return
	}
	// transfer to pipeline for default ensureGC options
	var oldPipelines []spec.Pipeline
	for _, oldBase := range oldPipelineBases {
		oldPipelines = append(oldPipelines, spec.Pipeline{
			PipelineBase: oldBase,
			PipelineExtra: spec.PipelineExtra{
				Extra: spec.PipelineExtraInfo{
					GC: apistructs.PipelineGC{},
				},
			},
		})
	}
	for _, p := range oldPipelines {
		// get default gc options
		p.EnsureGC()
		// already expired, so ttl 5s is enough
		var ttl uint64 = 5
		// archive according to status
		needArchive := false
		if p.Status.IsEndStatus() {
			needArchive = true
		}
		// put into etcd dbgc
		logrus.Infof("dbgc ensure: put old non-dbgc pipeline with etcd dbgc key, id: %d, needArchive: %t", p.ID, needArchive)
		r.WaitDBGC(p.ID, ttl, needArchive)
	}
}
