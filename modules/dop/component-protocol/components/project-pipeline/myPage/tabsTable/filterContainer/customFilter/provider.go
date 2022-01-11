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

	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
)

type CustomFilter struct {
	impl.DefaultFilter

	bdl      *bundle.Bundle
	gsHelper *gshelper.GSHelper
	sdk      *cptype.SDK
	State    State    `json:"-"`
	InParams InParams `json:"-"`
}

type State struct {
	Base64UrlQueryParams    string             `json:"issueFilter__urlQuery,omitempty"`
	FrontendConditionValues FrontendConditions `json:"values,omitempty"`
	SelectedFilterSet       string             `json:"selectedFilterSet,omitempty"`
}

type FrontendConditions struct {
	Status            []string `json:"status"`
	Creator           []string `json:"creator"`
	AppIDs            []uint64 `json:"appIDs"`
	Executor          []string `json:"executor"`
	CreatedAtStartEnd []string `json:"createdAtStartEnd"`
	StartedAtStartEnd []string `json:"startedAtStartEnd"`
}

func (p *CustomFilter) BeforeHandleOp(sdk *cptype.SDK) {
	p.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	p.gsHelper = gshelper.NewGSHelper(sdk.GlobalState)
	p.sdk = sdk
	if err := p.setInParams(); err != nil {
		panic(err)
	}
	cputil.MustObjJSONTransfer(&p.StdStatePtr, &p.State)
}

func (p *CustomFilter) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		conditions, err := p.ConditionRetriever()
		if err != nil {
			panic(err)
		}
		p.StdDataPtr = &filter.Data{
			Conditions: conditions,
			Operations: map[cptype.OperationKey]cptype.Operation{
				filter.OpFilter{}.OpKey():         cputil.NewOpBuilder().Build(),
				filter.OpFilterItemSave{}.OpKey(): cputil.NewOpBuilder().Build(),
			},
		}
	}
}

func (p *CustomFilter) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

func (p *CustomFilter) RegisterFilterOp(opData filter.OpFilter) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		fmt.Println("state values", p.StdStatePtr)
		fmt.Println("op come", opData.ClientData)
	}
}

func (p *CustomFilter) RegisterFilterItemSaveOp(opData filter.OpFilterItemSave) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		fmt.Println("op come", opData.ClientData)
	}
}

func (p *CustomFilter) RegisterFilterItemDeleteOp(opData filter.OpFilterItemDelete) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		fmt.Println("op come", opData.ClientData.DataRef)
	}
}
