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

package workloadInfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rancher/wrangler/pkg/data"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	workloaddetail "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-workload-detail"
)

func RenderCreator() protocol.CompRender {
	return &ComponentWorkloadInfo{}
}

func (i *ComponentWorkloadInfo) Render(ctx context.Context, component *apistructs.Component, _ apistructs.ComponentProtocolScenario,
	event apistructs.ComponentEvent, _ *apistructs.GlobalStateData) error {
	if err := i.SetCtxBundle(ctx); err != nil {
		return fmt.Errorf("failed to set workloadInfo component ctx bundle, %v", err)
	}
	if err := i.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadInfo component state, %v", err)
	}
	if err := i.SetComponentValue(); err != nil {
		return fmt.Errorf("failed to set workloadInfo component value, %v", err)
	}
	return nil
}

func (i *ComponentWorkloadInfo) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil {
		return errors.New("bundle in context can not be empty")
	}
	i.ctxBdl = bdl
	return nil
}

func (i *ComponentWorkloadInfo) GenComponentState(component *apistructs.Component) error {
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
	i.State = state
	return nil
}

func (i *ComponentWorkloadInfo) SetComponentValue() error {
	kind, namespace, name, err := workloaddetail.ParseWorkloadID(i.State.WorkloadID)
	if err != nil {
		return err
	}
	userID := i.ctxBdl.Identity.UserID
	orgID := i.ctxBdl.Identity.OrgID

	req := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        kind,
		ClusterName: i.State.ClusterName,
		Name:        name,
		Namespace:   namespace,
	}
	obj, err := i.ctxBdl.Bdl.GetSteveResource(&req)
	if err != nil {
		return err
	}

	age, images, err := getWorkloadAgeAndImage(obj)
	if err != nil {
		return err
	}

	labels := obj.Map("metadata", "labels")
	labelTag := make([]Tag, 0)
	for key, value := range labels {
		labelTag = append(labelTag, Tag{Label: fmt.Sprintf("%s=%v", key, value)})
	}
	annotations := obj.Map("metadata", "annotations")
	annotationTag := make([]Tag, 0)
	for key, value := range annotations {
		annotationTag = append(annotationTag, Tag{Label: fmt.Sprintf("%s=%v", key, value)})
	}
	data := DataInData{
		Namespace:   namespace,
		Age:         age,
		Images:      images,
		Labels:      labelTag,
		Annotations: annotationTag,
	}
	i.Data.Data = data

	i.Props.ColumnNum = 4
	i.Props.Fields = []Field{
		{
			Label:    "Namespace",
			ValueKey: "namespace",
		},
		{
			Label:    "Age",
			ValueKey: "age",
		},
		{
			Label:      "Images",
			ValueKey:   "images",
			RenderType: "copyText",
		},
		{
			Label:      "Labels",
			ValueKey:   "labels",
			RenderType: "tagsRow",
			SpaceNum:   2,
		},
		{
			Label:      "Annotations",
			ValueKey:   "annotations",
			RenderType: "tagsRow",
			SpaceNum:   2,
		},
	}
	return nil
}

func getWorkloadAgeAndImage(obj data.Object) (string, string, error) {
	kind := obj.String("kind")
	fields := obj.StringSlice("metadata", "fields")

	switch kind {
	case "Deployment":
		if len(fields) != 8 {
			return "", "", fmt.Errorf("deployment %s has invalid fields length", obj.String("metadata", "name"))
		}
		return fields[4], fields[6], nil
	case "DaemonSet":
		if len(fields) != 11 {
			return "", "", fmt.Errorf("daemonset %s has invalid fields length", obj.String("metadata", "name"))
		}
		return fields[7], fields[9], nil
	case "StatefulSet":
		if len(fields) != 5 {
			return "", "", fmt.Errorf("statefulSet %s has invalid fields length", obj.String("metadata", "name"))
		}
		return fields[2], fields[4], nil
	case "Job":
		if len(fields) != 7 {
			return "", "", fmt.Errorf("job %s has invalid fields length", obj.String("metadata", "name"))
		}
		return fields[3], fields[5], nil
	case "CronJob":
		if len(fields) != 9 {
			return "", "", fmt.Errorf("cronJob %s has invalid fields length", obj.String("metadata", "name"))
		}
		return fields[5], fields[7], nil
	default:
		return "", "", fmt.Errorf("invalid workload kind: %s", kind)
	}
}
