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

package crondsvc

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	etcdCronWatchPrefix     = "/devops/pipeline/cron/"
	etcdCronPrefixDeleteKey = etcdCronWatchPrefix + "delete-"
	etcdCronPrefixAddKey    = etcdCronWatchPrefix + "add-"
)

// Listen the key in the channel
// decide to delete or add a scheduled task according to the prefix of the key
func (s *CrondSvc) ListenCrond(ctx context.Context, pipelineCronFunc func(id uint64)) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_ = s.js.IncludeWatch().Watch(ctx, etcdCronWatchPrefix, true, true, true, nil, func(key string, _ interface{}, t storetypes.ChangeType) (_ error) {
				go func() {
					logrus.Infof("crond: watched update etcd key: %s, changeType: %d", key, t)

					// if change type is delete, don't need do anything
					if t == storetypes.Del {
						return
					}

					if err := s.js.Remove(ctx, key, nil); err != nil {
						logrus.Errorf("crond: failed to delete key: %s", key)
						return
					}
					cronID, err := parseCronIDFromWatchedKey(key)
					if err != nil {
						logrus.Errorf("crond: failed to parse cronID from watched key, key: %s, err: %v", key, err)
						return
					}

					if strings.HasPrefix(key, etcdCronPrefixAddKey) {
						pc, err := s.dbClient.GetPipelineCron(cronID)
						if err != nil {
							logrus.Errorf("crond: failed to get cron cronID: %v error: %v", cronID, err)
							return
						}
						// why delete it first, because crond.AddFunc cannot add a scheduled task with the same name
						err = s.crond.Remove(makePipelineCronName(cronID))
						if err != nil {
							logrus.Errorf("crond: failed to remove cron cronID: %v error: %v", cronID, err)
							return
						}

						// determine whether there is a scheduled task
						if pc.Enable != nil && *pc.Enable && pc.CronExpr != "" {
							err = s.crond.AddFunc(pc.CronExpr, func() {
								pipelineCronFunc(pc.ID)
							}, makePipelineCronName(pc.ID))
							if err != nil {
								logrus.Errorf("crond: failed to update cron cronID: %v cronExpr: %v  error: %v", cronID, pc.CronExpr, err)
								return
							}
						}
					} else if strings.HasPrefix(key, etcdCronPrefixDeleteKey) {
						err = s.crond.Remove(makePipelineCronName(cronID))
						if err != nil {
							logrus.Errorf("crond: failed to remove cron cronID: %v error: %v", cronID, err)
							return
						}
					}
					logrus.Infof("crond: watched and reload successfully")
				}()
				return nil
			})
		}
	}
}

// Remove the prefix of the key and get the cronid
func parseCronIDFromWatchedKey(key string) (uint64, error) {
	pipelineIDStr := strutil.TrimPrefixes(key, etcdCronPrefixDeleteKey)
	pipelineIDStr = strutil.TrimPrefixes(pipelineIDStr, etcdCronPrefixAddKey)
	return strconv.ParseUint(pipelineIDStr, 10, 64)
}

// delete the timing task of a certain timing id
// todo Multi-instance watch etcd delete key
func (s *CrondSvc) DeletePipelineCrond(cronID uint64) error {
	if cronID <= 0 {
		return nil
	}
	return s.js.Put(context.Background(), etcdCronPrefixDeleteKey+strconv.FormatUint(cronID, 10), nil)
}

// add the timing task of a certain timing id
// todo Multi-instance watch etcd add key
func (s *CrondSvc) AddIntoPipelineCrond(cronID uint64) error {
	if cronID <= 0 {
		return nil
	}
	return s.js.Put(context.Background(), etcdCronPrefixAddKey+strconv.FormatUint(cronID, 10), nil)
}

// ReloadCrond 触发当前 crond 实例更新任务.
func (s *CrondSvc) ReloadCrond(ctx context.Context, pipelineCronFunc func(uint64)) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var logs []string
	logs = append(logs, "reloading crond from db `pipeline_crons` table ...")

	pcs, err := s.dbClient.ListAllPipelineCrons()
	if err != nil {
		return logs, err
	}

	// 关闭 crond，停止定时任务和etcd连接
	if s.crond != nil {
		s.crond.Close()
		s.crond = nil
	}

	// 初始化一个新的 crond
	s.crond = cron.New(cron.WithoutDLock(true))
	s.crond.Start()
	go func() {
		select {
		case <-ctx.Done():
			s.crond.Stop()
			return
		default:

		}
	}()

	for i := range pcs {
		pc := pcs[i]
		//todo 校验pc.CronExpr是否合法
		if pc.Enable != nil && *pc.Enable && pc.CronExpr != "" {
			if err = s.crond.AddFunc(pc.CronExpr, func() { pipelineCronFunc(pc.ID) }, makePipelineCronName(pc.ID)); err != nil {
				l := fmt.Sprintf("failed to load pipeline cron item: %s, cronExpr: %v, err: %v", makePipelineCronName(pc.ID), pc.CronExpr, err)
				logs = append(logs, l)
				logrus.Errorln("[alert]", l)
				continue
			}
			logs = append(logs, fmt.Sprintf("loaded pipeline cron item: %s, cronExpr: %v", makePipelineCronName(pc.ID), pc.CronExpr))
		}
	}

	// clean build cache cron task
	buildCacheCleanJobName := makeCleanBuildCacheJobName(conf.BuildCacheCleanJobCron())
	if err = s.crond.AddFunc(conf.BuildCacheCleanJobCron(), s.CleanBuildCacheImages, buildCacheCleanJobName); err != nil {
		l := fmt.Sprintf("failed to load build cache clean cron task: %s, err: %v", buildCacheCleanJobName, err)
		logs = append(logs, l)
		logrus.Errorln("[alert]", l)
	} else {
		logs = append(logs, fmt.Sprintf("loaded build cache clean cron task: %s", buildCacheCleanJobName))
	}

	logs = append(logs, "reload crond DONE")
	logs = append(logs, s.crondSnapshot()...)

	return logs, nil
}

func (s *CrondSvc) CrondSnapshot() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.crondSnapshot()
}

func (s *CrondSvc) crondSnapshot() []string {
	var logs []string
	logs = append(logs, "inspecting cron daemon ...")
	for _, entry := range s.crond.Entries() {
		logs = append(logs, fmt.Sprintf("cron daemon task: %s", entry.Name))
	}
	logs = append(logs, "inspect cron daemon DONE")
	return logs
}

func makePipelineCronName(cronID uint64) string {
	return fmt.Sprintf("pipeline-cron[%v]", cronID)
}

func makeCleanBuildCacheJobName(cronExpr string) string {
	return fmt.Sprintf("clean-build-cache-image-[%s]", cronExpr)
}
