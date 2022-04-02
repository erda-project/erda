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

package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/providers/cron/db"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/loop"
	"github.com/erda-project/erda/pkg/strutil"
)

const (
	etcdCronWatchPrefix     = "/devops/pipeline/cron/"
	etcdCronPrefixDeleteKey = etcdCronWatchPrefix + "delete-"
	etcdCronPrefixAddKey    = etcdCronWatchPrefix + "add-"
)

func (d *provider) runCronPipelineFunc(id uint64) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			logrus.Errorf("crond: pipelineCronID: %d, err: %v", id, err)
		}
	}()

	// Get the trigger time immediately
	cronTriggerTime := time.Now()

	// Query cron details
	pc, err := d.dbClient.GetPipelineCron(id)
	if err != nil {
		return
	}

	// if trigger time less than cronStartFrom, return directly
	if d.isCronShouldBeIgnored(pc) {
		logrus.Warnf("crond: pipelineCronID: %d, triggered but ignored, triggerTime: %s, cronStartFrom: %s",
			pc.ID, cronTriggerTime, *pc.Extra.CronStartFrom)
		return
	}

	if pc.Extra.NormalLabels == nil {
		pc.Extra.NormalLabels = make(map[string]string)
	}
	if pc.Extra.FilterLabels == nil {
		pc.Extra.FilterLabels = make(map[string]string)
	}

	// userID
	if pc.Extra.NormalLabels[apistructs.LabelUserID] == "" {
		pc.Extra.NormalLabels[apistructs.LabelUserID] = conf.InternalUserID()
		if err = d.dbClient.UpdatePipelineCron(pc.ID, &pc); err != nil {
			return
		}
	}

	// cron
	if _, ok := pc.Extra.FilterLabels[apistructs.LabelPipelineTriggerMode]; ok {
		pc.Extra.FilterLabels[apistructs.LabelPipelineTriggerMode] = apistructs.PipelineTriggerModeCron.String()
	}

	pc.Extra.NormalLabels[apistructs.LabelPipelineTriggerMode] = apistructs.PipelineTriggerModeCron.String()
	pc.Extra.NormalLabels[apistructs.LabelPipelineType] = apistructs.PipelineTypeNormal.String()
	pc.Extra.NormalLabels[apistructs.LabelPipelineYmlSource] = apistructs.PipelineYmlSourceContent.String()
	pc.Extra.NormalLabels[apistructs.LabelPipelineCronTriggerTime] = strconv.FormatInt(cronTriggerTime.UnixNano(), 10)
	pc.Extra.NormalLabels[apistructs.LabelPipelineCronID] = strconv.FormatUint(pc.ID, 10)

	_, err = d.createPipelineFunc(&apistructs.PipelineCreateRequestV2{
		PipelineYml:            pc.Extra.PipelineYml,
		ClusterName:            pc.Extra.ClusterName,
		PipelineYmlName:        pc.PipelineYmlName,
		PipelineSource:         pc.PipelineSource,
		Labels:                 pc.Extra.FilterLabels,
		NormalLabels:           pc.Extra.NormalLabels,
		Envs:                   pc.Extra.Envs,
		ConfigManageNamespaces: pc.Extra.ConfigManageNamespaces,
		AutoRunAtOnce:          true,
		AutoStartCron:          false,
		IdentityInfo: apistructs.IdentityInfo{
			UserID:         pc.Extra.NormalLabels[apistructs.LabelUserID],
			InternalClient: "system-cron",
		},
		DefinitionID: pc.PipelineDefinitionID,
	})
}

func (s *provider) isCronShouldBeIgnored(pc db.PipelineCron) bool {
	if pc.Extra.CronStartFrom == nil {
		return false
	}
	triggerTime := time.Now()
	return triggerTime.Before(*pc.Extra.CronStartFrom)
}

func (d *provider) DoCrondAbout(ctx context.Context) {

	// load cron info
	logs, err := d.reloadCrond(ctx)
	for _, log := range logs {
		logrus.Info(log)
	}
	if err != nil {
		logrus.Errorf("failed to reload crond from db (%v)", err)
		return
	}

	// watch crond
	go d.listenCrond(ctx)

	// Print a snapshot of a scheduled task regularly
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_ = loop.New(loop.WithInterval(time.Minute)).Do(
					func() (bool, error) {
						for _, log := range d.crondSnapshot() {
							logrus.Debug(log)
						}
						return false, nil
					})
			}
		}
	}()
}

