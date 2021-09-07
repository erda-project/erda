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

package nodeTitle

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"
)

func (nodeTitle *NodeTitle) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	nodeTitle.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	nodeTitle.Ctx = ctx
	nodeTitle.SDK = cputil.SDK(ctx)
	node := (*gs)["node"].(data.Object)
	nodeTitle.Props.Title = nodeTitle.SDK.I18n("node") + node.String("id")
	c.Props = nodeTitle.Props
	return nil
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodeDetail", "nodeTitle", func() servicehub.Provider {
		return &NodeTitle{
			Props: Props{Title: "Title"},
		}
	})
}
