// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package pipelinesvc

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/numeral"
)

func (s *PipelineSvc) convertPipelineBase(p spec.PipelineBase) apistructs.PipelineDTO {
	var result apistructs.PipelineDTO
	result.ID = p.ID
	result.CronID = p.CronID
	result.Source = p.PipelineSource
	result.YmlName = p.PipelineYmlName
	result.Type = p.Type.String()
	result.TriggerMode = p.TriggerMode.String()
	result.ClusterName = p.ClusterName
	result.Status = p.Status
	result.CostTimeSec = p.CostTimeSec
	result.TimeBegin = p.TimeBegin
	result.TimeEnd = p.TimeEnd
	result.TimeCreated = p.TimeCreated
	result.TimeUpdated = p.TimeUpdated
	return result
}

func (s *PipelineSvc) ConvertPipeline(p *spec.Pipeline) *apistructs.PipelineDTO {
	if p == nil {
		return nil
	}

	var result = s.convertPipelineBase(p.PipelineBase)

	// from extra
	if p.TriggerMode == apistructs.PipelineTriggerModeCron && p.Extra.CronTriggerTime != nil {
		result.TimeCreated = p.Extra.CronTriggerTime
		result.TimeBegin = p.Extra.CronTriggerTime
	}
	result.Namespace = p.Extra.Namespace
	result.OrgName = p.GetOrgName()
	result.ProjectName = p.GetLabel(apistructs.LabelProjectName)
	result.ApplicationName = p.GetLabel(apistructs.LabelAppName)
	result.Commit = p.GetCommitID()
	result.CommitDetail = p.CommitDetail
	result.YmlSource = p.GetLabel(apistructs.LabelPipelineYmlSource)
	result.YmlNameV1 = p.Extra.PipelineYmlNameV1
	result.YmlContent = p.PipelineYml
	result.Extra.DiceWorkspace = p.Extra.DiceWorkspace.String()
	result.Extra.SubmitUser = p.Extra.SubmitUser
	result.Extra.RunUser = p.Extra.RunUser
	result.Extra.CancelUser = p.Extra.CancelUser
	result.Extra.CronExpr = p.Extra.CronExpr
	result.Extra.CronTriggerTime = p.Extra.CronTriggerTime
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

func (s *PipelineSvc) Convert2PagePipeline(p *spec.Pipeline) *apistructs.PagePipeline {
	result := apistructs.PagePipeline{
		ID:      p.ID,
		CronID:  p.CronID,
		Commit:  p.GetCommitID(),
		Source:  p.PipelineSource,
		YmlName: p.PipelineYmlName,
		Extra: apistructs.PipelineExtra{
			DiceWorkspace:          p.Extra.DiceWorkspace.String(),
			SubmitUser:             p.Extra.SubmitUser,
			RunUser:                p.Extra.RunUser,
			CancelUser:             p.Extra.CancelUser,
			CronExpr:               p.Extra.CronExpr,
			CronTriggerTime:        p.Extra.CronTriggerTime,
			ShowMessage:            p.Extra.ShowMessage,
			ConfigManageNamespaces: p.GetConfigManageNamespaces(),
			IsAutoRun:              p.Extra.IsAutoRun,
			CallbackURLs:           p.Extra.CallbackURLs,
			PipelineYmlNameV1:      p.Extra.PipelineYmlNameV1,
		},
		FilterLabels:     p.Labels,
		NormalLabels:     p.NormalLabels,
		Type:             p.Type.String(),
		TriggerMode:      p.TriggerMode.String(),
		ClusterName:      p.ClusterName,
		Status:           p.Status,
		Progress:         s.convertProgress(*p),
		IsSnippet:        p.IsSnippet,
		ParentPipelineID: p.ParentPipelineID,
		ParentTaskID:     p.ParentTaskID,
		CostTimeSec:      p.CostTimeSec,
		TimeBegin:        p.TimeBegin,
		TimeEnd:          p.TimeEnd,
		TimeCreated:      p.TimeCreated,
		TimeUpdated:      p.TimeUpdated,
	}
	if p.TriggerMode == apistructs.PipelineTriggerModeCron && p.Extra.CronTriggerTime != nil {
		result.TimeCreated = p.Extra.CronTriggerTime
		result.TimeBegin = p.Extra.CronTriggerTime
	}
	return &result
}

func (s *PipelineSvc) BatchConvert2PagePipeline(pipelines []spec.Pipeline) []apistructs.PagePipeline {
	result := make([]apistructs.PagePipeline, 0, len(pipelines))
	for _, p := range pipelines {
		result = append(result, *s.Convert2PagePipeline(&p))
	}
	return result
}

// input  progress: int
// output progress: float64
func (s *PipelineSvc) convertProgress(p spec.Pipeline) float64 {
	progress, err := s.calculateProgress(p)
	if err != nil {
		return 0
	}
	return numeral.Round(float64(progress)/float64(100), 2)
}
