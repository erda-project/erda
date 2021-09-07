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

package infoDetail

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rancher/wrangler/pkg/data"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (infoDetail *InfoDetail) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	infoDetail.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	infoDetail.Ctx = ctx
	infoDetail.SDK = cputil.SDK(ctx)
	var node data.Object
	if (*gs)["node"] == nil {
		req := apistructs.SteveRequest{}
		req.ClusterName = infoDetail.State.ClusterName
		req.OrgID = infoDetail.SDK.Identity.OrgID
		req.UserID = infoDetail.SDK.Identity.UserID
		req.Type = apistructs.K8SNode
		req.Name = infoDetail.State.NodeID
		resp, err := infoDetail.CtxBdl.GetSteveResource(&req)
		if err != nil {
			return err
		}
		(*gs)["node"] = resp
	} else {
		node = (*gs)["node"].(data.Object)
	}
	infoDetail.Props = infoDetail.getProps(node)
	infoDetail.setOperation(node.String("id"))
	d := Data{}
	d.Os = node.String("status", "nodeInfo", "osImage")
	d.Version = node.String("status", "nodeInfo", "kubeletVersion")
	d.ContainerRuntime = node.StringSlice("metadata", "fields")[9]
	d.NodeIp = infoDetail.getIp(node)
	d.PodNum = node.String("status", "capacity", "pods")
	d.Tags = infoDetail.getTags(node)
	d.Annotation = infoDetail.getAnnotations(node)
	t, err := infoDetail.parseTime(node)
	if err != nil {
		return err
	}
	d.Survive = t
	c.Props = infoDetail.Props
	c.Data = map[string]interface{}{"data": d}
	return nil
}

func (infoDetail *InfoDetail) parseTime(node data.Object) (string, error) {
	t, err := time.Parse("2006-01-02T15:04:05Z", node.String("metadata", "creationTimestamp"))
	if err != nil {
		return "", err
	}
	return time.Now().Sub(t).String(), nil
}

func (infoDetail *InfoDetail) setOperation(nodeId string) {
	for _, f := range infoDetail.Props.Fields {
		if f.ValueKey == "tag" {
			op := f.Operations["add"]
			op.Command.CommandState.FormData.RecordId = nodeId
			op = f.Operations["delete"]
			op.Meta["RecordId"] = nodeId
			return
		}
	}
}

func (infoDetail *InfoDetail) getIp(node data.Object) string {
	for _, addr := range node.Slice("status", "addresses") {
		if addr.String("type") == "InternalIP" {
			return addr.String("address")
		}
	}
	return ""
}
func GetLabelGroup(label string) string {
	ls := []string{
		"dev", "test", "staging", "prod", "stateful", "stateless", "packJob", "cluster-service", "mono", "cordon", "drain", "platform",
	}
	groups := make(map[string]string)
	groups["dev"] = "env"
	groups["test"] = "env"
	groups["staging"] = "env"
	groups["prod"] = "env"

	groups["stateful"] = "service"
	groups["stateless"] = "service"

	groups["packJob"] = "packjob"

	groups["cluster-service"] = "other"
	groups["mono"] = "other"
	groups["cordon"] = "other"
	groups["drain"] = "other"
	groups["platform"] = "other"

	for _, l := range ls {
		if strings.Contains(label, l) {
			return groups[l]
		}
	}
	return "custom"
}

func (infoDetail *InfoDetail) getTags(node data.Object) []Field {
	//groups := filter.GetGroups()
	tag := make([]Field, 0)
	for k, v := range node.Map("metadata", "labels") {
		g := GetLabelGroup(k)
		tag = append(tag, Field{
			Label: fmt.Sprintf("%s:%s", k, v),
			Group: g,
		})
	}
	return tag
}

func (infoDetail *InfoDetail) getAnnotations(node data.Object) []Field {
	desc := make([]Field, 0)
	for k, v := range node.Map("metadata", "annotations") {
		desc = append(desc, Field{
			Label: fmt.Sprintf("%s:%s", k, v),
		})
	}
	return desc
}

func (infoDetail *InfoDetail) getProps(node data.Object) Props {
	return Props{
		ColumnNum: 4,
		Fields: []Field{
			{Label: infoDetail.SDK.I18n("survive"), ValueKey: "survive"},
			{Label: infoDetail.SDK.I18n("nodeIp"), ValueKey: "nodeIp"},
			{Label: infoDetail.SDK.I18n("version"), ValueKey: "version"},
			{Label: infoDetail.SDK.I18n("os"), ValueKey: "os"},
			{Label: infoDetail.SDK.I18n("containerRuntime"), ValueKey: "containerRuntime"},
			{Label: infoDetail.SDK.I18n("podNum"), ValueKey: "podNum"},
			{Label: infoDetail.SDK.I18n("tag"), ValueKey: "tag", RenderType: "tagsRow", SpaceNum: 2, Operations: map[string]Operation{
				"add": {
					Key:    "addLabel",
					Reload: false,
					Command: Command{
						Key:    "set",
						Target: "addLabelModal",
						CommandState: CommandState{
							Visible:  true,
							FormData: FormData{RecordId: node.String("id")},
						},
					},
				},
				"delete": {
					Key:      "deleteLabel",
					Reload:   true,
					FillMeta: "deleteData",
					Meta: map[string]interface{}{
						"RecordId":   node.String("id"),
						"DeleteData": map[string]string{"label": ""},
					},
				}}},
			{Label: infoDetail.SDK.I18n("annotation"), ValueKey: "annotation", SpaceNum: 2, RenderType: "tagsRow"},
		},
	}
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodeDetail", "infoDetail", func() servicehub.Provider {
		return &InfoDetail{}
	})
}
