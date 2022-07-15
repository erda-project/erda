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

package actionagent

import "github.com/erda-project/erda/apistructs"

type Interface interface {
	// Ensure ensures that the agent is available:
	// 1. When the cluster type is k8s, download through initContainer
	// 2. When the cluster type is non-k8s, call the solderer through the existing path to download
	Ensure(clusterInfo apistructs.ClusterInfoData, agentImage string, agentMD5 string) error
}
