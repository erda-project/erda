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
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"
	"strings"
	"time"
)

func (infoDetail *InfoDetail) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	infoDetail.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	infoDetail.Ctx = ctx
	infoDetail.SDK = cputil.SDK(ctx)
	node := (*gs)["node"].(data.Object)
	infoDetail.Props = infoDetail.getProps()
	infoDetail.setOperation(node.String("id"))
	d := Data{}
	d.Os = node.String("nodeInfo", "osImage")
	d.Version = node.String("nodeInfo", "kubeletVersion")
	d.ContainerRuntime = node.StringSlice("metadata", "fields")[9]
	d.NodeIp = infoDetail.getIp(node)
	d.PodNum = node.String("status", "capacity", "pods")
	d.Tags = infoDetail.getTags(node)
	d.Desc = infoDetail.getDesc(node)
	t, err := infoDetail.parseTime(node)
	if err != nil {
		return err
	}
	d.Survive = t
	c.Props = infoDetail.Props
	c.Data["data"] = d
	return nil
}

func (infoDetail *InfoDetail) parseTime(node data.Object) (string, error) {
	t, err := time.Parse("2006-01-02 15:04:05", node.String("metadata", "creationTimestamp"))
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

func (infoDetail *InfoDetail) getTags(node data.Object) []Field {
	//groups := filter.GetGroups()
	tag := make([]Field, 0)
	for _, l := range node.StringSlice("metadata", "labels") {
		ls := strings.Split(l, ":")
		//for k, _ := range groups {
		//	if strings.Contains(ls[0], k) {
		tag = append(tag, Field{
			Label: ls[0],
			//Group: k,
		})
		//}
		//}
	}
	return tag
}

func (infoDetail *InfoDetail) getDesc(node data.Object) []Field {
	desc := make([]Field, 0)
	for _, l := range node.StringSlice("metadata", "annotations") {
		desc = append(desc, Field{
			Label: l,
		})
	}
	return desc
}

func (infoDetail *InfoDetail) getProps() Props {
	return Props{
		ColumnNum: 2,
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
							FormData: FormData{RecordId: ""},
						},
					},
				},
				"delete": {
					Key:      "deleteLabel",
					Reload:   true,
					FillMeta: "deleteData",
					Meta: map[string]interface{}{
						"RecordId":   "",
						"DeleteData": map[string]string{"label": ""},
					},
				}}},
			{Label: infoDetail.SDK.I18n("survive"), ValueKey: "createCost"},
		},
	}
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodeDetail", "infoDetail", func() servicehub.Provider {
		return &InfoDetail{}
	})
}
