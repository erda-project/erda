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

package nodeStatus

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"
	"strings"
)

func (nodeStatus *NodeStatus) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	nodeStatus.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	nodeStatus.Ctx = ctx
	nodeStatus.SDK = cputil.SDK(ctx)
	node := (*gs)["node"].(data.Object)
	nodeStatus.Props.Text = nodeStatus.SDK.I18n(node.StringSlice("metadata", "fields")[1])
	if strings.ToLower(nodeStatus.Props.Text) != "ready" {
		nodeStatus.Props.Text = "error"
	} else {
		nodeStatus.Props.Text = "success"
	}
	c.Props = nodeStatus.Props
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodeDetail", "nodeStatus", func() servicehub.Provider {
		return &NodeStatus{Type: "Badge", Props: Props{
			Text:   "",
			Status: "",
		}}
	})
}
