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

package issueFilter

import (
	"encoding/base64"
	"encoding/json"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-kanban/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/issuefilterbm"
	"github.com/erda-project/erda/modules/dop/services/issuestate"
	"github.com/erda-project/erda/pkg/strutil"
)

type IssueFilter struct {
	impl.DefaultFilter

	bdl              *bundle.Bundle
	issueStateSvc    *issuestate.IssueState
	issueFilterBmSvc *issuefilterbm.IssueFilterBookmark
	gsHelper         *gshelper.GSHelper
	sdk              *cptype.SDK

	filterReq apistructs.IssuePagingRequest `json:"-"`
	State     State                         `json:"_"`
	InParams  InParams                      `json:"-"`
	Bms       []issuefilterbm.MyFilterBm    `json:"-"`
}

type State struct {
	Base64UrlQueryParams    string             `json:"issueFilter__urlQuery,omitempty"`
	FrontendConditionValues FrontendConditions `json:"values,omitempty"`
	SelectedFilterSet       string             `json:"selectedFilterSet,omitempty"`
}

func init() {
	base.InitProviderToDefaultNamespace("issueFilter", func() servicehub.Provider {
		return &IssueFilter{}
	})
}

func (f *IssueFilter) BeforeHandleOp(sdk *cptype.SDK) {
	f.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	f.issueFilterBmSvc = sdk.Ctx.Value(types.IssueFilterBmService).(*issuefilterbm.IssueFilterBookmark)
	f.issueStateSvc = sdk.Ctx.Value(types.IssueStateService).(*issuestate.IssueState)
	f.gsHelper = gshelper.NewGSHelper(sdk.GlobalState)
	f.sdk = sdk
	if err := f.setInParams(); err != nil {
		panic(err)
	}
	cputil.MustObjJSONTransfer(&f.StdStatePtr, &f.State)
	if err := f.initFilterBms(); err != nil {
		panic(err)
	}
}

func (f *IssueFilter) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		conditions, err := f.ConditionRetriever()
		if err != nil {
			panic(err)
		}
		f.StdDataPtr.Conditions = conditions
		options, err := f.FilterSet()
		if err != nil {
			panic(err)
		}
		f.StdDataPtr.FilterSet = options
		if f.InParams.FrontendUrlQuery != "" {
			if err := f.flushOptsByFilter(f.InParams.FrontendUrlQuery); err != nil {
				panic(err)
			}
		}
		// if f.State.FrontendConditionValues.States == nil {
		// 	if err := f.setDefaultState(); err != nil {
		// 		panic(err)
		// 	}
		// }
		f.StdDataPtr.Operations = map[cptype.OperationKey]cptype.Operation{
			filter.OpFilter{}.OpKey():           cputil.NewOpBuilder().Build(),
			filter.OpFilterItemSave{}.OpKey():   cputil.NewOpBuilder().Build(),
			filter.OpFilterItemDelete{}.OpKey(): cputil.NewOpBuilder().Build(),
		}
	}
}

func (f *IssueFilter) flushOptsByFilter(filterEntity string) error {
	b, err := base64.StdEncoding.DecodeString(filterEntity)
	if err != nil {
		return err
	}
	f.State.FrontendConditionValues = FrontendConditions{}
	return json.Unmarshal(b, &f.State.FrontendConditionValues)
}

func (f *IssueFilter) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return f.RegisterInitializeOp()
}

func (f *IssueFilter) RegisterFilterOp(opData filter.OpFilter) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
	}
}

func (f *IssueFilter) AfterHandleOp(sdk *cptype.SDK) {
	query, err := f.generateUrlQueryParams()
	if err != nil {
		panic(err)
	}
	f.State.Base64UrlQueryParams = query
	cputil.MustObjJSONTransfer(&f.State, &f.StdStatePtr)
}

