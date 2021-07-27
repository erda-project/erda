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

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/pbutil"
	"github.com/erda-project/erda/pkg/numeral"
)

func (s *PipelineSvc) convertPipelineBase(p spec.PipelineBase) *pb.PipelineInstance {
	var result pb.PipelineInstance
	result.ID = p.ID
	result.CronID = pbutil.MustGetUint64(p.CronID)
	result.Source = p.PipelineSource.String()
	result.YmlName = p.PipelineYmlName
	result.Type = p.Type.String()
	result.TriggerMode = p.TriggerMode.String()
	result.ClusterName = p.ClusterName
	result.Status = p.Status.String()
	result.CostTimeSec = p.CostTimeSec
	result.TimeBegin = pbutil.GetTimestamp(p.TimeBegin)
	result.TimeEnd = pbutil.GetTimestamp(p.TimeEnd)
	result.TimeCreated = pbutil.GetTimestamp(p.TimeCreated)
	result.TimeUpdated = pbutil.GetTimestamp(p.TimeUpdated)
	return &result
}

func (s *PipelineSvc) ConvertPipeline(p *spec.Pipeline) *pb.PipelineInstance {
	if p == nil {
		return nil
	}

	var result = s.convertPipelineBase(p.PipelineBase)

	// from extra
	if p.TriggerMode == apistructs.PipelineTriggerModeCron && p.Extra.CronTriggerTime != nil {
		result.TimeCreated = pbutil.GetTimestamp(p.Extra.CronTriggerTime)
		result.TimeBegin = pbutil.GetTimestamp(p.Extra.CronTriggerTime)
	}
	result.Namespace = p.Extra.Namespace
	result.OrgName = p.GetOrgName()
	result.ProjectName = p.NormalLabels[apistructs.LabelProjectName]
	result.ApplicationName = p.NormalLabels[apistructs.LabelAppName]
	result.Commit = p.GetCommitID()
	result.CommitDetail = &commonpb.CommitDetail{
		CommitID: p.CommitDetail.CommitID,
		Repo:     p.CommitDetail.Repo,
		RepoAbbr: p.CommitDetail.RepoAbbr,
		Author:   p.CommitDetail.Author,
		Email:    p.CommitDetail.Email,
		Time:     p.CommitDetail.Time,
		Comment:  p.CommitDetail.Comment,
	}
	result.YmlSource = p.NormalLabels[apistructs.LabelPipelineYmlSource]
	result.YmlContent = p.PipelineYml
	result.Extra.DiceWorkspace = p.Extra.DiceWorkspace.String()
	result.Extra.SubmitUser = p.Extra.SubmitUser
	result.Extra.RunUser = p.Extra.RunUser
	result.Extra.CancelUser = p.Extra.CancelUser
	result.Extra.CronExpr = p.Extra.CronExpr
	result.Extra.CronTriggerTime = pbutil.GetTimestamp(p.Extra.CronTriggerTime)
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

	return result
}

func (s *PipelineSvc) Convert2PagePipeline(p *spec.Pipeline) *pb.PagePipeline {
	result := &pb.PagePipeline{
		ID:      p.ID,
		CronID:  pbutil.MustGetUint64(p.CronID),
		Commit:  p.GetCommitID(),
		Source:  p.PipelineSource.String(),
		YmlName: p.PipelineYmlName,
		Extra: &pb.PipelineExtra{
			DiceWorkspace:          p.Extra.DiceWorkspace.String(),
			SubmitUser:             p.Extra.SubmitUser,
			RunUser:                p.Extra.RunUser,
			CancelUser:             p.Extra.CancelUser,
			CronExpr:               p.Extra.CronExpr,
			CronTriggerTime:        pbutil.GetTimestamp(p.Extra.CronTriggerTime),
			ShowMessage:            p.Extra.ShowMessage,
			ConfigManageNamespaces: p.GetConfigManageNamespaces(),
			IsAutoRun:              p.Extra.IsAutoRun,
			CallbackURLs:           p.Extra.CallbackURLs,
		},
		FilterLabels:     p.Labels,
		NormalLabels:     p.NormalLabels,
		Type:             p.Type.String(),
		TriggerMode:      p.TriggerMode.String(),
		ClusterName:      p.ClusterName,
		Status:           p.Status.String(),
		Progress:         s.convertProgress(*p),
		IsSnippet:        p.IsSnippet,
		ParentPipelineID: pbutil.MustGetUint64(p.ParentPipelineID),
		ParentTaskID:     pbutil.MustGetUint64(p.ParentTaskID),
		CostTimeSec:      p.CostTimeSec,
		TimeBegin:        pbutil.GetTimestamp(p.TimeBegin),
		TimeEnd:          pbutil.GetTimestamp(p.TimeEnd),
		TimeCreated:      pbutil.GetTimestamp(p.TimeCreated),
		TimeUpdated:      pbutil.GetTimestamp(p.TimeUpdated),
	}
	if p.TriggerMode == apistructs.PipelineTriggerModeCron && p.Extra.CronTriggerTime != nil {
		result.TimeCreated = pbutil.GetTimestamp(p.Extra.CronTriggerTime)
		result.TimeBegin = pbutil.GetTimestamp(p.Extra.CronTriggerTime)
	}
	return result
}

func (s *PipelineSvc) BatchConvert2PagePipeline(pipelines []spec.Pipeline) []*pb.PagePipeline {
	result := make([]*pb.PagePipeline, 0, len(pipelines))
	for _, p := range pipelines {
		result = append(result, s.Convert2PagePipeline(&p))
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
