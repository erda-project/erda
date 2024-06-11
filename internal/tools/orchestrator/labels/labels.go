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

	LabelErdaCloudTenantId = "monitor.erda.cloud/tenant-id"

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
	LabelAddonErdaCloudId    = "addon.erda.cloud/id"
	LabelAddonErdaCloudScope = "addon.erda.cloud/scope"
	LabelAddonErdaCloudName  = "addon.erda.cloud/name"
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

func MergeAddonCoreErdaLabels(target map[string]string, source map[string]string) {
	for core, dice := range labelMappings {
		if v, exist := source[dice]; exist {
			target[core] = v
		}
	}
}

func SetAddonErdaLabels(labels map[string]string, ins *dbclient.AddonInstance) {
	labels[LabelAddonErdaCloudId] = ins.ID
	labels[LabelAddonErdaCloudScope] = ins.ShareScope
	labels[LabelAddonErdaCloudName] = ins.AddonName
}

func SetCoreErdaLabels(sg *apistructs.ServiceGroup, service *apistructs.Service, labels map[string]string) error {
	if labels == nil {
		return nil
	}
	if service != nil {
		for coreLabel, diceLabel := range labelMappings {
			if value, exists := service.Labels[diceLabel]; exists {
				labels[coreLabel] = value
			}
		}
		if publicHost, exists := service.Env[ServiceEnvPublicHost]; exists {
			var publicHostMap map[string]string
			if err := json.Unmarshal([]byte(publicHost), &publicHostMap); err != nil {
				return err
			}
			if terminusKey, exists := publicHostMap[PublicHostTerminusKey]; exists {
				labels[LabelErdaCloudTenantId] = terminusKey
			}
		}
	}

	if sgId, exists := labels[LabelServiceGroupId]; exists {
		labels[LabelCoreErdaCloudServiceGroupId] = sgId
	}

	if sg != nil && sg.ID != "" {
		labels[LabelCoreErdaCloudServiceGroupId] = sg.ID
	}
	return nil
}
