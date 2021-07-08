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
  name: erda-system
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: erda
  namespace: erda-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: erda-crb
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: erda
  namespace: erda-system
---
apiVersion: batch/v1
kind: Job
metadata:
  labels:
    job-name: erda-cluster-init
  name: erda-cluster-init
  namespace: erda-system
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
      serviceAccountName: erda
      restartPolicy: Never
      containers:
        - name: init
          env:
            - name: "DEBUG"
              value: "true"
            - name: "ERDA_CHART_VERSION"
              value: "{{.ErdaHelmChartVersion}}"
            - name: "INSTALL_MODE"
              value: "local"
            - name: "REPO_MODE"
              value: "local"
            - name: "HELM_NAMESPACE"
              value: "default"
            - name: "ERDA_BASE_VALUES"
              value: "configmap.clustername={{.ClusterName}},configmap.domain={{.RootDomain}}"
            - name: "ERDA_VALUES"
              value: "domain={{.CustomDomain}},clusterName={{.ClusterName}},clusterDomain={{.RootDomain}}"
          command:
            - sh
            - -c
            - /app/cluster-init
          image: {{.InitJobImage}}
          imagePullPolicy: Always
`
