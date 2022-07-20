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
	"fmt"
	"strconv"

	model "github.com/erda-project/erda-infra/providers/component-protocol/components/filter/models"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/components/project-pipeline-exec-list/common"
	"github.com/erda-project/erda/internal/apps/dop/component-protocol/components/util"
	"github.com/erda-project/erda/internal/pkg/component-protocol/condition"
	"github.com/erda-project/erda/pkg/limit_sync_group"
)

type UserType string

const (
	ExecutorUser UserType = "executor"
	OwnerUser    UserType = "owner"
)

func (p *CustomFilter) ConditionRetriever() ([]interface{}, error) {
	conditions := make([]interface{}, 0)
	conditions = append(conditions, p.StatusCondition())

	var (
		appCondition, executorCondition, ownerCondition *model.SelectCondition
	)
	worker := limit_sync_group.NewWorker(2)
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		var err error
		appCondition, err = p.AppCondition()
		return err
	})
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		var err error
		members, err := p.getMembers()
		if err != nil {
			return err
		}
		executorCondition, err = p.MemberCondition(ExecutorUser, members)
		if err != nil {
			return err
		}
		ownerCondition, err = p.MemberCondition(OwnerUser, members)
		if err != nil {
			return err
		}

		return err
	})
	if err := worker.Do().Error(); err != nil {
		return nil, err
	}

	if p.InParams.AppIDInt == 0 {
		conditions = append(conditions, appCondition)
	}
	conditions = append(conditions, p.TriggerModeCondition())

	conditions = append(conditions, model.NewDateRangeCondition("startedAtStartEnd", cputil.I18n(p.sdk.Ctx, "start-time")))
	conditions = append(conditions, executorCondition)
	conditions = append(conditions, ownerCondition)
	conditions = append(conditions, condition.ExternalInputCondition("title", "title", cputil.I18n(p.sdk.Ctx, "searchByPipelineName")))
	return conditions, nil
}

func (p *CustomFilter) StatusCondition() *model.SelectCondition {
	statuses := util.PipelineDefinitionStatus
	var opts []model.SelectOption
	for _, status := range statuses {
		opts = append(opts, *model.NewSelectOption(cputil.I18n(p.sdk.Ctx, common.ColumnPipelineStatus+status.String()), status.String()))
	}
	condition := model.NewSelectCondition("status", cputil.I18n(p.sdk.Ctx, "status"), opts)
	condition.ConditionBase.Placeholder = cputil.I18n(p.sdk.Ctx, "please-choose-status")
	return condition
}

func (p *CustomFilter) TriggerModeCondition() *model.SelectCondition {
	triggerModes := util.PipelineDefinitionTriggers
	var opts []model.SelectOption
	for _, triggerMode := range triggerModes {
		opts = append(opts, *model.NewSelectOption(cputil.I18n(p.sdk.Ctx, common.ColumnPipelineTrigger+triggerMode.String()), triggerMode.String()))
	}
	condition := model.NewSelectCondition("triggerMode", cputil.I18n(p.sdk.Ctx, "triggerMode"), opts)
	condition.ConditionBase.Placeholder = cputil.I18n(p.sdk.Ctx, "please-choose-triggerMode")
	return condition
}

