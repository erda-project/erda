package AddLabelModal

import (
	"context"
	"errors"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodeDetail/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"strconv"
	"strings"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (alm *AddLabelModal) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	alm.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	alm.Ctx = ctx
	alm.SDK = cputil.SDK(ctx)

	alm.Props = Props{
		Title: alm.SDK.I18n("addLabel"),
	}
	if event.Operation == cptype.InitializeOperation {
		alm.getProps(ctx)
		c.Props = alm.Props
	}
	req := &apistructs.SteveRequest{
		UserID:      alm.SDK.Identity.UserID,
		OrgID:       alm.SDK.Identity.OrgID,
		Type:        apistructs.K8SNode,
		ClusterName: alm.SDK.InParams["clusterName"].(string),
		Name:        alm.State.FormData["recordId"],
	}
	labelGroup := alm.State.FormData["labelGroup"]
	labelName := alm.State.FormData[labelGroup]
	switch event.Operation {
	case common.CMPDashboardAddLabel:
		err := alm.CtxBdl.LabelNode(req, map[string]string{labelName: "true"})
		if err != nil {
			return err
		}
	case common.CMPDashboardRemoveLabel:
		err := alm.CtxBdl.UnlabelNode(req, []string{labelName})
		if err != nil {
			return err
		}
	}
	delete(*gs,"nodes")
	return nil
}

func (alm *AddLabelModal) getProps(ctx context.Context) {
	fields := []Fields{
		{
			Label:     cputil.I18n(ctx, "category"),
			Component: "select",
			Required:  true,
			ComponentProps: ComponentProps{
				Options: []Option{
					{
						Name:  cputil.I18n(ctx, "platform"),
						Value: "platform",
					},
					{
						Name:  cputil.I18n(ctx, "environment"),
						Value: "environment",
					},
					{
						Name:  cputil.I18n(ctx, "service"),
						Value: "service",
					},
					{
						Name:  cputil.I18n(ctx, "job"),
						Value: "job",
					},
					{
						Name:  cputil.I18n(ctx, "other"),
						Value: "other",
					},
					{
						Name:  cputil.I18n(ctx, "custom"),
						Value: "custom",
					},
				},
			},
		},
		{
			Key:            "platform",
			ComponentProps: ComponentProps{Options: []Option{{Name: cputil.I18n(ctx, "platform"), Value: "platform"}}},
			Label:          cputil.I18n(ctx, "label"),
			Component:      "select",
			Required:       true,
		},
		{
			Key:            "environment",
			ComponentProps: ComponentProps{Options: []Option{{Name: cputil.I18n(ctx, "workspace-dev"), Value: "workspace-dev"}, {Name: cputil.I18n(ctx, "workspace-test"), Value: "workspace-test"}, {Name: cputil.I18n(ctx, "workspace-staging"), Value: "workspace-staging"}, {Name: cputil.I18n(ctx, "workspace-prod"), Value: "workspace-prod"}}},
			Label:          cputil.I18n(ctx, "label"),
			Component:      "select",
			Required:       true,
			RemoveWhen: [][]RemoveWhen{
				{
					{
						Field: "labelGroup", Operator: "!=", Value: "environment",
					},
				},
			},
		},
		{
			Key:            "service",
			ComponentProps: ComponentProps{Options: []Option{{Name: cputil.I18n(ctx, "stateful-service"), Value: "stateful-service"}, {Name: cputil.I18n(ctx, "stateless-service"), Value: "stateless-service"}, {Name: cputil.I18n(ctx, "location-cluster-service"), Value: "location-cluster-service"}}},
			Label:          cputil.I18n(ctx, "label"),
			Component:      "select",
			Required:       true,
			RemoveWhen: [][]RemoveWhen{
				{
					{
						Field: "labelGroup", Operator: "!=", Value: "service",
					},
				},
			},
		},
		{
			Key:            "job",
			ComponentProps: ComponentProps{Options: []Option{{Name: cputil.I18n(ctx, "pack-job"), Value: "pack-job"}, {Name: cputil.I18n(ctx, "bigdata-job"), Value: "bigdata-job"}, {Name: cputil.I18n(ctx, "job"), Value: "job"}}},
			Label:          cputil.I18n(ctx, "label"),
			Component:      "select",
			Required:       true,
			RemoveWhen: [][]RemoveWhen{
				{
					{
						Field: "labelGroup", Operator: "!=", Value: "job",
					},
				},
			},
		},
		{
			Key:            "other",
			ComponentProps: ComponentProps{Options: []Option{{Name: cputil.I18n(ctx, "locked"), Value: "locked"}, {Name: cputil.I18n(ctx, "topology-zone"), Value: "topology-zone"}}},
			Label:          cputil.I18n(ctx, "label"),
			Component:      "select",
			Required:       true,
			RemoveWhen: [][]RemoveWhen{
				{
					{
						Field: "labelGroup", Operator: "!=", Value: "other",
					},
				},
			},
		},
		{
			Key:       "custom",
			Label:     cputil.I18n(ctx, "label"),
			Component: "input",
			Required:  true,
			Rules: Rules{
				Msg:     "格式:",
				Pattern: "pattern: '/^[.a-z\\u4e00-\\u9fa5A-Z0-9_-\\s]*$/'",
			},
			RemoveWhen: [][]RemoveWhen{
				{
					{
						Field: "labelGroup", Operator: "!=", Value: "custom",
					},
				},
			},
		},
	}
	alm.Props = Props{
		Fields: fields,
		Title:  alm.SDK.I18n("addLabel"),
	}
}

func hasSuffix(name string) (string, bool) {
	suffixes := []string{"-dev", "-staging", "-test", "-prod"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return suffix, true
		}
	}
	return "", false
}

func (alm *AddLabelModal) getDisplayName(name string) (string, error) {
	splits := strings.Split(name, "-")
	if len(splits) != 3 {
		return "", errors.New("invalid name")
	}
	id := splits[1]
	num, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return "", err
	}
	project, err := alm.CtxBdl.GetProject(uint64(num))
	if err != nil {
		return "", err
	}
	return project.DisplayName, nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodes", "addLabelModal", func() servicehub.Provider {
		al := &AddLabelModal{
			Type: "FormModal",
			State: State{
				Visible:  false,
				FormData: nil,
			},
			Operations: map[string]Operations{
				"submit": {
					Key:    "submit",
					Reload: true,
				},
			},
		}
		return al
	})
}
