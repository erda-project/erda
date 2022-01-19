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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline"
)

type CustomFilter struct {
	impl.DefaultFilter

	bdl      *bundle.Bundle
	gsHelper *gshelper.GSHelper
	sdk      *cptype.SDK
	State    State     `json:"state"`
	InParams *InParams `json:"-"`

	projectPipelineSvc projectpipeline.Service `autowired:"erda.dop.projectpipeline.ProjectPipelineService"`
}

type State struct {
	Base64UrlQueryParams    string             `json:"issueFilter__urlQuery,omitempty"`
	FrontendConditionValues FrontendConditions `json:"values,omitempty"`
	SelectedFilterSet       string             `json:"selectedFilterSet,omitempty"`
}

type FrontendConditions struct {
	Status            []string `json:"status"`
	Creator           []string `json:"creator"`
	App               []string `json:"app"`
	Executor          []string `json:"executor"`
	CreatedAtStartEnd []int64  `json:"createdAtStartEnd"`
	StartedAtStartEnd []int64  `json:"startedAtStartEnd"`
}

func (p *CustomFilter) BeforeHandleOp(sdk *cptype.SDK) {
	p.bdl = sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	p.gsHelper = gshelper.NewGSHelper(sdk.GlobalState)
	p.sdk = sdk
	var err error
	p.InParams.OrgID, err = strconv.ParseUint(p.sdk.Identity.OrgID, 10, 64)
	if err != nil {
		panic(err)
	}
	cputil.MustObjJSONTransfer(&p.StdStatePtr, &p.State)
}

func (p *CustomFilter) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
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
			HideSave: true,
		}
		p.clearState()
		var appNames []string
		if p.InParams.AppID != 0 {
			app, err := p.bdl.GetApp(p.InParams.AppID)
			if err != nil {
				logrus.Errorf("failed to GetApp,err %s", err.Error())
				panic(err)
			}
			appNames = []string{app.Name}
		}

		p.State.FrontendConditionValues.App = appNames
		p.State.FrontendConditionValues.Creator = func() []string {
			if p.gsHelper.GetGlobalPipelineTab() == common.MineState.String() {
				return []string{p.sdk.Identity.UserID}
			}
			return nil
		}()
		p.gsHelper.SetGlobalTableFilter(gshelper.TableFilter{
			App:     p.State.FrontendConditionValues.App,
			Creator: p.State.FrontendConditionValues.Creator,
		})

		return nil
	}
}

func (p *CustomFilter) clearState() {
	p.State.FrontendConditionValues = FrontendConditions{}
}

func (p *CustomFilter) AfterHandleOp(sdk *cptype.SDK) {
	cputil.MustObjJSONTransfer(&p.State, &p.StdStatePtr)
}

func (p *CustomFilter) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

func (p *CustomFilter) RegisterFilterOp(opData filter.OpFilter) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
		values := p.State.FrontendConditionValues
		p.gsHelper.SetGlobalTableFilter(gshelper.TableFilter{
			Status:            values.Status,
			Creator:           values.Creator,
			App:               values.App,
			Executor:          values.Executor,
			CreatedAtStartEnd: values.CreatedAtStartEnd,
			StartedAtStartEnd: values.StartedAtStartEnd,
		})
		return nil
	}
}

func (p *CustomFilter) RegisterFilterItemSaveOp(opData filter.OpFilterItemSave) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
		fmt.Println("op come", opData.ClientData)
		return nil
	}
}

func (p *CustomFilter) RegisterFilterItemDeleteOp(opData filter.OpFilterItemDelete) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
		fmt.Println("op come", opData.ClientData.DataRef)
		return nil
	}
}

func (p *CustomFilter) flushOptsByFilter(filterEntity string) error {
	b, err := base64.StdEncoding.DecodeString(filterEntity)
	if err != nil {
		return err
	}
	p.State.FrontendConditionValues = FrontendConditions{}
	return json.Unmarshal(b, &p.State.FrontendConditionValues)
}
