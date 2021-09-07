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

package nodeFilter

import (
	"context"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"strings"

	"github.com/rancher/wrangler/pkg/data"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (i *NodeFilter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	sdk := cputil.SDK(ctx)
	i.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	i.SDK = sdk
	var nodes []data.Object
	// Get all nodes by cluster name
	nodeReq := &apistructs.SteveRequest{}
	nodeReq.OrgID = sdk.Identity.OrgID
	nodeReq.UserID = sdk.Identity.UserID
	nodeReq.Type = apistructs.K8SNode
	if event.Operation == cptype.InitializeOperation {
		if sdk.InParams["clusterName"] != nil {
			nodeReq.ClusterName = sdk.InParams["clusterName"].(string)
			i.ClusterName = nodeReq.ClusterName
		} else {
			return common.ClusterNotFoundErr
		}
	} else {
		nodeReq.ClusterName = i.ClusterName
	}
	resp, err := i.CtxBdl.ListSteveResource(nodeReq)
	if err != nil {
		return err
	}
	nodeList := resp.Slice("data")
	switch event.Operation {
	case cptype.InitializeOperation:
		nodes = nodeList
		break
	case common.CMPDashboardFilterOperationKey:
		if err := common.Transfer(c.State, &i.State); err != nil {
			return err
		}
		labels := make([]string, 0)
		nodeNameFilter := ""
		if len(i.State.Values.Keys) == 0 {
			nodes = nodeList
		} else {
			for k, v := range i.State.Values.Keys {
				if k != "Q" {
					labels = append(labels, v...)
				} else {
					nodeNameFilter = v[0]
				}
			}
			// Filter by node name or node uid
			for _, node := range nodeList {
				for _, l := range labels {
					for _, nl := range node.StringSlice("metadata", "labels") {
						if strings.Contains(nl, l) && strings.Contains(node.String("metadata", "name"), nodeNameFilter) || strings.Contains(node.String("id"), nodeNameFilter) {
							nodes = append(nodes, node)
						}
					}
				}
			}
		}
	}
	(*gs)["nodes"] = nodes
	return i.SetComponentValue(c)
}

// SetComponentValue mapping CpuInfoTable properties to Component
func (i *NodeFilter) SetComponentValue(c *cptype.Component) error {
	var (
		Ops map[string]interface{}
	)

	c.Operations = Ops
	c.Props = i.GetFilterProps()
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "nodeFilter", func() servicehub.Provider {
		return &NodeFilter{}
	})
}
