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

package vars

const (
	LabelAppKubernetesComponent = "app.kubernetes.io/component"
	LabelAppKubernetesName      = "app.kubernetes.io/name"
)

const (
	LabelMcpErdaCloudName          = "mcp.erda.cloud/name"
	LabelMcpErdaCloudVersion       = "mcp.erda.cloud/version"
	LabelMcpErdaCloudIsPublished   = "mcp.erda.cloud/is-published"
	LabelMcpErdaCloudIsDefault     = "mcp.erda.cloud/is-default"
	LabelMcpErdaCloudTransportType = "mcp.erda.cloud/transport-type"
	LabelMcpErdaCloudServicePort   = "mcp.erda.cloud/service-port"
)

const (
	AnnotationMcpErdaCloudDescription = "mcp.erda.cloud/description"
	AnnotationMcpErdaCloudConnectUrl  = "mcp.erda.cloud/connect-url"
)

const (
	LabelDiceOrg = "DICE_ORG_ID"
)
