// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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

	// default sa
	defaultServiceAccountName = "default"
)

var envReg = regexp.MustCompile(`\$\{([^}]+?)\}`)

type StatefulsetInfo struct {
	sg          *apistructs.ServiceGroup
	namespace   string
	envs        map[string]string
	annotations map[string]string
}

// OneGroupInfo Returns information about the statefulset corresponding to the group
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
