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
	// 判断这个集群是否支持 addon operator
	IsSupported() bool
	// 验证由 diceyml 转化而来的 sg 的合法性
	Validate(*apistructs.ServiceGroup) error
	// 将 sg 转化为 cr, 由于各个 cr 的定义不同, 这里用 interface{}
	Convert(*apistructs.ServiceGroup) interface{}
	// 将 Convert 所转化的 cr 在 k8s 中部署
	Create(interface{}) error
	// 检查运行状态
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
	// 在 default namespace 下的用来pull镜像的secret复制到当前 ns 中,
	// 再将这个 secret 加到这个 ns 的 default sa 中去
	NewImageSecret(ns string) error
}

type HealthcheckUtil interface {
	NewHealthcheckProbe(*apistructs.Service) *corev1.Probe
}

type PVCUtil interface {
	Create(pvc *corev1.PersistentVolumeClaim) error
}
type OvercommitUtil interface {
	// 5 -> 2.5  == 超卖 2 倍 (单位 核数)
	CPUOvercommit(limit float64) float64
	// 1000 -> 500 == 超卖 2 倍 (单位 Mi)
	MemoryOvercommit(limit int) int
}
