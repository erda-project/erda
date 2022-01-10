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

package metric_table_filter

import (
	"fmt"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/recallsong/go-utils/logs"

	slow_transaction "github.com/erda-project/erda/modules/msp/apm/service/common/slow-transaction"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"

	model "github.com/erda-project/erda-infra/providers/component-protocol/components/filter/models"

	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter"
	"github.com/erda-project/erda-infra/providers/component-protocol/components/filter/impl"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
)

type provider struct {
	impl.DefaultFilter
	Log  logs.Logger
	I18n i18n.Translator `autowired:"i18n" translator:"msp-i18n"`
}

func (p *provider) RegisterInitializeOp() (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
		p.StdDataPtr = &filter.Data{
			Conditions: []interface{}{
				model.NewDateRangeCondition(slow_transaction.StateKeyTransactionDurationFilter, p.I18n.Text(sdk.Lang, "Duration")),
			},
			Operations: map[cptype.OperationKey]cptype.Operation{
				filter.OpFilter{}.OpKey(): cputil.NewOpBuilder().Build(),
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
		// todo@ggp get filter data and transfer to global state
	}
}

func (p *provider) RegisterFilterItemSaveOp(opData filter.OpFilterItemSave) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
	}
}

func (p *provider) RegisterFilterItemDeleteOp(opData filter.OpFilterItemDelete) (opFunc cptype.OperationFunc) {
	return func(sdk *cptype.SDK) {
	}
}

func init() {
	name := "component-protocol.components.transaction-rpc-slow.metricTableFilter"
	cpregister.AllExplicitProviderCreatorMap[name] = nil
	base.InitProviderWithCreator("transaction-rpc-slow", "metricTableFilter",
		func() servicehub.Provider { return &provider{} },
	)
}
