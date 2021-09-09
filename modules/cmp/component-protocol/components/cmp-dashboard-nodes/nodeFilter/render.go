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
	"strings"

	"github.com/rancher/wrangler/pkg/data"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/filter"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (nf *NodeFilter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	sdk := cputil.SDK(ctx)
	nf.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	nf.SDK = sdk
	nf.Operations = getFilterOperation()
	var nodes []data.Object
	// Get all nodes by cluster name
	nodeReq := &apistructs.SteveRequest{}
	nodeReq.OrgID = sdk.Identity.OrgID
	nodeReq.UserID = sdk.Identity.UserID
	nodeReq.Type = apistructs.K8SNode
	if sdk.InParams["clusterName"] != nil {
		nodeReq.ClusterName = sdk.InParams["clusterName"].(string)
	} else {
		return common.ClusterNotFoundErr
	}

	resp, err := nf.CtxBdl.ListSteveResource(nodeReq)
	if err != nil {
		return err
	}
	nodeList := resp.Slice("data")
	labels := make(map[string]struct{})
	for _, node := range nodeList {
		for k, v := range node.Map("metadata", "labels") {
			labels[k+"="+v.(string)] = struct{}{}
		}
	}
	nf.Props = nf.GetFilterProps(labels)
	switch event.Operation {
	case common.CMPDashboardFilterOperationKey:
		if err := common.Transfer(c.State, &nf.State); err != nil {
			return err
		}
		nodes = DoFilter(nodeList, nf.State.Values)
	default:
		nodes = nodeList
	}
	nf.Operations = getFilterOperation()
	(*gs)["nodes"] = nodes
	return nf.SetComponentValue(c)
}

func DoFilter(nodeList []data.Object, values filter.Values) []data.Object {
	var nodes []data.Object
	labels := make([]string, 0)
	nodeNameFilter := ""
	if values == nil || len(values) == 0 {
		nodes = nodeList
	} else {
		for k, v := range values {
			if k != "Q" {
				labels = append(labels, v)
			} else {
				nodeNameFilter = v
			}
		}
		// Filter by node name
		for _, node := range nodeList {
		NEXT:
			for _, l := range labels {
				for nl := range node.Map("metadata", "labels") {
					if strings.Contains(nl, l) && strings.Contains(node.String("metadata", "name"), nodeNameFilter) || strings.Contains(node.String("id"), nodeNameFilter) {
						nodes = append(nodes, node)
						break NEXT
					}
				}
			}
		}
	}
	return nodes
}

func getFilterOperation() map[string]interface{} {
	ops := filter.Operation{Key: "filter", Reload: true}
	return map[string]interface{}{"filter": ops}
}

// SetComponentValue mapping properties to Component
func (nf *NodeFilter) SetComponentValue(c *cptype.Component) error {
	var (
		err error
	)
	if err = common.Transfer(nf.State, &c.State); err != nil {
		return err
	}
	if err = common.Transfer(nf.Props, &c.Props); err != nil {
		return err
	}
	if err = common.Transfer(nf.Operations, &c.Operations); err != nil {
		return err
	}
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "nodeFilter", func() servicehub.Provider {
		return &NodeFilter{}
	})
}
