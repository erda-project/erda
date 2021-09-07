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

package tableTabs

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	common "github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

	"github.com/sirupsen/logrus"
)

var (
	ops = map[string]interface{}{
		"onChange": Operation{
			Key:      "changeTab",
			Reload:   true,
			FillMeta: "activeKey",
		},
	}
	state = State{
		ActiveKey: CPU_TAB,
	}
)

func (t *TableTabs) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	// import components data
	if err := common.Transfer(c.State, &t.State); err != nil {
		logrus.Errorf("import components failed, err:%v", err)
		return err
	}
	t.SDK = cputil.SDK(ctx)
	t.Operations = ops
	(*gs)["activeKey"] = t.State.ActiveKey
	t.Props = Props{[]MenuPair{
		{
			Key:  CPU_TAB,
			Name: t.SDK.I18n(CPU_TAB),
		},
		{
			Key:  MEM_TAB,
			Name: t.SDK.I18n(MEM_TAB),
		},
		{
			Key:  POD_TAB,
			Name: t.SDK.I18n(POD_TAB),
		},
	},
	}
	switch event.Operation {
	case cptype.InitializeOperation:
		t.State.ActiveKey = CPU_TAB
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}
	(*gs)["activeKey"] = t.State.ActiveKey
	return t.RenderProtocol(c)
}

func (t *TableTabs) RenderProtocol(c *cptype.Component) error {
	return common.Transfer(*t, c)
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "tableTabs", func() servicehub.Provider {
		return &TableTabs{Type: "Tabs"}
	})
}
