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

package project_report

import (
	"context"
	"time"

	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda-proto-go/dop/issue/core/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/metrics/report"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/reconciler/rutil"
)

func (p *provider) doReportMetricTask(ctx context.Context) {
	rutil.ContinueWorking(ctx, p.Log, func(ctx context.Context) rutil.WaitDuration {
		// analyzed snippet and non-snippet pipeline database gc
		p.DoReportMetricForAllProjects(ctx)

		return rutil.ContinueWorkingWithDefaultInterval
	}, rutil.WithContinueWorkingDefaultRetryInterval(p.Cfg.ProjectReportMetricDuration))
}

func (p *provider) DoReportMetricForAllProjects(ctx context.Context) {
	_, orgs, err := p.Org.ListOrgs(ctx, []int64{}, &orgpb.ListOrgRequest{
		PageNo:   1,
		PageSize: 9999,
	}, true)
	if err != nil {
		p.Log.Errorf("failed to get all orgs, err: %v", err)
		return
	}
	allOrgs := make(map[uint64]*orgpb.Org)
	for _, org := range orgs {
		allOrgs[org.ID] = org
	}
	projects, err := p.bdl.GetAllProjects()
	if err != nil {
		p.Log.Errorf("failed to get all projects, err: %v", err)
		return
	}
	for _, projectDto := range projects {
		orgDto, ok := allOrgs[projectDto.OrgID]
		if !ok {
			p.Log.Errorf("failed to find org for project: %s, projectID: %d, orgID: %d",
				projectDto.Name, projectDto.ID, projectDto.OrgID)
			continue
		}
		numberFields, err := p.CalculateProjectMetrics(projectDto.ID)
		if err != nil {
			p.Log.Errorf("failed to calculate project metrics for project: %s, projectID: %d, err: %v",
				projectDto.Name, projectDto.ID, err)
			continue
		}
		_, _, metricLabels := generateProjectMetricLabels(&projectDto, orgDto)
		if err := p.ReportClient.Send([]*report.Metric{{
			Name:      metricGroupName,
			Timestamp: time.Now().UnixNano(),
			Tags:      metricLabels,
			Fields:    numberFields,
		}}); err != nil {
			p.Log.Errorf("failed to push project project metrics, projectID: %d, err: %v", projectDto.ID, err)
			continue
		}
		p.Log.Infof("success to push project metrics, projectName: %s, projectID: %d", projectDto.Name, projectDto.ID)
	}
}

func (p *provider) CalculateProjectMetrics(projectID uint64) (map[string]interface{}, error) {
	fields := make(map[string]interface{})

	requirementTotal, err := p.getRequirementTotalNum(projectID)
	if err != nil {
		return nil, err
	}
	fields[TotalRequirementNum] = requirementTotal

	taskTotal, err := p.getTaskTotalNum(projectID)
	if err != nil {
		return nil, err
	}
	fields[TotalTaskNum] = taskTotal

	return fields, nil
}

// getRequirementTotalNum get requirement issue total num for all states of the project
func (p *provider) getRequirementTotalNum(projectID uint64) (uint64, error) {
	_, requirementTotal, err := p.IssueSvc.Paging(pb.PagingIssueRequest{
		ProjectID: projectID,
		Type: []string{
			apistructs.IssueTypeRequirement.String(),
		},
		External: true,
	})
	if err != nil {
		return 0, err
	}

	return requirementTotal, nil
}

// getTaskTotalNum task type issue total num for all states of the project
func (p *provider) getTaskTotalNum(projectID uint64) (uint64, error) {
	_, taskTotal, err := p.IssueSvc.Paging(pb.PagingIssueRequest{
		ProjectID: projectID,
		Type: []string{
			apistructs.IssueTypeTask.String(),
		},
		External: true,
	})
	if err != nil {
		return 0, err
	}

	return taskTotal, nil
}
