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

package labels

import (
	"encoding/json"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/sirupsen/logrus"
)

// k8s labels
const (
	LabelCoreErdaCloudClusterName    = "core.erda.cloud/cluster-name"
	LabelCoreErdaCloudOrgId          = "core.erda.cloud/org-id"
	LabelCoreErdaCloudOrgName        = "core.erda.cloud/org-name"
	LabelCoreErdaCloudAppId          = "core.erda.cloud/app-id"
	LabelCoreErdaCloudAppName        = "core.erda.cloud/app-name"
	LabelCoreErdaCloudProjectId      = "core.erda.cloud/project-id"
	LabelCoreErdaCloudProjectName    = "core.erda.cloud/project-name"
	LabelCoreErdaCloudRuntimeId      = "core.erda.cloud/runtime-id"
	LabelCoreErdaCloudServiceName    = "core.erda.cloud/service-name"
	LabelCoreErdaCloudWorkSpace      = "core.erda.cloud/workspace"
	LabelCoreErdaCloudServiceType    = "core.erda.cloud/service-type"
	LabelCoreErdaCloudServiceGroupId = "core.erda.cloud/servicegroup-id"

	LabelMonitorErdaCloudTenantId = "monitor.erda.cloud/tenant-id"
	LabelMonitorErdaCloudEnabled  = "monitor.erda.cloud/enabled"
	LabelMonitorErdaCloudExporter = "monitor.erda.cloud/exporter"

	LabelDiceClusterName = "DICE_CLUSTER_NAME"
	LabelDiceOrgId       = "DICE_ORG_ID"
	LabelDiceOrgName     = "DICE_ORG_NAME"
	LabelDiceAppId       = "DICE_APPLICATION_ID"
	LabelDiceAppName     = "DICE_APPLICATION_NAME"
	LabelDiceProjectId   = "DICE_PROJECT_ID"
	LabelDiceProjectName = "DICE_PROJECT_NAME"
	LabelDiceRuntimeId   = "DICE_RUNTIME_ID"
	LabelDiceServiceName = "DICE_SERVICE_NAME"
	LabelDiceWorkSpace   = "DICE_WORKSPACE"
	LabelDiceServiceType = "SERVICE_TYPE"
	LabelServiceGroupId  = "servicegroup-id"

	ServiceEnvPublicHost = "PUBLIC_HOST"

	PublicHostTerminusKey = "terminusKey"
)

const (
	LabelAddonErdaCloudId      = "addon.erda.cloud/id"
	LabelAddonErdaCloudScope   = "addon.erda.cloud/scope"
	LabelAddonErdaCloudType    = "addon.erda.cloud/type"
	LabelAddonErdaCloudVersion = "addon.erda.cloud/version"
)

var labelMappings = map[string]string{
	LabelCoreErdaCloudClusterName: LabelDiceClusterName,
	LabelCoreErdaCloudOrgId:       LabelDiceOrgId,
	LabelCoreErdaCloudOrgName:     LabelDiceOrgName,
	LabelCoreErdaCloudAppId:       LabelDiceAppId,
	LabelCoreErdaCloudAppName:     LabelDiceAppName,
	LabelCoreErdaCloudProjectId:   LabelDiceProjectId,
	LabelCoreErdaCloudProjectName: LabelDiceProjectName,
	LabelCoreErdaCloudRuntimeId:   LabelDiceRuntimeId,
	LabelCoreErdaCloudServiceName: LabelDiceServiceName,
	LabelCoreErdaCloudWorkSpace:   LabelDiceWorkSpace,
	LabelCoreErdaCloudServiceType: LabelDiceServiceType,
}

// AddonLabelSetter Set labels for addon
type AddonLabelSetter struct {
	target   map[string]string
	source   map[string]string
	instance *dbclient.AddonInstance
	params   *apistructs.AddonHandlerCreateItem
}

func NewAddonLabelSetter(target map[string]string, source map[string]string, ins *dbclient.AddonInstance, params *apistructs.AddonHandlerCreateItem) *AddonLabelSetter {
	return &AddonLabelSetter{target: target, source: source, instance: ins, params: params}
}

func (a *AddonLabelSetter) SetCoreErdaLabels() *AddonLabelSetter {
	if a.source == nil || a.target == nil {
		return a
	}
	for core, dice := range labelMappings {
		if v, exist := a.source[dice]; exist {
			a.target[core] = v
		}
	}
	return a
}

func (a *AddonLabelSetter) SetAddonErdaLabels() *AddonLabelSetter {
	if a.instance == nil || a.target == nil {
		return a
	}
	a.target[LabelAddonErdaCloudId] = a.instance.ID
	a.target[LabelAddonErdaCloudScope] = a.instance.ShareScope
	a.target[LabelAddonErdaCloudType] = a.instance.AddonName
	a.target[LabelAddonErdaCloudVersion] = a.instance.Version
	return a
}

func (a *AddonLabelSetter) SetMonitorErdaCloudLabels() *AddonLabelSetter {
	if a.target == nil {
		return a
	}
	if a.params != nil && a.params.TenantId != "" {
		a.target[LabelMonitorErdaCloudTenantId] = a.params.TenantId
	}
	a.target[LabelMonitorErdaCloudEnabled] = "true"
	a.target[LabelMonitorErdaCloudExporter] = "true"
	return nil
}

// RuntimeLabelSetter  Set labels for runtime
type RuntimeLabelSetter struct {
	service *apistructs.Service
	sg      *apistructs.ServiceGroup
	labels  map[string]string
}

func NewRuntimeLabelSetter(service *apistructs.Service, sg *apistructs.ServiceGroup, labels map[string]string) *RuntimeLabelSetter {
	return &RuntimeLabelSetter{service: service, sg: sg, labels: labels}
}

func (r *RuntimeLabelSetter) SetMonitorErdaCloudLabels() *RuntimeLabelSetter {
	if r.labels == nil {
		r.labels = make(map[string]string)
	}
	r.labels[LabelMonitorErdaCloudEnabled] = "true"
	r.labels[LabelMonitorErdaCloudExporter] = "true"

	if r.service == nil {
		return r
	}

	if publicHost, exists := r.service.Env[ServiceEnvPublicHost]; exists {
		var publicHostMap map[string]string
		if err := json.Unmarshal([]byte(publicHost), &publicHostMap); err != nil {
			logrus.Errorf("failed to unmarshal public host map for service: %v, err: %v", r.service.Name, err)
		}
		if terminusKey, exists := publicHostMap[PublicHostTerminusKey]; exists {
			r.labels[LabelMonitorErdaCloudTenantId] = terminusKey
		}
	}
	return r
}

func (r *RuntimeLabelSetter) SetCoreErdaLabels() *RuntimeLabelSetter {
	if r.labels == nil {
		r.labels = make(map[string]string)
	}
	if r.service != nil {
		for coreLabel, diceLabel := range labelMappings {
			if value, exists := r.service.Labels[diceLabel]; exists {
				r.labels[coreLabel] = value
			}
		}
	}

	if sgId, exists := r.labels[LabelServiceGroupId]; exists {
		r.labels[LabelCoreErdaCloudServiceGroupId] = sgId
	}

	if r.sg != nil && r.sg.ID != "" {
		r.labels[LabelCoreErdaCloudServiceGroupId] = r.sg.ID
	}

	return r
}
