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

package pipelinesvc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/thirdparty/gittarutil"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/services/appsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/discover"
)

// RunCronPipelineFunc 定时触发时会先创建 pipeline 记录，然后尝试执行；
// 如果因为某些因素（比如并行数量限制）不能立即执行，用户后续仍然可以再次手动执行这条 pipeline。
func (s *PipelineSvc) RunCronPipelineFunc(id uint64) {
	var err error
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
		if err != nil {
			logrus.Errorf("crond: pipelineCronID: %d, err: %v", id, err)
		}
	}()

	// 立即获取触发时间
	cronTriggerTime := time.Now()

	// 查询 cron 详情
	pc, err := s.dbClient.GetPipelineCron(id)
	if err != nil {
		return
	}

	// 如果当前触发时间小于定时开始时间，return
	if pc.Extra.CronStartFrom != nil && cronTriggerTime.Before(*pc.Extra.CronStartFrom) {
		logrus.Warnf("crond: pipelineCronID: %d, triggered but ignored, triggerTime: %s, cronStartFrom: %s",
			pc.ID, cronTriggerTime, *pc.Extra.CronStartFrom)
		return
	}

	if err = s.UpgradePipelineCron(&pc); err != nil {
		err = errors.Errorf("failed to upgrade pipeline cron, err: %v", err)
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
		if err = s.dbClient.UpdatePipelineCron(pc.ID, &pc); err != nil {
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

	// 使用 v2 方式创建定时流水线
	_, err = s.CreateV2(&apistructs.PipelineCreateRequestV2{
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
	})
}

// 更新老数据
// 1) 根据 application 和 branch 获取元数据
// 2) 根据 basePipelineID 获取元数据
func (s *PipelineSvc) UpgradePipelineCron(pc *spec.PipelineCron) error {

	// 升级后，extra 里 version 为 v2
	if pc.Extra.Version == "v2" {
		return nil
	}

	pc.Extra.Version = "v2"

	if pc.ApplicationID > 0 && pc.Branch != "" {
		var app *appsvc.WorkspaceApp
		app, err := s.appSvc.GetWorkspaceApp(pc.ApplicationID, pc.Branch)
		if err != nil {
			return err
		}
		repo := gittarutil.NewRepo(discover.Gittar(), app.GitRepoAbbrev)
		var f []byte
		f, err = repo.FetchFile(pc.Branch, pc.PipelineYmlName)
		if err != nil {
			return err
		}
		pc.PipelineSource = apistructs.PipelineSourceDice
		if pc.Branch == "master" && pc.PipelineYmlName != apistructs.DefaultPipelineYmlName {
			pc.PipelineSource = apistructs.PipelineSourceBigData
		}
		pc.PipelineYmlName = app.GenerateV1UniquePipelineYmlName(pc.PipelineYmlName)
		pc.Extra.PipelineYml = string(f)
		pc.Extra.ClusterName = app.ClusterName
		pc.Extra.NormalLabels = app.GenerateLabels()
		// commit
		commit, err := repo.GetCommit(app.Branch)
		if err != nil {
			return apierrors.ErrGetGittarRepo.InternalError(err)
		}
		pc.Extra.NormalLabels[apistructs.LabelCommit] = commit.ID
		commitDetail := apistructs.CommitDetail{
			Repo:     app.GitRepo,
			RepoAbbr: app.GitRepoAbbrev,
			Author:   commit.Committer.Name,
			Email:    commit.Committer.Email,
			Time: func() *time.Time {
				commitTime, err := time.Parse("2006-01-02T15:04:05-07:00", commit.Committer.When)
				if err != nil {
					return nil
				}
				return &commitTime
			}(),
			Comment: commit.CommitMessage,
		}
		commitDetailStr, _ := json.Marshal(commitDetail)
		pc.Extra.NormalLabels[apistructs.LabelCommitDetail] = string(commitDetailStr)
	} else if pc.BasePipelineID > 0 {
		var basePipeline spec.Pipeline
		basePipeline, err := s.dbClient.GetPipeline(pc.BasePipelineID)
		if err != nil {
			return err
		}
		pc.PipelineSource = basePipeline.PipelineSource
		pc.PipelineYmlName = basePipeline.GenerateV1UniquePipelineYmlName(pc.PipelineYmlName)
		pc.Extra.PipelineYml = basePipeline.PipelineYml
		pc.Extra.ClusterName = basePipeline.ClusterName
		pc.Extra.NormalLabels = basePipeline.GenerateNormalLabelsForCreateV2()
		pc.Extra.FilterLabels = basePipeline.Labels
	} else {
		return errors.Errorf("both invalid (application_id + branch) and (base_pipeline_id)")
	}

	// 更新 cron 配置至最新
	if err := s.dbClient.UpdatePipelineCron(pc.ID, pc); err != nil {
		return err
	}
	return nil
}
