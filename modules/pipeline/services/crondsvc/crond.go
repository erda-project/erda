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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/strutil"
)

const cronChanPrefixDeleteKey = "delete-"
const cronChanPrefixAddKey = "add-"

// Listen the key in the channel
// decide to delete or add a scheduled task according to the prefix of the key
func (s *CrondSvc) ListenCrond(pipelineCronFunc func(id uint64)) {
	for {
		for key := range s.cronChan {
			logrus.Infof("crond: watched update chan key: %s", key)

			cronID, err := parseCronIDFromWatchedKey(key)
			if err != nil {
				logrus.Errorf("crond: failed to parse cronID from watched key, key: %s, err: %v", key, err)
				continue
			}

			if strings.HasPrefix(key, cronChanPrefixAddKey) {
				pc, err := s.dbClient.GetPipelineCron(cronID)
				if err != nil {
					logrus.Errorf("crond: failed to get cron cronID: %v error: %v", cronID, err)
					continue
				}
				// why delete it first, because crond.AddFunc cannot add a scheduled task with the same name
				err = s.crond.Remove(makePipelineCronName(cronID))
				if err != nil {
					logrus.Errorf("crond: failed to remove cron cronID: %v error: %v", cronID, err)
					continue
				}

				// determine whether there is a scheduled task
				if pc.Enable != nil && *pc.Enable && pc.CronExpr != "" {
					err = s.crond.AddFunc(pc.CronExpr, func() {
						pipelineCronFunc(pc.ID)
					}, makePipelineCronName(pc.ID))
					if err != nil {
						logrus.Errorf("crond: failed to update cron cronID: %v cronExpr: %v  error: %v", cronID, pc.CronExpr, err)
						continue
					}
				}
			} else if strings.HasPrefix(key, cronChanPrefixDeleteKey) {
				err = s.crond.Remove(makePipelineCronName(cronID))
				if err != nil {
					logrus.Errorf("crond: failed to remove cron cronID: %v error: %v", cronID, err)
					continue
				}
			}
			logrus.Infof("crond: watched and reload successfully")
		}
		time.Sleep(10 * time.Second)
	}
}

// Remove the prefix of the key and get the cronid
func parseCronIDFromWatchedKey(key string) (uint64, error) {
	pipelineIDStr := strutil.TrimPrefixes(key, cronChanPrefixDeleteKey)
	pipelineIDStr = strutil.TrimPrefixes(pipelineIDStr, cronChanPrefixAddKey)
	return strconv.ParseUint(pipelineIDStr, 10, 64)
}

// delete the timing task of a certain timing id
// todo Multi-instance watch etcd delete key
func (s *CrondSvc) DeletePipelineCrond(cronID uint64) error {
	if s.cronChan == nil {
		return fmt.Errorf("delete cron error: cron chan was empty")
	}
	if cronID <= 0 {
		return nil
	}
	s.cronChan <- cronChanPrefixDeleteKey + strconv.FormatUint(cronID, 10)
	return nil
}

// add the timing task of a certain timing id
// todo Multi-instance watch etcd add key
func (s *CrondSvc) AddIntoPipelineCrond(cronID uint64) error {
	if s.cronChan == nil {
		return fmt.Errorf("delete cron error: cron chan was empty")
	}
	if cronID <= 0 {
		return nil
	}
	s.cronChan <- cronChanPrefixAddKey + strconv.FormatUint(cronID, 10)
	return nil
}

// ReloadCrond 触发当前 crond 实例更新任务.
func (s *CrondSvc) ReloadCrond(pipelineCronFunc func(uint64)) ([]string, error) {
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
