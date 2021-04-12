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

package addon

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/erda-project/erda/apistructs"
)

type AddonOperator interface {
	// Determine whether this cluster supports addon operator
	IsSupported() bool
	// Verify the legality of sg converted from diceyml
	Validate(*apistructs.ServiceGroup) error
	// Convert sg to cr. Because of the different definitions of cr, use interface() here
	Convert(*apistructs.ServiceGroup) interface{}
	// the cr converted by Convert in k8s deploying
	Create(interface{}) error
	// Check running status
	Inspect(*apistructs.ServiceGroup) (*apistructs.ServiceGroup, error)

	Remove(*apistructs.ServiceGroup) error

	Update(interface{}) error
}

type K8SUtil interface {
	GetK8SAddr() string
}

type DeploymentUtil interface {
	Create(*appsv1.Deployment) error
	Get(namespace, name string) (*appsv1.Deployment, error)
	List(namespace string, labelSelector map[string]string) (appsv1.DeploymentList, error)
	Delete(namespace, name string) error
}

type StatefulsetUtil interface {
	Create(*appsv1.StatefulSet) error
	Delete(namespace, name string) error
	Get(namespace, name string) (*appsv1.StatefulSet, error)
	List(namespace string) (appsv1.StatefulSetList, error)
}

type DaemonsetUtil interface {
	Create(*appsv1.DaemonSet) error
	Update(*appsv1.DaemonSet) error
	Delete(namespace, name string) error
	Get(namespace, name string) (*appsv1.DaemonSet, error)
	List(namespace string, labelSelector map[string]string) (appsv1.DaemonSetList, error)
}

type ServiceUtil interface {
	List(namespace string) (corev1.ServiceList, error)
}

type NamespaceUtil interface {
	Exists(ns string) error
	Create(ns string, labels map[string]string) error
	Delete(ns string) error
}

type SecretUtil interface {
	Get(ns, name string) (*corev1.Secret, error)
	Create(*corev1.Secret) error
	CreateIfNotExist(secret *corev1.Secret) error
}

type ImageSecretUtil interface {
	//The secret used to pull the mirror under the default namespace is copied to the current ns,
	// Then add this secret to the default sa of this ns
	NewImageSecret(ns string) error
}

type HealthcheckUtil interface {
	NewHealthcheckProbe(*apistructs.Service) *corev1.Probe
}

type PVCUtil interface {
	Create(pvc *corev1.PersistentVolumeClaim) error
}
type OvercommitUtil interface {
	CPUOvercommit(limit float64) float64
	MemoryOvercommit(limit int) int
}
