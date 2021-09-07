package PodInfo

import (
	"context"
	"fmt"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-podDetail/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"
	"strings"
)

func (podInfo *PodInfo) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	pod := (*gs)["pod"].(data.Object)
	podInfo.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	podInfo.SDK = cputil.SDK(ctx)
	podInfo.Props = podInfo.getProps(pod)
	surviveTime, err := common.SurviveTime(pod.String("status", "startTime"))
	if err != nil {
		return err
	}
	data := Data{
		Namespace:        pod.String("metadata", "namespace"),
		Survive:          surviveTime,
		Ip:               pod.String("status", "podIP"),
		PodNum:           "",
		Workload:         "workload...",
		Node:             "containerRuntime",
		ContainerRuntime: "",
		Tag:              podInfo.getTags(pod),
		Desc:             podInfo.getDesc(pod),
	}
	c.Props = podInfo.Props
	c.Data["data"] = data
	return nil
}

func (podInfo *PodInfo) getTags(pod data.Object) []Tag {
	tags := make([]Tag, 0)
	for _, s := range pod.StringSlice("metadata", "labels") {
		tags = append(tags, Tag{
			Label: s,
			Group: "",
		})
	}
	return tags
}

func (podInfo *PodInfo) getProps(pod data.Object) Props {
	return Props{
		ColumnNum: 4,
		Fields: []Field{
			{Label: podInfo.SDK.I18n("namespace"), ValueKey: "namespace"},
			{Label: podInfo.SDK.I18n("survive"), ValueKey: "survive"},
			{Label: "Pod" + podInfo.SDK.I18n("ip"), ValueKey: "ip"},
			{Label: podInfo.SDK.I18n("workload"), ValueKey: "workload", Operation: map[string]Operation{
				"click": {
					Key:    "gotoWorkloadDetail",
					Reload: false,
					Command: Command{
						Key: "goto",
						// todo
						Target: "cmpClustersWorkloadDetail",
						State: CommandState{Params: map[string]string{
							"workloadID": podInfo.getWorkloadID(pod),
						}},
						JumpOut: true,
					},
				},
			}},
			{Label: podInfo.SDK.I18n("node"), ValueKey: "node", RenderType: "linkText", Operation: map[string]Operation{
				"click": {
					Key:    "gotoNodeDetail",
					Reload: false,
					Command: Command{
						Key: "goto",
						// todo
						Target: "cmpClustersNodeDetail",
						State: CommandState{Params: map[string]string{
							"nodeId": pod.String("spec", "nodeName"),
						}},
						JumpOut: true,
					},
				},
			}},
			{
				Label:      podInfo.SDK.I18n("tag"),
				ValueKey:   "tag",
				RenderType: "tagsRow",
				SpaceNum:   2,
			},
			{Label: podInfo.SDK.I18n("desc"), ValueKey: "desc", SpaceNum: 2, RenderType: "tagsRow"},
		},
	}
}

func (podInfo *PodInfo) getDesc(pod data.Object) []Desc {
	desces := make([]Desc, 0)
	for _, s := range pod.StringSlice("metadata", "annotations") {
		desces = append(desces, Desc{
			Label: s,
			Group: "",
		})
	}
	return desces
}

func (podInfo *PodInfo) getWorkloadID(pod data.Object) string {
	var name = pod.String("metadata", "ownerReferences", "name")
	if strings.HasPrefix(pod.String("metadata", "ownerReferences", "kind"), "Replica") {
		//"name": "traffic-manager-6458cb9bf9",
		name = name[:strings.LastIndex(name, "-")]
	}
	return fmt.Sprintf("%s_%s_%s", apistructs.K8SCronJob, pod.String("metadata", "namespace"), name)
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "podInfo", func() servicehub.Provider {
		return &PodInfo{
			Type: "Panel",
		}
	})
}
