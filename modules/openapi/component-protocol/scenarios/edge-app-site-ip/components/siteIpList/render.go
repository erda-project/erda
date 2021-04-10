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

package siteiplist

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

const (
	statusStarting  = "Starting" // 已启动，但未收到健康检查事件，瞬态
	statusRunning   = "Running"
	statusHealthy   = "Healthy"
	statusUnHealthy = "UnHealthy" // 已启动但收到未通过健康检查事件
	statusFinished  = "Finished"  // 已完成，退出码为0
	statusFailed    = "Failed"    // 已退出，退出码非0
	statusKilled    = "Killed"    // 已被杀
	statusStopped   = "Stopped"   // 已停止，Scheduler与DCOS断连期间事件丢失，后续补偿时，需将Healthy置为Stopped
	statusUnknown   = "Unknown"
	statusOOM       = "OOM"
	statusDead      = "Dead"
)

type EdgeSiteMachineItem struct {
	ID        int64                         `json:"id"`
	IP        string                        `json:"ip"`
	Address   string                        `json:"address"`
	Status    apistructs.EdgeTextBadge      `json:"status"`
	CreatedAt string                        `json:"createdAt"`
	Operate   apistructs.EdgeItemOperations `json:"operate"`
}

func (c *ComponentList) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}

	if err := c.SetComponent(component); err != nil {
		return err
	}

	if c.component.State == nil {
		c.component.State = make(map[string]interface{})
	}

	orgID, err := strconv.ParseInt(c.ctxBundle.Identity.OrgID, 10, 64)
	if err != nil {
		return fmt.Errorf("component %s parse org id error: %v", component.Name, err)
	}

	identity := c.ctxBundle.Identity

	if strings.HasPrefix(event.Operation.String(), apistructs.EdgeOperationChangePage) ||
		event.Operation == apistructs.InitializeOperation ||
		event.Operation == apistructs.RenderingOperation {
		err = c.OperateChangePage(orgID, identity)
		if err != nil {
			return err
		}
	}

	c.component.Props = getProps()
	c.component.Operations = getOperations()

	return nil
}

func getOperations() apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		"changePageNo": apistructs.EdgeOperation{
			Key:    "changePageNo",
			Reload: true,
		},
		"changePageSize": apistructs.EdgeOperation{
			Key:    "changePageSize",
			Reload: true,
		},
	}
}

func getProps() apistructs.EdgeTableProps {
	return apistructs.EdgeTableProps{
		PageSizeOptions: []string{"10", "20", "50", "100"},
		RowKey:          "id",
		Columns: []apistructs.EdgeColumns{
			{Title: "实例IP", DataIndex: "ip", Width: 150},
			{Title: "主机地址", DataIndex: "address", Width: 150},
			{Title: "状态", DataIndex: "status", Width: 150},
			{Title: "创建时间", DataIndex: "createdAt", Width: 150},
			{Title: "操作", DataIndex: "operate", Width: 150},
		},
	}
}

func getItemOperations(containerID, ip, clusterName string) apistructs.EdgeItemOperations {
	return apistructs.EdgeItemOperations{
		RenderType: "tableOperation",
		Operations: map[string]apistructs.EdgeItemOperation{
			"viewTerminal": {
				Key:    "viewTerminal",
				Text:   "控制台",
				Reload: false,
				Meta: map[string]interface{}{
					"clusterName": clusterName,
					"instance": map[string]interface{}{
						"id":          containerID,
						"containerId": containerID,
						"clusterName": clusterName,
						"hostIP":      ip,
					},
					"type": "terminal",
				},
			},
			"viewMonitor": {
				Key:    "viewMonitor",
				Text:   "容器监控",
				Reload: false,
				Meta: map[string]interface{}{
					"api": "/api/orgCenter/metrics",
					"instance": map[string]interface{}{
						"id":          containerID,
						"containerId": containerID,
						"clusterName": clusterName,
						"hostIP":      ip,
					},
					"type": "monitor",
					"extraQuery": map[string]interface{}{
						"filter_cluster_name": clusterName,
					},
				},
			},
			"viewLog": {
				Key:    "viewLog",
				Text:   "日志",
				Reload: false,
				Meta: map[string]interface{}{
					"fetchApi": "/api/orgCenter/logs",
					"instance": map[string]interface{}{
						"id":          containerID,
						"containerId": containerID,
						"clusterName": clusterName,
						"hostIP":      ip,
					},
					"extraQuery": map[string]interface{}{
						"clusterName": clusterName,
					},
					"sourceType": "container",
					"type":       "log",
				},
			},
		},
	}
}

func getStatus(phase string) apistructs.EdgeTextBadge {
	status := apistructs.EdgeTextBadge{
		RenderType: "textWithBadge",
	}

	switch phase {
	case statusRunning, statusHealthy:
		status.Value = phase
		status.Status = "success"
	case statusStarting, statusUnHealthy, statusUnknown:
		status.Value = phase
		status.Status = "warning"
	case statusOOM, statusFinished, statusFailed, statusStopped, statusKilled, statusDead:
		status.Value = phase
		status.Status = "error"
	}
	return status
}

func GetEdgeApplicationContainerStatus(phase string) string {
	switch phase {
	// Is running
	case statusRunning, statusHealthy, statusStarting, statusUnHealthy, statusUnknown:
		return "success"
	// Stopped
	case statusOOM, statusFinished, statusFailed, statusStopped, statusKilled, statusDead:
		return "error"
	default:
		return "error"
	}
}
