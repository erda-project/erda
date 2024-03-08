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

package types

import (
	"regexp"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
)

const (
	// to be deprecated.
	// ADDON_GROUPS -> SERVICE_GROUPS
	// ADDON_GROUP_ID -> SERVICE_GROUP_ID
	GroupNum = "ADDON_GROUPS"
	GroupID  = "ADDON_GROUP_ID"

	GroupNum2 = "SERVICE_GROUPS"
	GroupID2  = "SERVICE_GROUP_ID"

	// local volume storageclass name
	LocalStorage = "dice-local-volume"

	// default sa
	DefaultServiceAccountName = "default"
	DiceWorkSpace             = "DICE_WORKSPACE"
)

var EnvReg = regexp.MustCompile(`\$\{([^}]+?)\}`)

type StatefulsetInfo struct {
	Sg          *apistructs.ServiceGroup
	Namespace   string
	Envs        map[string]string
	Annotations map[string]string
}

// OneGroupInfo Returns information about the statefulset corresponding to the group
type OneGroupInfo struct {
	Sg  *apistructs.ServiceGroup
	Sts *appsv1.StatefulSet
}

const (
	ServiceType        = "SERVICE_TYPE"
	StatelessService   = "STATELESS_SERVICE"
	IsStatelessService = "true"
	ServiceAddon       = "ADDONS"
	ServicePerNode     = "per_node"
	ServiceJob         = "JOB"
)

type (
	PatchStruct struct {
		Spec Spec `json:"spec"`
	}

	Spec struct {
		Template PodTemplateSpec `json:"template"`
	}

	PodTemplateSpec struct {
		Spec PodSpec `json:"spec"`
	}

	PodSpec struct {
		Containers []corev1.Container `json:"containers"`
	}
)