// Listen the key in the channel
// decide to delete or add a scheduled task according to the prefix of the key
func (s *provider) listenCrond(ctx context.Context) {
	s.LeaderWorker.ListenPrefix(ctx, etcdCronWatchPrefix, func(ctx context.Context, event *clientv3.Event) {
		t := event.Type
		key := string(event.Kv.Key)

		s.Log.Infof("crond: watched update etcd key: %s, changeType: %d", key, t)

		if _, err := s.EtcdClient.Delete(ctx, key); err != nil {
			s.Log.Errorf("crond: failed to delete key: %s", key)
			return
		}
		cronID, err := parseCronIDFromWatchedKey(key)
		if err != nil {
			s.Log.Errorf("crond: failed to parse cronID from watched key, key: %s, err: %v", key, err)
			return
		}

		if strings.HasPrefix(key, etcdCronPrefixAddKey) {
			pc, err := s.dbClient.GetPipelineCron(cronID)
			if err != nil {
				s.Log.Errorf("crond: failed to get cron cronID: %v error: %v", cronID, err)
				return
			}
			// why delete it first, because crond.AddFunc cannot add a scheduled task with the same name
			err = s.crond.Remove(makePipelineCronName(cronID))
			if err != nil {
				s.Log.Errorf("crond: failed to remove cron cronID: %v error: %v", cronID, err)
				return
			}

			// determine whether there is a scheduled task
			if pc.Enable != nil && *pc.Enable && pc.CronExpr != "" {
				err = s.crond.AddFunc(pc.CronExpr, func() {
					s.runCronPipelineFunc(pc.ID)
				}, makePipelineCronName(pc.ID))
				if err != nil {
					s.Log.Errorf("crond: failed to update cron cronID: %v cronExpr: %v  error: %v", cronID, pc.CronExpr, err)
					return
				}
			}
		} else if strings.HasPrefix(key, etcdCronPrefixDeleteKey) {
			err = s.crond.Remove(makePipelineCronName(cronID))
			if err != nil {
				s.Log.Errorf("crond: failed to remove cron cronID: %v error: %v", cronID, err)
				return
			}
		}
		s.Log.Infof("crond: watched and reload successfully")
	}, nil)
}

// ReloadCrond triggers the current crond instance update task.
func (s *provider) reloadCrond(ctx context.Context) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var logs []string
	logs = append(logs, "reloading crond from db `pipeline_crons` table ...")

	pcs, err := s.dbClient.ListAllPipelineCrons()
	if err != nil {
		return logs, err
	}

	// Close crond, stop scheduled tasks and etcd connections
	if s.crond != nil {
		s.crond.Close()
		s.crond = nil
	}

	// Initialize a new crond
	s.crond = cron.New(cron.WithoutDLock(true))
	s.crond.Start()
	go func() {
		select {
		case <-ctx.Done():
			s.crond.Stop()
		}
	}()

	for i := range pcs {
		pc := pcs[i]
		if pc.Enable != nil && *pc.Enable && pc.CronExpr != "" {
			if err = s.crond.AddFunc(pc.CronExpr, func() { s.runCronPipelineFunc(pc.ID) }, makePipelineCronName(pc.ID)); err != nil {
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
	if err = s.crond.AddFunc(conf.BuildCacheCleanJobCron(), s.cleanBuildCacheImages, buildCacheCleanJobName); err != nil {
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

func (s *provider) cleanBuildCacheImages() {
	alertErr := func(err error) {
		logrus.Errorf("[alert] failed to clean build cache images, err: %v", err)
	}

	alertErrWithCluster := func(err error, clusterName string) {
		logrus.Errorf("[alert] failed to clean build cache images, clusterName: %s, err: %v", clusterName, err)
	}
	// If the time has not changed for a few days
	date := time.Now().Add(-conf.BuildCacheExpireIn())

	var toDeleteCacheImages []spec.CIV3BuildCache
	if err := s.dbClient.DB().Where("last_pull_at is null and created_at < ?", date).Or("last_pull_at < ?", date).
		Find(&toDeleteCacheImages); err != nil {
		alertErr(err)
		return
	}

	if len(toDeleteCacheImages) == 0 {
		return
	}

	imageMap := make(map[string][]spec.CIV3BuildCache, 0)
	for _, v := range toDeleteCacheImages {
		images, ok := imageMap[v.ClusterName]
		if ok {
			images = append(images, v)
			imageMap[v.ClusterName] = images
		} else {
			imageMap[v.ClusterName] = []spec.CIV3BuildCache{v}
		}
	}

	for clusterName, images := range imageMap {
		var imageNames []string
		for _, v := range images {
			imageNames = append(imageNames, v.Name)
		}
		result, err := s.bdl.DeleteImageManifests(clusterName, imageNames)
		if err != nil {
			alertErrWithCluster(err, clusterName)
			continue
		}
		bytes, err := json.Marshal(result)
		if err != nil {
			alertErrWithCluster(err, clusterName)
			continue
		}

		logrus.Errorf("[alert] clusterName: %s, delete build cache success: %s", clusterName, string(bytes))

		for _, name := range result.Succeed {
			cache, err := s.dbClient.GetBuildCache(clusterName, name)
			if err != nil {
				alertErrWithCluster(err, clusterName)
				continue
			}
			if err = s.dbClient.DeleteBuildCache(cache.ID); err != nil {
				alertErrWithCluster(err, clusterName)
				continue
			}
		}
	}
}

func (s *provider) crondSnapshot() []string {
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

// Remove the prefix of the key and get the cronid
func parseCronIDFromWatchedKey(key string) (uint64, error) {
	pipelineIDStr := strutil.TrimPrefixes(key, etcdCronPrefixDeleteKey)
	pipelineIDStr = strutil.TrimPrefixes(pipelineIDStr, etcdCronPrefixAddKey)
	return strconv.ParseUint(pipelineIDStr, 10, 64)
}
