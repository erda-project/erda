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

package podFilter

import (
	"context"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-pods/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-pods/common/filter"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

var (
	inputSc = filter.StateCondition{
		Key:         "q",
		Label:       "标题",
		Placeholder: "请输入关键字查询",
		Type:        "input",
		Fixed:       true,
	}
	namespaceSc = filter.StateCondition{
		Key:   "namespace",
		Label: "命名空间",
		Type:  "select",
		Fixed: true,
	}
	statusSc = filter.StateCondition{
		Key:   "status",
		Label: "状态",
		Type:  "select",
		Fixed: true,
	}
	props = filter.Props{
		Delay: 1000,
	}
	ops = map[string]interface{}{
		apistructs.CMPDashboardFilterOperationKey.String(): filter.Operation{
			Reload: true,
			Key:    "clusterFilter",
		},
	}
	state = filter.State{
		Conditions:    []filter.StateCondition{inputSc, namespaceSc, statusSc},
		IsFirstFilter: false,
	}
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

func (i *NodeFilter) SetComponentValue() {
	i.Props = props
	i.Operations = ops
	i.State = state
	i.Type = "ContractiveFilter"
}

// RenderProtocol 渲染组件
func (i *NodeFilter) RenderProtocol(c *apistructs.Component) error {
	if err := common.Transfer(i.State, &c.State); err != nil {
		return err
	}
	c.Props = i.Props
	c.Operations = i.Operations
	return nil
}

func (i *NodeFilter) Render(ctx context.Context, c *apistructs.Component, s apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	var (
		namespaceLabels, statusLabels []filter.Options
		err                           error
	)
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err = i.SetCtxBundle(bdl); err != nil {
		return err
	}
	if err = common.Transfer(c.State, &i.State); err != nil {
		return err
	}
	switch event.Operation {
	case apistructs.InitializeOperation:
		namespaceLabels, statusLabels, err = i.getFilterOptions()
		if err != nil {
			return err
		}
		i.State.Conditions[1].Options = namespaceLabels
		i.State.Conditions[2].Options = statusLabels
		i.SetComponentValue()
	default:
		logrus.Warnf("operation [%s] not support, scenario:%v, event:%v", event.Operation, s, event)
	}
	return i.RenderProtocol(c)
}

func (i *NodeFilter) getFilterOptions() ([]filter.Options, []filter.Options, error) {
	var (
		err         error
		clusterName string
		resp        *apistructs.SteveCollection
		podList     []apistructs.SteveResource
		namespaces  []filter.Options
		statuses    []filter.Options
		k8sPod v1.Pod
	)
	// Get all pods by cluster name
	clusterName = ""
	nodeReq := &apistructs.SteveRequest{}
	nodeReq.Name = clusterName
	nodeReq.ClusterName = clusterName
	resp, err = bdl.Bdl.ListSteveResource(nodeReq)
	if err != nil {
		return nil, nil, err
	}
	podList = resp.Data
	namespaceSet := map[string]struct{}{}
	statusSet := map[string]struct{}{}
	for _, pod := range podList {
		namespaceSet[pod.Metadata.Namespace] = struct{}{}
		err := common.Transfer(pod, &k8sPod)
		if err != nil {
			return nil, nil, err
		}
		statusSet[string(k8sPod.Status.Phase)] = struct{}{}
	}
	for k := range statusSet {
		statuses = append(statuses, filter.Options{
			Label:    k,
			Value:    k,
		})
	}
	for k := range namespaceSet {
		namespaces = append(namespaces, filter.Options{
			Label:    k,
			Value:    k,
		})
	}
	return statuses, namespaces,nil
}

func RenderCreator() protocol.CompRender {
	return &NodeFilter{}
}
