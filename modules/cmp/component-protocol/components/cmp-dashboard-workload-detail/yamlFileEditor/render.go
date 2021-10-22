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

package yamlFileEditor

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/pkg/k8sclient"
)

func init() {
	base.InitProviderWithCreator("cmp-dashboard-workload-detail", "yamlFileEditor", func() servicehub.Provider {
		return &ComponentYamlFileEditor{}
	})
}

var steveServer cmp.SteveServer

func (f *ComponentYamlFileEditor) Init(ctx servicehub.Context) error {
	server, ok := ctx.Service("cmp").(cmp.SteveServer)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a steveServer")
	}
	steveServer = server
	return f.DefaultProvider.Init(ctx)
}

func (f *ComponentYamlFileEditor) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if _, ok := (*gs)["deleted"]; ok {
		delete(*gs, "deleted")
		return nil
	}
	f.InitComponent(ctx)
	if err := f.GenComponentState(component); err != nil {
		return errors.Errorf("failed to gen yamlFileEditor component state, %v", err)
	}

	switch event.Operation {
	case cptype.RenderingOperation:
		if err := f.RenderFile(); err != nil {
			return errors.Errorf("failed to render yaml file, %v", err)
		}
	case "submit":
		if err := f.UpdateWorkload(); err != nil {
			return errors.Errorf("failed to update workload, %v", err)
		}
		delete(*gs, "drawerOpen")
	}
	f.SetComponentValue()
	f.Transfer(component)
	return nil
}

func (f *ComponentYamlFileEditor) InitComponent(ctx context.Context) {
	f.ctx = ctx
	sdk := cputil.SDK(ctx)
	f.sdk = sdk
	f.server = steveServer
}

func (f *ComponentYamlFileEditor) GenComponentState(component *cptype.Component) error {
	if component == nil || component.State == nil {
		return nil
	}
	var state State
	data, err := json.Marshal(component.State)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}
	f.State = state
	f.Transfer(component)
	return nil
}

func (f *ComponentYamlFileEditor) RenderFile() error {
	splits := strings.Split(f.State.WorkloadID, "_")
	if len(splits) != 3 {
		return errors.Errorf("invalid workload id: %s", f.State.WorkloadID)
	}

	kind, namespace, name := splits[0], splits[1], splits[2]
	client, err := k8sclient.New(f.State.ClusterName)
	if err != nil {
		return err
	}

	var workload interface{}
	switch kind {
	case string(apistructs.K8SDeployment):
		deploy, err := client.ClientSet.AppsV1().Deployments(namespace).Get(f.ctx, name, v1.GetOptions{})
		if err != nil {
			return errors.Errorf("failed to get deployment, %v", err)
		}
		workload = deploy
	case string(apistructs.K8SStatefulSet):
		sts, err := client.ClientSet.AppsV1().StatefulSets(namespace).Get(f.ctx, name, v1.GetOptions{})
		if err != nil {
			return errors.Errorf("failed to get statefulSet, %v", err)
		}
		workload = sts
	case string(apistructs.K8SDaemonSet):
		ds, err := client.ClientSet.AppsV1().DaemonSets(namespace).Get(f.ctx, name, v1.GetOptions{})
		if err != nil {
			return errors.Errorf("failed to get daemonSet, %v", err)
		}
		workload = ds
	case string(apistructs.K8SJob):
		job, err := client.ClientSet.BatchV1().Jobs(namespace).Get(f.ctx, name, v1.GetOptions{})
		if err != nil {
			return errors.Errorf("failed to get job, %v", err)
		}
		workload = job
	case string(apistructs.K8SCronJob):
		cj, err := client.ClientSet.BatchV1().CronJobs(namespace).Get(f.ctx, name, v1.GetOptions{})
		if err != nil {
			return errors.Errorf("failed to get deployment, %v", err)
		}
		workload = cj
	default:
		return errors.Errorf("invalid workload kind %s", kind)
	}

	data, err := json.Marshal(workload)
	if err != nil {
		return err
	}

	yamlData, err := yaml.JSONToYAML(data)
	if err != nil {
		return err
	}

	f.State.Value = string(yamlData)
	return nil
}

func (f *ComponentYamlFileEditor) UpdateWorkload() error {
	splits := strings.Split(f.State.WorkloadID, "_")
	if len(splits) != 3 {
		return errors.Errorf("invalid workload id: %s", f.State.WorkloadID)
	}
	kind, namespace, name := splits[0], splits[1], splits[2]

	jsonData, err := yaml.YAMLToJSON([]byte(f.State.Value))
	if err != nil {
		return errors.Errorf("failed to convert yaml to json, %v", err)
	}
	var workload map[string]interface{}
	if err = json.Unmarshal(jsonData, &workload); err != nil {
		return errors.Errorf("failed to unmarshal workload, %v", err)
	}

	req := &apistructs.SteveRequest{
		UserID:      f.sdk.Identity.UserID,
		OrgID:       f.sdk.Identity.OrgID,
		Type:        apistructs.K8SResType(kind),
		ClusterName: f.State.ClusterName,
		Name:        name,
		Namespace:   namespace,
		Obj:         workload,
	}

	_, err = f.server.UpdateSteveResource(f.ctx, req)
	return err
}

func (f *ComponentYamlFileEditor) SetComponentValue() {
	f.Props.Bordered = true
	f.Props.FileValidate = []string{"not-empty", "yaml"}
	f.Props.MinLines = 22
	f.Operations = map[string]interface{}{
		"submit": Operation{
			Key:    "submit",
			Reload: true,
		},
	}
}

func (f *ComponentYamlFileEditor) Transfer(component *cptype.Component) {
	component.Props = f.Props
	component.State = map[string]interface{}{
		"clusterName": f.State.ClusterName,
		"workloadId":  f.State.WorkloadID,
		"value":       f.State.Value,
	}
	component.Operations = f.Operations
}
