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

package filter

import (
	"fmt"
	"reflect"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter/impl"
	model "github.com/erda-project/erda-infra/providers/component-protocol/components/filter/models"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/protocol"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline"
)

type provider struct {
	impl.DefaultFilter
	I18n            i18n.Translator         `autowired:"i18n" translator:"msp-i18n"`
	ProjectPipeline projectpipeline.Service `autowired:"erda.dop.projectpipeline.ProjectPipelineService"`
}

func mockCondition() []interface{} {
	var opt1 = model.NewSelectOption("a-1", "123")
	var opt2 = model.NewSelectOption("b-1", "456")
	var opt3 = model.NewSelectOption("a-2", "234")
	var opt3Fix = model.NewSelectOption("a-2", "234").WithFix(true)
	var opt4 = model.NewSelectChildrenOption("B", "234", []model.SelectOption{*opt2})

	var c1 = model.NewSelectCondition("a", "1", []model.SelectOption{*opt1, *opt3, *opt3Fix}).
		WithMode("single").WithPlaceHolder("select a")
	var c2 = model.NewSelectConditionWithChildren("b", "2", []model.SelectOptionWithChildren{*opt4})
	var c3 = model.NewDateRangeCondition("d", "dasd")
	return []interface{}{c1, c2, c3}
}

func mockFilterSet() []filter.SetItem {
	var i1 = filter.SetItem{
		ID:       "f0",
		Label:    "f0",
		IsPreset: true,
		Values: cptype.ExtraMap{
			"a": "123",
			"b": "456",
		},
	}
	var i2 = filter.SetItem{
		ID:       "f1",
		Label:    "f1",
		IsPreset: false,
		Values: cptype.ExtraMap{
			"a": "234",
		},
	}
	return []filter.SetItem{i1, i2}
}

func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		p.StdDataPtr = &filter.Data{
			Conditions: mockCondition(),
			FilterSet:  mockFilterSet(),
			Operations: map[cptype.OperationKey]cptype.Operation{
				filter.OpFilter{}.OpKey():         cputil.NewOpBuilder().Build(),
				filter.OpFilterItemSave{}.OpKey(): cputil.NewOpBuilder().Build(),
			},
		}
	}
}

func (p *provider) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

func (p *provider) RegisterFilterOp(opData filter.OpFilter) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		fmt.Println("state values", p.StdStatePtr)
		fmt.Println("op come", opData.ClientData)
	}
}

func (p *provider) RegisterFilterItemSaveOp(opData filter.OpFilterItemSave) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		fmt.Println("op come", opData.ClientData)
	}
}

func (p *provider) RegisterFilterItemDeleteOp(opData filter.OpFilterItemDelete) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		fmt.Println("op come", opData.ClientData.DataRef)
	}
}

// Init .
func (p *provider) Init(ctx servicehub.Context) error {
	p.DefaultFilter = impl.DefaultFilter{}
	v := reflect.ValueOf(p)
	v.Elem().FieldByName("Impl").Set(v)
	compName := "customFilter"
	if ctx.Label() != "" {
		compName = ctx.Label()
	}
	protocol.MustRegisterComponent(&protocol.CompRenderSpec{
		Scenario: "project-pipeline-exec-list",
		CompName: compName,
		Creator:  func() cptype.IComponent { return p },
	})
	return nil
}

// Provide .
func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	return p
}

func init() {
	servicehub.Register("component-protocol.components.project-pipeline-exec-list.customFilter", &servicehub.Spec{
		Creator: func() servicehub.Provider { return &provider{} },
	})
}
