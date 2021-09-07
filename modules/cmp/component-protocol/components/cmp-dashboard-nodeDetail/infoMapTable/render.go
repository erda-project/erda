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

package infoMapTable

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"
	"strings"
)

func (infoMapTable *InfoMapTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	node := (*gs)["node"].(data.Object)
	infoMapTable.SDK = cputil.SDK(ctx)
	pairs := make([]Pair, 0)
	for _, k := range node.StringSlice("nodeInfo") {
		pair := strings.Split(k, ":")
		pairs = append(pairs, Pair{
			Value: infoMapTable.SDK.I18n(pair[0]),
			Label: Label{
				Value:       pair[1],
				RenderType:  "text",
				StyleConfig: StyleConfig{"bold"},
			},
		})
	}
	c.Data["list"] = pairs
	return nil
}

func getProps() Props {
	return Props{
		Bordered:   true,
		ShowHeader: false,
		Pagination: false,
		Columns: []Column{{
			DataIndex: "label",
			Title:     "",
			width:     200,
		}, {
			DataIndex: "value",
			Title:     "",
			width:     200},
		},
	}
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodeDetail", "infoMapTable", func() servicehub.Provider {
		return &InfoMapTable{Type: "Table", Props: getProps()}
	})
}
