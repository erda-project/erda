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
