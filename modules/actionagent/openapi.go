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
