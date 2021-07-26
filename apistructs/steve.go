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

// SteveCollection is a resource collection returned from steve server.
type SteveCollection struct {
	Type         string            `json:"type,omitempty"`
	Links        map[string]string `json:"links"`
	CreateTypes  map[string]string `json:"createTypes,omitempty"`
	Actions      map[string]string `json:"actions"`
	ResourceType string            `json:"resourceType"`
	Revision     string            `json:"revision"`
	Pagination   *Pagination       `json:"pagination,omitempty"`
	Continue     string            `json:"continue,omitempty"`
	// steve resources
	Data []SteveResource `json:"data"`
}

// SteveError is an error returned from steve server.
type SteveError struct {
	Type    string `json:"type,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Status  int    `json:"status,omitempty"`
}

// SteveResource is a steve resource returned from steve server.
type SteveResource struct {
	K8SResource
	ID    string            `json:"id,omitempty"`
	Type  string            `json:"type,omitempty"`
	Links map[string]string `json:"links"`
}

// K8SResource is a original k8s resource.
type K8SResource struct {
	metav1.TypeMeta
	Metadata metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec     interface{}       `json:"spec,omitempty"`
	Status   interface{}       `json:"status,omitempty"`
}

// Pagination used to paging query.
type Pagination struct {
	Limit   int    `json:"limit,omitempty"`   // maximum number of each page
	First   string `json:"first,omitempty"`   // first page link
	Next    string `json:"next,omitempty"`    // next page link
	Partial bool   `json:"partial,omitempty"` // whether partial
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

// SteveRequest used to query steve server by bundle.
type SteveRequest struct {
	UserID      string     // used to authentication, required
	OrgID       string     // used to authentication, required
	Type        K8SResType // type of resource, required
	ClusterName string     // cluster name, required
	Name        string     // name of resource，required when Get, Delete, Update
	Namespace   string     // namespace of resource
	// label selector, optional when list
	// format: "key=value"，or "key in (value1, value2)"，or "key notin (value1, value2)"
	LabelSelector []string
	// required in  Update, Create，obj is a pointer of original k8s resource，like *v1.pod, *v1.node
	Obj interface{}
}

// URLQueryString converts label selectors to url query params.
func (k *SteveRequest) URLQueryString() map[string][]string {
	query := make(map[string][]string)

	if len(k.LabelSelector) != 0 {
		labels := strutil.Join(k.LabelSelector, ",", true)
		query["labelSelector"] = []string{labels}
	}
	return query
}