func (p *CustomFilter) getMembers() ([]apistructs.Member, error) {
	members, err := p.bdl.ListMembers(apistructs.MemberListRequest{
		ScopeType: apistructs.ProjectScope,
		ScopeID:   int64(p.InParams.ProjectIDInt),
		PageNo:    1,
		PageSize:  500,
	})
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (p *CustomFilter) MemberCondition(userType UserType, members []apistructs.Member) (*model.SelectCondition, error) {
	var condition *model.SelectCondition
	selectOptionFunc := func() []model.SelectOption {
		selectOptions := make([]model.SelectOption, 0, len(members)+1)
		for _, v := range members {
			selectOptions = append(selectOptions, *model.NewSelectOption(
				v.GetUserName(),
				v.UserID,
			))
		}
		selectOptions = append(selectOptions, *model.NewSelectOption(cputil.I18n(p.sdk.Ctx, "choose-yourself"), p.sdk.Identity.UserID).WithFix(true))
		return selectOptions
	}
	switch userType {
	case ExecutorUser:
		condition = model.NewSelectCondition("executor", cputil.I18n(p.sdk.Ctx, "executor"), selectOptionFunc())
		condition.ConditionBase.Placeholder = cputil.I18n(p.sdk.Ctx, "please-choose-executor")
	case OwnerUser:
		condition = model.NewSelectCondition("owner", cputil.I18n(p.sdk.Ctx, "owner"), selectOptionFunc())
		condition.ConditionBase.Placeholder = cputil.I18n(p.sdk.Ctx, "please-choose-owner")
	default:
		return nil, fmt.Errorf("unknown user type: %s", userType)
	}

	return condition, nil
}

func (p *CustomFilter) AppCondition() (*model.SelectCondition, error) {
	if p.InParams.AppIDInt != 0 {
		return p.AppConditionWithInParamsAppID()
	}
	return p.AppConditionWithNoInParamsAppID()
}

func (p *CustomFilter) AppConditionWithNoInParamsAppID() (*model.SelectCondition, error) {
	var (
		allApps      []apistructs.ApplicationDTO
		myAppNames   []string
		appIDNameMap = make(map[string]string)
	)

	worker := limit_sync_group.NewWorker(2)
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		allAppResp, err := p.bdl.GetAppList(p.sdk.Identity.OrgID, p.sdk.Identity.UserID, apistructs.ApplicationListRequest{
			ProjectID: p.InParams.ProjectIDInt,
			PageNo:    1,
			PageSize:  999,
			IsSimple:  true,
		})
		if err != nil {
			return err
		}
		allApps = allAppResp.List
		return nil
	})
	worker.AddFunc(func(locker *limit_sync_group.Locker, i ...interface{}) error {
		myAppResp, err := p.bdl.GetMyAppsByProject(p.sdk.Identity.UserID, p.InParams.OrgIDInt, p.InParams.ProjectIDInt, "")
		if err != nil {
			return err
		}
		for _, v := range myAppResp.List {
			myAppNames = append(myAppNames, v.Name)
		}
		return nil
	})
	if err := worker.Do().Error(); err != nil {
		return nil, err
	}

	cond := model.NewSelectCondition("appList", cputil.I18n(p.sdk.Ctx, "application"), func() []model.SelectOption {
		selectOptions := make([]model.SelectOption, 0, len(allApps)+1)
		selectOptions = append(selectOptions, *model.NewSelectOption(cputil.I18n(p.sdk.Ctx, "participated"), uint64(common.Participated)))
		for _, v := range allApps {
			selectOptions = append(selectOptions, *model.NewSelectOption(v.Name, v.ID))
			appIDNameMap[strconv.FormatUint(v.ID, 10)] = v.Name
		}
		return selectOptions
	}())
	cond.ConditionBase.Disabled = false
	cond.ConditionBase.Placeholder = cputil.I18n(p.sdk.Ctx, "please-choose-application")

	p.gsHelper.SetGlobalMyAppNames(myAppNames)
	p.gsHelper.SetGlobalAppIDNameMap(appIDNameMap)
	return cond, nil
}

func (p *CustomFilter) AppConditionWithInParamsAppID() (*model.SelectCondition, error) {
	app, err := p.bdl.GetApp(p.InParams.AppIDInt)
	if err != nil {
		return nil, err
	}
	cond := model.NewSelectCondition("appList", cputil.I18n(p.sdk.Ctx, "application"), []model.SelectOption{*model.NewSelectOption(app.Name, app.ID)})
	cond.ConditionBase.Disabled = true
	cond.ConditionBase.Placeholder = cputil.I18n(p.sdk.Ctx, "please-choose-application")
	p.gsHelper.SetGlobalInParamsAppName(app.Name)
	return cond, nil
}
