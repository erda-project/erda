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

package addWorkloadFileEditor

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/appscode/go/encoding/yaml"
	"github.com/pkg/errors"
	"github.com/rancher/wrangler/pkg/data"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("cmp-dashboard-workloads-list", "addWorkloadFileEditor", func() servicehub.Provider {
		return &ComponentAddWorkloadFileEditor{}
	})
}

var steveServer cmp.SteveServer

func (e *ComponentAddWorkloadFileEditor) Init(ctx servicehub.Context) error {
	server, ok := ctx.Service("cmp").(cmp.SteveServer)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a steveServer")
	}
	steveServer = server
	return e.DefaultProvider.Init(ctx)
}

func (e *ComponentAddWorkloadFileEditor) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if _, ok := (*gs)["renderByFilter"]; ok {
		delete(*gs, "renderByFilter")
		return nil
	}

	e.InitComponent(ctx)
	if err := e.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen addWorkloadFileEditor component state, %v", err)
	}

	switch event.Operation {
	case cptype.RenderingOperation:
		if err := e.RenderFile(); err != nil {
			return err
		}
	case "submit":
		if err := e.CreateWorkload(); err != nil {
			return errors.Errorf("failed to create workload, %v", err)
		}
		delete(*gs, "drawerOpen")
	}
	e.SetComponentValue()
	e.Transfer(component)
	return nil
}

func (e *ComponentAddWorkloadFileEditor) InitComponent(ctx context.Context) {
	sdk := cputil.SDK(ctx)
	e.sdk = sdk
	e.ctx = ctx
	e.server = steveServer
}

func (e *ComponentAddWorkloadFileEditor) GenComponentState(component *cptype.Component) error {
	if component == nil || component.State == nil {
		return nil
	}
	var state State
	jsonData, err := json.Marshal(component.State)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(jsonData, &state); err != nil {
		return err
	}
	e.State = state
	return nil
}

func (e *ComponentAddWorkloadFileEditor) SetComponentValue() {
	e.Props.Bordered = true
	e.Props.FileValidate = []string{"not-empty", "yaml"}
	e.Props.MinLines = 22
	e.Operations = map[string]interface{}{
		"submit": Operation{
			Key:    "submit",
			Reload: true,
		},
	}
}

func (e *ComponentAddWorkloadFileEditor) RenderFile() error {
	e.State.Value = workloadTemplates[e.State.WorkloadKind]
	return nil
}

func (e *ComponentAddWorkloadFileEditor) CreateWorkload() error {
	var workload data.Object
	jsonData, err := yaml.ToJSON([]byte(e.State.Value))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsonData, &workload); err != nil {
		return err
	}

	if workload.String("metadata", "namespace") == "" {
		workload.SetNested(e.State.Values.Namespace, "metadata", "namespace")
	}

	req := &apistructs.SteveRequest{
		NoAuthentication: false,
		UserID:           e.sdk.Identity.UserID,
		OrgID:            e.sdk.Identity.OrgID,
		Type:             apistructs.K8SResType(e.State.WorkloadKind),
		ClusterName:      e.State.ClusterName,
		Obj:              workload,
	}

	if _, err := e.server.CreateSteveResource(e.ctx, req); err != nil {
		return err
	}
	return nil
}

func (e *ComponentAddWorkloadFileEditor) Transfer(component *cptype.Component) {
	component.Props = e.Props
	component.State = map[string]interface{}{
		"clusterName":  e.State.ClusterName,
		"workloadKind": e.State.WorkloadKind,
		"values":       e.State.Values,
		"value":        e.State.Value,
	}
	component.Operations = e.Operations
}
