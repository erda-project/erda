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

package nodeFilter

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common/filter"
)

// SetCtxBundle 设置bundle
func (i *NodeFilter) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		return common.BundleEmptyErr
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.CtxBdl = b
	return nil
}

func (i *NodeFilter) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	return i.SetComponentValue(c)
}
// SetComponentValue mapping CpuInfoTable properties to Component
func (i *NodeFilter) SetComponentValue(c *apistructs.Component) error {
	var (
		err   error
		state map[string]interface{}
		Ops map[string]interface{}
	)
	if state, err = common.ConvertToMap(i.State); err != nil {
		return err
	}
	if Ops, err = common.ConvertToMap(i.Operations); err != nil {
		return err
	}
	c.State = state
	c.Operations = Ops
	return nil
}

func getFilterProps() filter.Props {
	return filter.PropsInstance
}

func getOperation() map[string]filter.FilterOperation {
	o := make(map[string]filter.FilterOperation)
	o["filter"] = filter.FilterOperation{
		Key:    "filter",
		Reload: true,
	}
	return o
}

func RenderCreator() protocol.CompRender {
	return &NodeFilter{Filter:filter.Filter{
		Type:       "TiledFilter",
		Operations: getOperation(),
		State:      filter.State{},
		Props:      getFilterProps(),
	}}
}
