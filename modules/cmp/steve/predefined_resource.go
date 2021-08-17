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

package steve

import (
	"encoding/json"
	"strings"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

var (
	predefinedServiceAccount     = getPredefinedServiceAccount()
	predefinedClusterRole        = getPredefinedClusterRole()
	predefinedClusterRoleBinding = getPredefinedClusterRoleBinding()
)

func getPredefinedServiceAccount() []*corev1.ServiceAccount {
	yamls := strings.Split(ServiceAccountExpression, "\n---\n")
	var sa []*corev1.ServiceAccount
	for _, yml := range yamls {
		var tmp corev1.ServiceAccount
		if err := yamlToObject(yml, &tmp); err != nil {
			panic(err)
		}
		if tmp.Name == "" {
			continue
		}
		sa = append(sa, &tmp)
	}
	return sa
}

func getPredefinedClusterRole() []*rbacv1.ClusterRole {
	yamls := strings.Split(ClusterRoleExpression, "\n---\n")
	var cr []*rbacv1.ClusterRole
	for _, yml := range yamls {
		var tmp rbacv1.ClusterRole
		if err := yamlToObject(yml, &tmp); err != nil {
			panic(err)
		}
		if tmp.Name == "" {
			continue
		}
		cr = append(cr, &tmp)
	}
	return cr
}

func getPredefinedClusterRoleBinding() []*rbacv1.ClusterRoleBinding {
	yamls := strings.Split(ClusterRoleBindingExpression, "\n---\n")
	var crb []*rbacv1.ClusterRoleBinding
	for _, yml := range yamls {
		var tmp rbacv1.ClusterRoleBinding
		if err := yamlToObject(yml, &tmp); err != nil {
			panic(err)
		}
		if tmp.Name == "" {
			continue
		}
		crb = append(crb, &tmp)
	}
	return crb
}

func yamlToObject(yml string, obj interface{}) error {
	jsondata, err := yaml.YAMLToJSON([]byte(yml))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(jsondata, obj); err != nil {
		return err
	}
	return nil
}

type UserGroupType string

const (
	OrgManagerGroup = "erda-org-manager"
	OrgSupportGroup = "erda-support-manager"
)

var (
	UserGroups = map[string]UserGroupInfo{
		"erda-org-manager": {"erda-org-manager", "default"},
		"erda-org-support": {"erda-org-support", "default"},
	}
)

type UserGroupInfo struct {
	ServiceAccountName      string
	ServiceAccountNamespace string
}

var (
	ServiceAccountExpression = `
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: erda-org-manager
  namespace: default
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: erda-org-support
  namespace: default
`
	ClusterRoleExpression = `
---
---
`
	ClusterRoleBindingExpression = `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: erda-readonly
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view
subjects:
- kind: Group
  name: erda-org-support
- kind: ServiceAccount
  name: erda-org-support
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: erda-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: Group
  name: erda-org-manager
- kind: ServiceAccount
  name: erda-org-manager
  namespace: default
`
)
