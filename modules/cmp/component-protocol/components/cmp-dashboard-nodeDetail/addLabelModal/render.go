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

package AddLabelModal

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (alm *AddLabelModal) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	alm.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	alm.Ctx = ctx
	alm.SDK = cputil.SDK(ctx)
	alm.GetState()
	err := common.Transfer(c.State, &alm.State)
	if err != nil {
		return err
	}
	alm.getProps()
	alm.GetOperations()
	clusterNameIter := alm.SDK.InParams["clusterName"]
	if clusterNameIter == nil {
		return errors.New("clusterName not found")
	}
	switch event.Operation {
	case common.CMPDashboardSubmit:
		c.Props = alm.Props
		req := &apistructs.SteveRequest{
			UserID:      alm.SDK.Identity.UserID,
			OrgID:       alm.SDK.Identity.OrgID,
			Type:        apistructs.K8SNode,
			Name:        alm.State.FormData["recordId"],
			ClusterName: clusterNameIter.(string),
		}
		labelGroup := alm.State.FormData["labelGroup"]
		labelName := alm.State.FormData[labelGroup]
		strs := strings.Split(labelName, "=")
		if len(strs) == 1 {
			return errors.New("label format illegal, contact key and value with '=' is required")
		}
		err := alm.CtxBdl.LabelNode(req, map[string]string{strs[0]: strs[1]})
		if err != nil {
			return err
		}
	}
	delete(*gs, "nodes")
	return alm.SetComponentValue(c)
}

// SetComponentValue mapping properties to Component
func (alm *AddLabelModal) SetComponentValue(c *cptype.Component) error {
	var (
		err error
	)
	if err = common.Transfer(alm.State, &c.State); err != nil {
		return err
	}
	if err = common.Transfer(alm.Props, &c.Props); err != nil {
		return err
	}
	if err = common.Transfer(alm.Operations, &c.Operations); err != nil {
		return err
	}
	return nil
}

func (alm *AddLabelModal) getProps() {
	fields := []Fields{
		{
			Label:     alm.SDK.I18n("category"),
			Component: "select",
			Required:  true,
			Key:       "labelGroup",
			ComponentProps: ComponentProps{
				Options: []Option{
					{
						Name:  alm.SDK.I18n("platform"),
						Value: "platform",
					},
					{
						Name:  alm.SDK.I18n("env"),
						Value: "environment",
					},
					{
						Name:  alm.SDK.I18n("service"),
						Value: "service",
					},
					{
						Name:  alm.SDK.I18n("job"),
						Value: "job",
					},
					{
						Name:  alm.SDK.I18n("other"),
						Value: "other",
					},
					{
						Name:  alm.SDK.I18n("custom"),
						Value: "custom",
					},
				},
			},
		},
		{
			Key:            "platform",
			ComponentProps: ComponentProps{Options: []Option{{Name: "platform", Value: "platform=true"}}},
			Label:          alm.SDK.I18n("label"),
			Component:      "select",
			Required:       true,
			RemoveWhen: [][]RemoveWhen{
				{
					{
						Field: "labelGroup", Operator: "!=", Value: "platform",
					},
				},
			},
		},
		{
			Key:            "environment",
			ComponentProps: ComponentProps{Options: []Option{{Name: "workspace-dev", Value: "workspace-dev=true"}, {Name: "workspace-tes", Value: "workspace-test=true"}, {Name: "workspace-staging", Value: "workspace-staging=true"}, {Name: "workspace-prod", Value: "workspace-prod=true"}}},
			Label:          alm.SDK.I18n("label"),
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
			ComponentProps: ComponentProps{Options: []Option{{Name: "stateful-service", Value: "stateful-service=true"}, {Name: "stateless-service", Value: "stateless-service=true"}, {Name: "location-cluster-service", Value: "location-cluster-service=true"}}},
			Label:          alm.SDK.I18n("label"),
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
			ComponentProps: ComponentProps{Options: []Option{{Name: "pack-job", Value: "pack-job=true"}, {Name: "bigdata-job", Value: "bigdata-job=true"}, {Name: "job", Value: "job=true"}}},
			Label:          alm.SDK.I18n("label"),
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
			ComponentProps: ComponentProps{Options: []Option{{Name: "locked", Value: "locked=true"}, {Name: "topology-zone", Value: "topology-zone=true"}}},
			Label:          alm.SDK.I18n("label"),
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
			Label:     alm.SDK.I18n("label"),
			Component: "input",
			Required:  true,
			Rules: Rules{
				Msg:     alm.SDK.I18n("regex") + "label=true",
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

func (alm *AddLabelModal) GetOperations() {
	alm.Operations = map[string]Operations{
		"submit": {
			Key:    "submit",
			Reload: true,
		},
	}
}

func (alm *AddLabelModal) GetState() {
	alm.State = State{
		Visible:  false,
		FormData: nil,
	}
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-nodeDetail", "addLabelModal", func() servicehub.Provider {
		return &AddLabelModal{}
	})
}