func (f *IssueFilter) RegisterFilterItemSaveOp(opData filter.OpFilterItemSave) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		var data FilterSetData
		cputil.MustObjJSONTransfer(&opData.ClientData, &data)
		entity, err := f.CreateFilterSetEntity(opData.ClientData.Values)
		if err != nil {
			panic(err)
		}
		pageKey := f.issueFilterBmSvc.GenPageKey(f.InParams.FrontendFixedIteration, f.InParams.FrontendFixedIssueType)
		_, err = f.issueFilterBmSvc.Create(&dao.IssueFilterBookmark{
			Name:         data.Label,
			UserID:       f.sdk.Identity.UserID,
			ProjectID:    strutil.String(f.InParams.ProjectID),
			PageKey:      pageKey,
			FilterEntity: entity,
		})
		if err != nil {
			panic(err)
		}
		if err := f.initFilterBms(); err != nil {
			panic(err)
		}
		options, err := f.FilterSet()
		if err != nil {
			panic(err)
		}
		f.StdDataPtr.FilterSet = options
	}
}

func (f *IssueFilter) RegisterFilterItemDeleteOp(opData filter.OpFilterItemDelete) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		// fmt.Println("delete op come", opData.ClientData.DataRef)
		if err := f.issueFilterBmSvc.Delete(opData.ClientData.DataRef.ID); err != nil {
			panic(err)
		}
		if err := f.initFilterBms(); err != nil {
			panic(err)
		}
		options, err := f.FilterSet()
		if err != nil {
			panic(err)
		}
		f.StdDataPtr.FilterSet = options
	}
}

func (f *IssueFilter) Finalize(sdk *cptype.SDK) {
	issuePagingRequest := f.generateIssuePagingRequest()
	if req, ok := f.gsHelper.GetIssuePagingRequest(); ok {
		issuePagingRequest.Title = req.Title
	}
	f.gsHelper.SetIssuePagingRequest(issuePagingRequest)
}

