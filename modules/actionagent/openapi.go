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

import (
	"os"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

const (
	EnvDiceOpenapiPublicUrl = "DICE_OPENAPI_PUBLIC_URL"
	EnvDiceOpenapiAddr      = "DICE_OPENAPI_ADDR"
	EnvDiceIsEdge           = "DICE_IS_EDGE"
)

var getOpenAPILock sync.Once

func (agent *Agent) getOpenAPIInfo() {
	getOpenAPILock.Do(
		func() {
			diceIsEdge := os.Getenv(EnvDiceIsEdge)
			openapiPublicURL := os.Getenv(EnvDiceOpenapiPublicUrl)
			openapiAddr := os.Getenv(EnvDiceOpenapiAddr)

			// 判断是否是边缘集群
			agent.isEdgeCluster()

			// 根据是否是边缘集群对 openapi 环境变量进行转换
			agent.convertEnvsByClusterLocation()
			openAPIAddr := os.Getenv(EnvDiceOpenapiAddr)
			if openAPIAddr == "" {
				agent.AppendError(errors.Errorf("failed to get openapi addr, %s: %s, %s: %s, %s: %s",
					EnvDiceIsEdge, diceIsEdge,
					EnvDiceOpenapiPublicUrl, openapiPublicURL,
					EnvDiceOpenapiAddr, openapiAddr))
				return
			}
			agent.EasyUse.OpenAPIAddr = openAPIAddr
		})
}

// isEdgeCluster 是否是边缘集群
func (agent *Agent) isEdgeCluster() {
	isEdgeCluster, _ := strconv.ParseBool(os.Getenv(EnvDiceIsEdge))
	agent.EasyUse.IsEdgeCluster = isEdgeCluster
}
