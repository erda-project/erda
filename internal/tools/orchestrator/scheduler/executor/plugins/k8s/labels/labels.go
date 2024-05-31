package labels

import (
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
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

	ServiceEnvPublicHost = "PUBLIC_HOST"

	PublicHostTerminusKey = "terminusKey"
)

func SetCoreErdaLabels(sg *apistructs.ServiceGroup, service *apistructs.Service, labels map[string]string) error {
	if labels == nil {
		return nil
	}
	if service != nil {
		labels[LabelCoreErdaCloudClusterName] = service.Labels[LabelDiceClusterName]
		labels[LabelCoreErdaCloudOrgId] = service.Labels[LabelDiceOrgId]
		labels[LabelCoreErdaCloudOrgName] = service.Labels[LabelDiceOrgName]
		labels[LabelCoreErdaCloudAppId] = service.Labels[LabelDiceAppId]
		labels[LabelCoreErdaCloudAppName] = service.Labels[LabelDiceAppName]
		labels[LabelCoreErdaCloudProjectId] = service.Labels[LabelDiceProjectId]
		labels[LabelCoreErdaCloudProjectName] = service.Labels[LabelDiceProjectName]
		labels[LabelCoreErdaCloudRuntimeId] = service.Labels[LabelDiceRuntimeId]
		labels[LabelCoreErdaCloudServiceName] = service.Labels[LabelDiceServiceName]
		labels[LabelCoreErdaCloudWorkSpace] = service.Labels[LabelDiceWorkSpace]
		labels[LabelCoreErdaCloudServiceType] = service.Labels[LabelDiceServiceType]
		publicHost := make(map[string]string)
		err := json.Unmarshal([]byte(service.Env[ServiceEnvPublicHost]), &publicHost)
		if err != nil {
			return err
		}
		labels[LabelErdaCloudTenantId] = publicHost[PublicHostTerminusKey]
	}

	if sg != nil {
		labels[LabelCoreErdaCloudServiceGroupId] = sg.ID
	}
	return nil
}