func (f *IssueFilter) generateIssuePagingRequest() apistructs.IssuePagingRequest {
	var (
		startCreatedAt, endCreatedAt, startFinishedAt, endFinishedAt, startClosedAt, endClosedAt int64
	)
	if len(f.State.FrontendConditionValues.CreatedAtStartEnd) >= 2 {
		if f.State.FrontendConditionValues.CreatedAtStartEnd[0] != nil {
			startCreatedAt = *f.State.FrontendConditionValues.CreatedAtStartEnd[0]
			if f.State.FrontendConditionValues.CreatedAtStartEnd[1] == nil {
				endCreatedAt = 0
			} else {
				endCreatedAt = *f.State.FrontendConditionValues.CreatedAtStartEnd[1]
			}
		} else if f.State.FrontendConditionValues.CreatedAtStartEnd[1] != nil {
			startCreatedAt = 0
			endCreatedAt = *f.State.FrontendConditionValues.CreatedAtStartEnd[1]
		}
	}
	if len(f.State.FrontendConditionValues.FinishedAtStartEnd) >= 2 {
		if f.State.FrontendConditionValues.FinishedAtStartEnd[0] != nil {
			startFinishedAt = *f.State.FrontendConditionValues.FinishedAtStartEnd[0]
			if f.State.FrontendConditionValues.FinishedAtStartEnd[1] == nil {
				endFinishedAt = 0
			} else {
				endFinishedAt = *f.State.FrontendConditionValues.FinishedAtStartEnd[1]
			}
		} else if f.State.FrontendConditionValues.FinishedAtStartEnd[1] != nil {
			startFinishedAt = 0
			endFinishedAt = *f.State.FrontendConditionValues.FinishedAtStartEnd[1]
		}
	}
	if len(f.State.FrontendConditionValues.ClosedAtStartEnd) >= 2 {
		if f.State.FrontendConditionValues.ClosedAtStartEnd[0] != nil {
			startClosedAt = *f.State.FrontendConditionValues.ClosedAtStartEnd[0]
			if f.State.FrontendConditionValues.ClosedAtStartEnd[1] == nil {
				endClosedAt = 0
			} else {
				endClosedAt = *f.State.FrontendConditionValues.ClosedAtStartEnd[1]
			}
		} else if f.State.FrontendConditionValues.ClosedAtStartEnd[1] != nil {
			startClosedAt = 0
			endClosedAt = *f.State.FrontendConditionValues.ClosedAtStartEnd[1]
		}
	}

	req := apistructs.IssuePagingRequest{
		PageNo:   1, // 每次走 filter，都需要重新查询，调整 pageNo 为 1
		PageSize: 0,
		OrgID:    int64(f.InParams.OrgID),
		IssueListRequest: apistructs.IssueListRequest{
			// Title:           f.State.FrontendConditionValues.Title,
			Type:            f.InParams.IssueTypes,
			ProjectID:       f.InParams.ProjectID,
			IterationID:     f.InParams.IterationID,
			IterationIDs:    f.State.FrontendConditionValues.IterationIDs,
			AppID:           nil,
			RequirementID:   nil,
			State:           f.State.FrontendConditionValues.States,
			StateBelongs:    nil,
			Creators:        f.State.FrontendConditionValues.CreatorIDs,
			Assignees:       f.State.FrontendConditionValues.AssigneeIDs,
			Label:           f.State.FrontendConditionValues.LabelIDs,
			StartCreatedAt:  startCreatedAt,
			EndCreatedAt:    endCreatedAt,
			StartFinishedAt: startFinishedAt,
			EndFinishedAt:   endFinishedAt,
			StartClosedAt:   startClosedAt,
			EndClosedAt:     endClosedAt,
			Priority:        f.State.FrontendConditionValues.Priorities,
			Complexity:      f.State.FrontendConditionValues.Complexities,
			Severity:        f.State.FrontendConditionValues.Severities,
			RelatedIssueIDs: nil,
			Source:          "",
			OrderBy:         "updated_at",
			TaskType:        nil,
			BugStage:        f.State.FrontendConditionValues.BugStages,
			Owner:           f.State.FrontendConditionValues.OwnerIDs,
			Asc:             false,
			IDs:             nil,
			IdentityInfo:    apistructs.IdentityInfo{UserID: f.sdk.Identity.UserID},
			External:        false,

			WithProcessSummary: f.InParams.FrontendFixedIssueType == apistructs.IssueTypeRequirement.String(),
		},
	}
	return req
}

func (f *IssueFilter) setDefaultState() error {
	stateBelongs := map[string][]apistructs.IssueStateBelong{
		"TASK":        {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking},
		"REQUIREMENT": {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking},
		"BUG":         {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking, apistructs.IssueStateBelongWontfix, apistructs.IssueStateBelongReopen, apistructs.IssueStateBelongResolved},
		"ALL":         {apistructs.IssueStateBelongOpen, apistructs.IssueStateBelongWorking, apistructs.IssueStateBelongWontfix, apistructs.IssueStateBelongReopen, apistructs.IssueStateBelongResolved},
	}[f.InParams.FrontendFixedIssueType]
	types := []apistructs.IssueType{apistructs.IssueTypeRequirement, apistructs.IssueTypeTask, apistructs.IssueTypeBug}
	res := make(map[string][]int64)
	res["ALL"] = make([]int64, 0)
	for _, v := range types {
		req := &apistructs.IssueStatesGetRequest{
			ProjectID:    f.InParams.ProjectID,
			StateBelongs: stateBelongs,
			IssueType:    v,
		}
		ids, err := f.issueStateSvc.GetIssueStateIDs(req)
		if err != nil {
			return err
		}
		res[v.String()] = ids
		res["ALL"] = append(res["ALL"], ids...)
	}
	f.State.FrontendConditionValues.States = res[f.InParams.FrontendFixedIssueType]
	return nil
}
