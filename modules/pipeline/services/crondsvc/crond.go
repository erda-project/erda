package crondsvc

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/jsonstore/storetypes"
)

const etcdCrondUpdateWatchKey = "/devops/pipeline/crond/update"

func (s *CrondSvc) ListenCrond(pipelineCronFunc func(id uint64)) {
	logrus.Info("crond: start listen")
	for {
		// watch: isPrefix[true], filterDelete[true], keyOnly[true]
		err := s.js.IncludeWatch().Watch(context.Background(), etcdCrondUpdateWatchKey, false, true, true, nil,
			func(key string, _ interface{}, t storetypes.ChangeType) error {
				// sync update
				logrus.Infof("crond: watched update request key, key: %s", key)
				if _, err := s.ReloadCrond(pipelineCronFunc); err != nil {
					logrus.Errorf("[alert] crond: failed to reload crond (%v)", err)
					return nil
				}
				logrus.Infof("crond: watched and reload successfully")
				return nil
			})
		if err != nil {
			logrus.Errorf("[alert] crond: watch failed, err: %v", err)
		}
	}
}

// DistributedReloadCrond 和 ReloadCrond 不同，触发所有 crond 实例更新任务.
func (s *CrondSvc) DistributedReloadCrond() error {
	return s.js.Put(context.Background(), etcdCrondUpdateWatchKey, "")
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
			if err = s.crond.AddFunc(pc.CronExpr, func() { pipelineCronFunc(pc.ID) }, makePipelineCronName(pc)); err != nil {
				l := fmt.Sprintf("failed to load pipeline cron item: %s, err: %v", makePipelineCronName(pc), err)
				logs = append(logs, l)
				logrus.Errorln("[alert]", l)
				continue
			}
			logs = append(logs, fmt.Sprintf("loaded pipeline cron item: %s", makePipelineCronName(pc)))
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
	logs = append(logs, s.CrondSnapshot()...)

	return logs, nil
}

func (s *CrondSvc) CrondSnapshot() []string {
	var logs []string
	logs = append(logs, "inspecting cron daemon ...")
	for _, entry := range s.crond.Entries() {
		logs = append(logs, fmt.Sprintf("cron daemon task: %s", entry.Name))
	}
	logs = append(logs, "inspect cron daemon DONE")
	return logs
}

func makePipelineCronName(cron spec.PipelineCron) string {
	return fmt.Sprintf("pipeline-cron[%d]-expr[%s]-source[%s]-ymlname[%s]", cron.ID, cron.CronExpr, cron.PipelineSource, cron.PipelineYmlName)
}

func makeCleanBuildCacheJobName(cronExpr string) string {
	return fmt.Sprintf("clean-build-cache-image-[%s]", cronExpr)
}
