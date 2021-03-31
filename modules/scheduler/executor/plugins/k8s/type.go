package k8s

import (
	"regexp"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/erda-project/erda/apistructs"
)

const (
	// to be deprecated.
	// ADDON_GROUPS -> SERVICE_GROUPS
	// ADDON_GROUP_ID -> SERVICE_GROUP_ID
	groupNum = "ADDON_GROUPS"
	groupID  = "ADDON_GROUP_ID"

	groupNum2 = "SERVICE_GROUPS"
	groupID2  = "SERVICE_GROUP_ID"

	// aliyun registry
	AliyunRegistry = "aliyun-registry"
	CustomRegistry = "custom-registry"
	// local volume storageclass name
	localStorage = "dice-local-volume"

	// 默认的 sa
	defaultServiceAccountName = "default"
)

var envReg = regexp.MustCompile(`\$\{([^}]+?)\}`)

type StatefulsetInfo struct {
	sg          *apistructs.ServiceGroup
	namespace   string
	envs        map[string]string
	annotations map[string]string
}

// OneGroupInfo 返回该group对应的statefulset的信息
type OneGroupInfo struct {
	sg  *apistructs.ServiceGroup
	sts *appsv1.StatefulSet
}

const (
	ServiceType        = "SERVICE_TYPE"
	StatelessService   = "STATELESS_SERVICE"
	IsStatelessService = "true"
	ServiceAddon       = "ADDONS"
	ServicePerNode     = "per_node"
)
