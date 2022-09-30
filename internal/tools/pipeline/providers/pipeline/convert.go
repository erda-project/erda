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

package pipeline

import (
	"strconv"

	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"

	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/numeral"
)

func (s *pipelineService) ConvertPipelineBase(p spec.PipelineBase) basepb.PipelineDTO {
	var result basepb.PipelineDTO
	result.ID = p.ID
	if p.CronID != nil {
		result.CronID = p.CronID
	}
	result.Source = p.PipelineSource.String()
	result.YmlName = p.PipelineYmlName
	result.Type = p.Type.String()
	result.TriggerMode = p.TriggerMode.String()
	result.ClusterName = p.ClusterName
	result.Status = p.Status.String()
	result.CostTimeSec = p.CostTimeSec
	if p.TimeBegin != nil {
		result.TimeBegin = timestamppb.New(*p.TimeBegin)
	}
	if p.TimeEnd != nil {
		result.TimeEnd = timestamppb.New(*p.TimeEnd)
	}
	if p.TimeCreated != nil {
		result.TimeCreated = timestamppb.New(*p.TimeCreated)
	}
	if p.TimeUpdated != nil {
		result.TimeUpdated = timestamppb.New(*p.TimeUpdated)
	}
	return result
}

func (s *pipelineService) ConvertPipeline(p *spec.Pipeline) *basepb.PipelineDTO {
	if p == nil {
		return nil
	}

	var result = s.ConvertPipelineBase(p.PipelineBase)

	// from extra
	if p.TriggerMode == apistructs.PipelineTriggerModeCron && p.Extra.CronTriggerTime != nil {
		// if pipeline is rerun and rerun failed, don't need to convert trigger time
		pipelineType := p.Labels[apistructs.LabelPipelineType]
		if pipelineType != apistructs.PipelineTypeRerun.String() &&
			pipelineType != apistructs.PipelineTypeRerunFailed.String() &&
			p.Extra.CronTriggerTime != nil {
			result.TimeCreated = timestamppb.New(*p.Extra.CronTriggerTime)
			result.TimeBegin = timestamppb.New(*p.Extra.CronTriggerTime)
		}
	}
	result.Extra = &basepb.PipelineExtra{}
	result.Status = s.transferStatusToAnalyzedFailedIfNeed(p).String()
	result.Namespace = p.Extra.Namespace
	result.OrgName = p.GetOrgName()
	result.ProjectName = p.GetLabel(apistructs.LabelProjectName)
	result.ApplicationName = p.GetLabel(apistructs.LabelAppName)
	result.Commit = p.GetCommitID()
	result.CommitDetail = &commonpb.CommitDetail{
		CommitID: p.CommitDetail.CommitID,
		Repo:     p.CommitDetail.Repo,
		RepoAbbr: p.CommitDetail.RepoAbbr,
		Author:   p.CommitDetail.Author,
		Email:    p.CommitDetail.Email,
		Comment:  p.CommitDetail.Comment,
	}
	if p.CommitDetail.Time != nil {
		result.CommitDetail.Time = timestamppb.New(*p.CommitDetail.Time)
	}
	result.YmlSource = p.GetLabel(apistructs.LabelPipelineYmlSource)
	result.YmlNameV1 = p.Extra.PipelineYmlNameV1
	result.YmlContent = p.PipelineYml
	result.Extra.DiceWorkspace = p.Extra.DiceWorkspace.String()
	result.Extra.SubmitUser = p.Extra.SubmitUser
	result.Extra.RunUser = p.Extra.RunUser
	result.Extra.CancelUser = p.Extra.CancelUser
	result.Extra.OwnerUser = p.Extra.OwnerUser
	result.Extra.CronExpr = p.Extra.CronExpr
	if p.Extra.CronTriggerTime != nil {
		result.Extra.CronTriggerTime = timestamppb.New(*p.Extra.CronTriggerTime)
	}
	result.Extra.ShowMessage = p.Extra.ShowMessage
	result.Extra.ConfigManageNamespaces = p.GetConfigManageNamespaces()
	result.Extra.IsAutoRun = p.Extra.IsAutoRun
	result.Extra.CallbackURLs = p.Extra.CallbackURLs
	result.Progress = s.convertProgress(*p)

	// from labels
	orgID, _ := strconv.ParseUint(p.Labels[apistructs.LabelOrgID], 10, 64)
	result.OrgID = orgID
	projectID, _ := strconv.ParseUint(p.Labels[apistructs.LabelProjectID], 10, 64)
	result.ProjectID = projectID
	appID, _ := strconv.ParseUint(p.Labels[apistructs.LabelAppID], 10, 64)
	result.ApplicationID = appID
	result.Branch = p.Labels[apistructs.LabelBranch]

	return &result
}

// transferStatusToAnalyzedIfNeed transfer status to analyzed failed if pipeline is abort run
func (s *pipelineService) transferStatusToAnalyzedFailedIfNeed(p *spec.Pipeline) apistructs.PipelineStatus {
	status := p.Status
	if status == apistructs.PipelineStatusAnalyzed {
		if p.Extra.ShowMessage != nil && p.Extra.ShowMessage.AbortRun {
			status = apistructs.PipelineStatusAnalyzeFailed
		}
	}
	return status
}

