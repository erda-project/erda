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

package steve

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

var (
	predefinedServiceAccount     = getPredefinedServiceAccount()
	predefinedClusterRole        = getPredefinedClusterRole()
	predefinedClusterRoleBinding = getPredefinedClusterRoleBinding()

	erdaSystemEnv   = "ERDA_NAMESPACE"
	diceSystemEnv   = "DICE_NAMESPACE"
	systemNamespace = getSystemNamespace()

	UserGroups map[string]UserGroupInfo
)

func init() {
	UserGroups = map[string]UserGroupInfo{
		OrgManagerGroup: {
			ServiceAccountName:      "erda-org-manager",
			ServiceAccountNamespace: systemNamespace,
		},
		OrgOpsGroup: {
			ServiceAccountName:      "erda-org-ops",
			ServiceAccountNamespace: systemNamespace,
		},
		OrgSupportGroup: {
			ServiceAccountName:      "erda-org-support",
			ServiceAccountNamespace: systemNamespace,
		},
	}
}

func getSystemNamespace() string {
	ns := ""
	ns = os.Getenv(erdaSystemEnv)
	if ns == "" {
		ns = os.Getenv(diceSystemEnv)
	}
	if ns == "" {
		ns = "erda-system"
	}
	return ns
}

func getPredefinedServiceAccount() []*corev1.ServiceAccount {
	tem, err := template.New("sa").Parse(ServiceAccountExpression)
	if err != nil {
		panic(fmt.Sprintf("failed to parse predefined serviceAccounts, %v", err))
	}
	buf := bytes.Buffer{}
	if err = tem.Execute(&buf, systemNamespace); err != nil {
		panic(fmt.Sprintf("failed to execute predefined serviceAccounts template, %v", err))
	}

	yamls := strings.Split(buf.String(), "\n---\n")
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
	tem, err := template.New("cr").Parse(ClusterRoleExpression)
	if err != nil {
		panic(fmt.Sprintf("failed to parse predefined clusterRoles, %v", err))
	}
	buf := bytes.Buffer{}
	if err := tem.Execute(&buf, systemNamespace); err != nil {
		panic(fmt.Sprintf("failed to parse predefined clusterRoles template, %v", err))
	}

	yamls := strings.Split(buf.String(), "\n---\n")
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
	tem := template.New("crb")
	tem = template.Must(tem.Parse(ClusterRoleBindingExpression))
	buf := bytes.Buffer{}
	if err := tem.Execute(&buf, systemNamespace); err != nil {
		panic(fmt.Sprintf("failed to parse predefined clusterRoleBindings, %v", err))
	}

	yamls := strings.Split(buf.String(), "\n---\n")
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
	OrgOpsGroup     = "erda-org-ops"
	OrgSupportGroup = "erda-org-support"
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
  namespace: {{.}}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: erda-org-ops
  namespace: {{.}}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: erda-org-support
  namespace: {{.}}
`
	ClusterRoleExpression = `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: erda-pod-exec
rules:
- apiGroups:
  - ""
  resources:
  - 'pods/exec'
  verbs:
  - '*'
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
  namespace: {{.}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: erda-pod-exec
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: erda-pod-exec
subjects:
- kind: Group
  name: erda-org-support
- kind: ServiceAccount
  name: erda-org-support
  namespace: {{.}}
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
  name: erda-org-ops
- kind: ServiceAccount
  name: erda-org-ops
  namespace: {{.}}
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
  namespace: {{.}}
`
)
