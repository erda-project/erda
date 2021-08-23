// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package tab

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
)

var (
	ops = map[string]interface{}{
		"onChange": Operation{
			Key:      "changeTab",
			Reload:   true,
			FillMeta: "activeKey",
			Meta:     Meta{ActiveKey: CPU_TAB},
		},
	}
	props = Props{[]MenuPair{
		{
			Key:  CPU_TAB,
			Name: CPU_TAB_ZH,
		},
		{
			Key:  MEM_TAB,
			Name: MEM_TAB_ZH,
		},
	},
	}
	state = State{
		ActiveKey: CPU_TAB,
	}
)

func (t *SteveTab) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	// import components data
	if err := common.Transfer(c.State, &t.State); err != nil {
		logrus.Errorf("import components failed, err:%v", err)
		return err
	}
	switch event.Operation {
	case apistructs.InitializeOperation:
		t.Type = "Tabs"
		t.Props = props
		t.Operations = ops
		t.State = state
	case apistructs.OnChangeOperation:
		if t.State.ActiveKey == CPU_TAB {

		}
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}
	return t.RenderProtocol(c)
}

func (t *SteveTab) RenderProtocol(c *apistructs.Component) error {
	return common.Transfer(*t, c)
}

func RenderCreator() protocol.CompRender {
	return &SteveTab{}
}