func (s *pipelineService) Convert2PagePipeline(p *spec.Pipeline) *pb.PagePipeline {
	result := pb.PagePipeline{
		ID:      p.ID,
		Commit:  p.GetCommitID(),
		Source:  p.PipelineSource.String(),
		YmlName: p.PipelineYmlName,
		Extra: &basepb.PipelineExtra{
			DiceWorkspace:          p.Extra.DiceWorkspace.String(),
			SubmitUser:             p.Extra.SubmitUser,
			RunUser:                p.Extra.RunUser,
			CancelUser:             p.Extra.CancelUser,
			OwnerUser:              p.Extra.OwnerUser,
			CronExpr:               p.Extra.CronExpr,
			ShowMessage:            p.Extra.ShowMessage,
			ConfigManageNamespaces: p.GetConfigManageNamespaces(),
			IsAutoRun:              p.Extra.IsAutoRun,
			CallbackURLs:           p.Extra.CallbackURLs,
			PipelineYmlNameV1:      p.Extra.PipelineYmlNameV1,
		},
		FilterLabels: p.Labels,
		NormalLabels: p.NormalLabels,
		Type:         p.Type.String(),
		TriggerMode:  p.TriggerMode.String(),
		ClusterName:  p.ClusterName,
		Status:       s.transferStatusToAnalyzedFailedIfNeed(p).String(),
		Progress:     s.convertProgress(*p),
		IsSnippet:    p.IsSnippet,
		CostTimeSec:  p.CostTimeSec,
	}
	if p.CronID != nil {
		result.CronID = *p.CronID
	}
	if p.Extra.CronTriggerTime != nil {
		result.Extra.CronTriggerTime = timestamppb.New(*p.Extra.CronTriggerTime)
	}
	if p.ParentPipelineID != nil {
		result.ParentPipelineID = *p.ParentPipelineID
	}
	if p.ParentTaskID != nil {
		result.ParentTaskID = *p.ParentTaskID
	}
	if p.TimeBegin != nil {
		result.TimeBegin = timestamppb.New(*p.TimeBegin)
	}
	if p.TimeEnd != nil {
		result.TimeEnd = timestamppb.New(*p.TimeEnd)
	}
	if p.TimeCreated != nil {
		result.TimeCreated = timestamppb.New(*p.TimeCreated)
	}
	if p.TimeUpdated != nil {
		result.TimeUpdated = timestamppb.New(*p.TimeUpdated)
	}
	if p.TriggerMode == apistructs.PipelineTriggerModeCron && p.Extra.CronTriggerTime != nil {
		pipelineType := p.Labels[apistructs.LabelPipelineType]
		if pipelineType != apistructs.PipelineTypeRerun.String() &&
			pipelineType != apistructs.PipelineTypeRerunFailed.String() &&
			p.Extra.CronTriggerTime != nil {
			result.TimeCreated = timestamppb.New(*p.Extra.CronTriggerTime)
			result.TimeBegin = timestamppb.New(*p.Extra.CronTriggerTime)
		}
	}
	if p.Definition != nil && p.Source != nil {
		definitionPageInfo := &basepb.DefinitionPageInfo{}
		definitionPageInfo.Name = p.Definition.Name
		definitionPageInfo.Creator = p.Definition.Creator
		definitionPageInfo.Executor = p.Definition.Executor
		definitionPageInfo.SourceRef = p.Source.Ref
		definitionPageInfo.SourceRemote = p.Source.Remote
		result.DefinitionPageInfo = definitionPageInfo
	}
	return &result
}

func (s *pipelineService) BatchConvert2PagePipeline(pipelines []spec.Pipeline) []*pb.PagePipeline {
	result := make([]*pb.PagePipeline, 0, len(pipelines))
	for _, p := range pipelines {
		result = append(result, s.Convert2PagePipeline(&p))
	}
	return result
}

// input  progress: int
// output progress: float64
func (s *pipelineService) convertProgress(p spec.Pipeline) float64 {
	progress, err := s.calculateProgress(p)
	if err != nil {
		return 0
	}
	return numeral.Round(float64(progress)/float64(100), 2)
}

// calculateProgress calculates progress based on pipeline tasks
// If progress < 0, it is considered that the progress has not been finalized, including running or not yet calculated;
// If progress >= 0, it is considered that the progress has been calculated (including 0), and return directly
func (s *pipelineService) calculateProgress(p spec.Pipeline) (int, error) {

	if p.Progress >= 0 {
		// pipeline 为成功状态，progress 不应该为 0，需要重新计算
		if p.Progress == 0 && p.Status.IsSuccessStatus() {
			// calculate progress
			goto CalculateStatus
		}
		// progress >= 0，直接返回
		return p.Progress, nil
	}

CalculateStatus:
	needStoreToDB := false
	if p.Status.IsEndStatus() {
		needStoreToDB = true
	}

	// calculate pipeline progress
	tasks, err := s.dbClient.ListPipelineTasksByPipelineID(p.ID)
	if err != nil {
		return -1, err
	}
	var successCount int
	for _, t := range tasks {
		if t.Status.IsSuccessStatus() {
			successCount++
		}
	}
	var progress int
	if len(tasks) == 0 { // 存在 task 为 0 的情况
		progress = 100
	} else {
		progress = (successCount / len(tasks)) * 100
	}

	if needStoreToDB {
		go func() {
			// 异步更新 pipeline progress
			err := s.dbClient.UpdatePipelineProgress(p.ID, progress)
			if err != nil {
				logrus.Errorf("[alert] failed to update pipeline progress, pipelineID: %d, progress: %d, err: %v",
					p.ID, progress, err)
			}
		}()
	}
	return progress, nil
}
