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

package customFilter

import (
	model "github.com/erda-project/erda-infra/providers/component-protocol/components/filter/models"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline-exec-list/common"
)

func (p *CustomFilter) ConditionRetriever() ([]interface{}, error) {
	conditions := make([]interface{}, 0)
	conditions = append(conditions, p.StatusCondition())
	appCondition, err := p.AppCondition()
	if err != nil {
		return nil, err
	}
	conditions = append(conditions, appCondition)

	executorCondition, err := p.MemberCondition("executor")
	if err != nil {
		return nil, err
	}
	executorCondition.ConditionBase.Placeholder = cputil.I18n(p.sdk.Ctx, "please-choose-executor")
	conditions = append(conditions, executorCondition)

	conditions = append(conditions, model.NewDateRangeCondition("startedAtStartEnd", cputil.I18n(p.sdk.Ctx, "started-at")))
	return conditions, nil
}

func (p *CustomFilter) StatusCondition() *model.SelectCondition {
	statuses := apistructs.PipelineAllStatuses
	var opts []model.SelectOption
	for _, status := range statuses {
		opts = append(opts, *model.NewSelectOption(cputil.I18n(p.sdk.Ctx, common.ColumnPipelineStatus+status.String()), status.String()))
	}
	condition := model.NewSelectCondition("status", cputil.I18n(p.sdk.Ctx, "status"), opts)
	condition.ConditionBase.Placeholder = cputil.I18n(p.sdk.Ctx, "please-choose-status")
	return condition
}

func (p *CustomFilter) MemberCondition(key string) (*model.SelectCondition, error) {
	members, err := p.bdl.ListMembers(apistructs.MemberListRequest{
		ScopeType: apistructs.ProjectScope,
		ScopeID:   int64(p.InParams.ProjectIDInt),
		PageNo:    1,
		PageSize:  500,
	})
	if err != nil {
		return nil, err
	}

	condition := model.NewSelectCondition(key, cputil.I18n(p.sdk.Ctx, key), func() []model.SelectOption {
		selectOptions := make([]model.SelectOption, 0, len(members)+1)
		for _, v := range members {
			selectOptions = append(selectOptions, *model.NewSelectOption(func() string {
				if v.Nick != "" {
					return v.Nick
				}
				if v.Name != "" {
					return v.Name
				}
				return v.Mobile
			}(),
				v.UserID,
			))
		}
		selectOptions = append(selectOptions, *model.NewSelectOption(cputil.I18n(p.sdk.Ctx, "choose-yourself"), p.sdk.Identity.UserID).WithFix(true))
		return selectOptions
	}())
	return condition, nil
}

func (p *CustomFilter) AppCondition() (*model.SelectCondition, error) {
	apps, err := p.bdl.GetMyAppsByProject(p.sdk.Identity.UserID, p.InParams.OrgIDInt, p.InParams.ProjectIDInt, "")
	if err != nil {
		return nil, err
	}
	condition := model.NewSelectCondition("appList", cputil.I18n(p.sdk.Ctx, "application"), func() []model.SelectOption {
		selectOptions := make([]model.SelectOption, 0, len(apps.List))
		for _, v := range apps.List {
			selectOptions = append(selectOptions, *model.NewSelectOption(func() string {
				return v.Name
			}(),
				v.ID,
			))
		}
		return selectOptions
	}())
	condition.ConditionBase.Placeholder = cputil.I18n(p.sdk.Ctx, "please-choose-application")
	return condition, nil
}
