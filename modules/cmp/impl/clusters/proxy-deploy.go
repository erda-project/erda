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

package clusters

// TODO: instead template conf load

var ProxyDeployTemplate = `
---
apiVersion: v1
kind: Namespace
metadata:
  name: {{ .ErdaNamespace }}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cluster-agent
  namespace: {{ .ErdaNamespace }}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-agent-cr
rules:
- apiGroups:
  - '*'
  resources:
  - '*'
  verbs:
  - '*'
- nonResourceURLs:
  - '*'
  verbs:
  - '*'
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: cluster-agent-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-agent-cr
subjects:
- kind: ServiceAccount
  name: cluster-agent
  namespace: {{ .ErdaNamespace }}
---
apiVersion: batch/v1
kind: Job
metadata:
  labels:
    job-name: erda-cluster-init
  name: erda-cluster-init
  namespace: {{ .ErdaNamespace }}
spec:
  backoffLimit: 0
  selector:
    matchLabels:
      job-name: erda-cluster-init
  template:
    metadata:
      labels:
        job-name: erda-cluster-init
    spec:
      serviceAccountName: cluster-agent
      restartPolicy: Never
      containers:
        - name: init
          env:
            {{- range .Envs }}
            - name: {{ .Name }}
              value: "{{ .Value }}"
            {{- end }}
          command:
            - sh
            - -c
            - /app/cluster-ops
          image: {{ .JobImage }}
          imagePullPolicy: Always
`
