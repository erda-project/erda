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

	"github.com/pkg/errors"
	"github.com/rancher/wrangler/pkg/data"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common/filter"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

var steveServer cmp.SteveServer

func (nf *NodeFilter) Init(ctx servicehub.Context) error {
	server, ok := ctx.Service("cmp").(cmp.SteveServer)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a steveServer")
	}
	steveServer = server
	return nf.DefaultProvider.Init(ctx)
}

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

	resp, err := steveServer.ListSteveResource(ctx, nodeReq)
	if err != nil {
		return err
	}
	var nodeList []data.Object
	for _, item := range resp {
		nodeList = append(nodeList, item.Data())
	}
	labels := make(map[string]struct{})
	for _, node := range nodeList {
		for k, v := range node.Map("metadata", "labels") {
			labels[k+"="+v.(string)] = struct{}{}
		}
	}
	nf.Props = nf.GetFilterProps(labels)
	switch event.Operation {
	case common.CMPDashboardFilterOperationKey, common.CMPDashboardTableTabs, cptype.RenderingOperation:
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

func isEmptyFilter(values filter.Values) bool {
	for k, v := range values {
		if k == "Q" {
			if v.(string) != "" {
				return false
			}
		} else {
			if v == nil || len(v.([]interface{})) != 0 {
				return false
			}
		}
	}
	return true
}

func DoFilter(nodeList []data.Object, values filter.Values) []data.Object {
	var nodes []data.Object
	labels := make([]string, 0)
	nodeNameFilterName := ""
	if isEmptyFilter(values) {
		nodes = nodeList
	} else {
		for k, v := range values {
			if k != "Q" {
				vs := v.([]interface{})
				ss := make([]string, 0)
				for _, s := range vs {
					ss = append(ss, s.(string))
				}
				labels = append(labels, ss...)
			} else {
				vs := v.(string)
				nodeNameFilterName = vs
			}
		}
		// Filter by node name
		if nodeNameFilterName != "" {
			for _, node := range nodeList {
				if strings.Contains(node.String("metadata", "name"), nodeNameFilterName) || strings.Contains(node.StringSlice("metadata", "fields")[5], nodeNameFilterName) {
					nodes = append(nodes, node)
				}
			}
			nodeList = nodes
			nodes = make([]data.Object, 0)
		}
		if len(labels) != 0 {
			for _, node := range nodeList {
			NEXT:
				for _, l := range labels {
					for k, v := range node.Map("metadata", "labels") {
						nl := k + "=" + v.(string)
						if strings.Contains(nl, l) {
							nodes = append(nodes, node)
							break NEXT
						}
					}
				}
			}
		} else {
			nodes = nodeList
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
