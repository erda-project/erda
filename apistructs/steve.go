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

package apistructs

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/erda-project/erda/pkg/strutil"
)

const (
	SteveErrorType = "error"
)

// SteveCollection 用于接收steve server返回的集合类型的数据（一次返回多条数据）
type SteveCollection struct {
	Type         string            `json:"type,omitempty"`
	Links        map[string]string `json:"links"`
	CreateTypes  map[string]string `json:"createTypes,omitempty"`
	Actions      map[string]string `json:"actions"`
	ResourceType string            `json:"resourceType"`
	Revision     string            `json:"revision"`
	Pagination   *Pagination       `json:"pagination,omitempty"`
	Continue     string            `json:"continue,omitempty"`
	// steve资源列表
	Data []SteveResource `json:"data"`
}

// SteveError 用于接收steve server返回的错误
type SteveError struct {
	Type    string `json:"type,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Status  int    `json:"status,omitempty"`
}

// SteveResource 用于接收steve server返回的实例类型的数据（一次返回一条数据）
type SteveResource struct {
	K8SResource
	ID    string            `json:"id,omitempty"`
	Type  string            `json:"type,omitempty"`
	Links map[string]string `json:"links"`
}

// K8SResource 为k8s原生资源
type K8SResource struct {
	metav1.TypeMeta
	Metadata metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec     interface{}       `json:"spec,omitempty"`
	Status   interface{}       `json:"status,omitempty"`
}

// Pagination 用于分页查询
type Pagination struct {
	Limit   int    `json:"limit,omitempty"`   // 每页数据条数
	First   string `json:"first,omitempty"`   // 第一页数据链接
	Next    string `json:"next,omitempty"`    // 下一页数据链接
	Partial bool   `json:"partial,omitempty"` // 是否为部分数据
}

type K8SResType string

const (
	K8SPod         K8SResType = "pods"
	K8SNode        K8SResType = "nodes"
	K8SDeployment  K8SResType = "apps.deployments"
	K8SReplicaSet  K8SResType = "apps.replicasets"
	K8SDaemonSet   K8SResType = "apps.daemonsets"
	K8SStatefulSet K8SResType = "apps.statefulsets"
	K8SEvent       K8SResType = "events"
)

// SteveRequest 用于向steve发送get或list请求
type SteveRequest struct {
	Type        K8SResType // 资源类型，必填
	ClusterName string     // 集群名，必填
	Name        string     // 资源名，Get, Delete, Update请求时必填
	Namespace   string     // 命名空间
	// 标签匹配，list时可选
	// 格式"key=value"，或"key in (value1, value2)"，或"key notin (value1, value2)"
	LabelSelector []string

	Obj interface{} // Update, Create请求时使用，obj为k8s原生资源的指针，如*v1.pod, *v1.node
}

func (k *SteveRequest) URLQueryString() map[string][]string {
	query := make(map[string][]string)

	if len(k.LabelSelector) != 0 {
		labels := strutil.Join(k.LabelSelector, ",", true)
		query["labelSelector"] = []string{labels}
	}
	return query
}
