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

package groupby_type_count

import (
	structure "github.com/erda-project/erda-infra/providers/component-protocol/components/commodel/data-structure"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/complexgraph"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/complexgraph/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/msp/apm/alert/components/msp-alert-overview/common"
)

type provider struct {
	impl.DefaultComplexGraph
}

// RegisterInitializeOp .
func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) cptype.IStdStructuredPtr {
		data := complexgraph.NewDataBuilder().
			WithTitle("test Graph").
			WithDimensions("Evaporation", "Precipitation", "Temperature").
			WithXAxis(complexgraph.NewAxisBuilder().
				WithType(complexgraph.Category).
				WithData("Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec").
				WithDataStructure(structure.String, "", false).
				Build()).
			WithYAxis(complexgraph.NewAxisBuilder().
				WithType(complexgraph.Value).
				WithName("Evaporation").
				WithPosition(complexgraph.Right).
				WithDimensions("Evaporation").
				WithDataStructure(structure.String, "ml", false).
				Build()).
			WithYAxis(complexgraph.NewAxisBuilder().
				WithType(complexgraph.Value).
				WithName("Precipitation").
				WithPosition(complexgraph.Right).
				WithDimensions("Precipitation").
				WithDataStructure(structure.String, "ml", false).
				Build()).
			WithYAxis(complexgraph.NewAxisBuilder().
				WithType(complexgraph.Value).
				WithName("Temperature").
				WithPosition(complexgraph.Left).
				WithDataStructure(structure.String, "°C", false).
				WithDimensions("Temperature").
				Build()).
			WithSeries(complexgraph.NewSereBuilder().
				WithName("Evaporation").
				WithType(complexgraph.Bar).
				WithDimension("Evaporation").
				WithData(2.0, 4.9, 7.0, 23.2, 25.6, 76.7, 135.6, 162.2, 32.6, 20.0, 6.4, 3.3).
				Build()).
			WithSeries(complexgraph.NewSereBuilder().
				WithName("Precipitation").
				WithType(complexgraph.Bar).
				WithDimension("Precipitation").
				WithData(2.6, 5.9, 9.0, 26.4, 28.7, 70.7, 175.6, 182.2, 48.7, 18.8, 6.0, 2.3).
				Build()).
			WithSeries(complexgraph.NewSereBuilder().
				WithName("Temperature").
				WithType(complexgraph.Line).
				WithDimension("Temperature").
				WithData(2.0, 2.2, 3.3, 4.5, 6.3, 10.2, 20.3, 23.4, 23.0, 16.5, 12.0, 6.2).
				Build()).
			Build()
		return &impl.StdStructuredPtr{
			StdDataPtr: data,
		}
	}
}

// RegisterRenderingOp .
func (p *provider) RegisterRenderingOp() (opFunc cptype.OperationFunc) {
	return p.RegisterInitializeOp()
}

func init() {
	cpregister.RegisterProviderComponent(common.ScenarioKey, common.ComponentNameAlertNotifyGroupByTypeCountLine, &provider{})
}
