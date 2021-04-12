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

package appconfigform

import (
	"context"
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

const (
	MIRROR     = "MIRROR"
	MIDDLEWARE = "MIDDLEWARE"

	AppNameMatchPattern = "^[a-z][a-z0-9-]*[a-z0-9]$"
	AppNameRegexpError  = "可输入小写字母、数字、中划线; 必须以小写字母开头, 以小写字母或数字结尾"
)

var (
	AppNameMatchRegexp = fmt.Sprintf("/%v/", AppNameMatchPattern)
)

type Application struct {
	ID                uint64                       `json:"id"`
	AppName           string                       `json:"appName"`
	DeployResource    string                       `json:"deployResource"`
	MiddlewareType    string                       `json:"middlewareType"`
	Cluster           string                       `json:"cluster"`
	Sites             []string                     `json:"sites,omitempty"`
	ConfigSet         string                       `json:"configSets"`
	Depends           []string                     `json:"depends,omitempty"`
	Replicaset        int32                        `json:"copyNum"`
	Storage           ApplicationStorage           `json:"storage"`
	Mirror            ApplicationMirror            `json:"mirror"`
	HealthCheckConfig ApplicationHealthCheckConfig `json:"healthCheckConfig"`
	PortMap           ApplicationPortMap           `json:"portMap"`
}

type ApplicationStorage struct {
	CpuRequest    float64 `json:"cpuRequest"`
	CpuLimit      float64 `json:"cpuLimit"`
	MemoryRequest float64 `json:"memoryRequest"`
	MemoryLimit   float64 `json:"memoryLimit"`
}

type ApplicationMirror struct {
	MirrorAddress    string `json:"mirrorAddress"`
	RegistryUser     string `json:"registryUser"`
	RegistryPassword string `json:"registryPassword"`
}

type ApplicationHealthCheckConfig struct {
	HealthCheckType string `json:"healthCheckType"`
	Path            string `json:"path"`
	Port            int    `json:"port"`
	HealthCheckExec string `json:"HealthCheckExec"`
}

type ApplicationPortMap struct {
	Data []ApplicationPortMapData `json:"data"`
}

type ApplicationPortMapData struct {
	Protocol  string `json:"protocol"`
	Container int    `json:"container"`
	Service   int32  `json:"service"`
}

type ChangeClusterReqForm struct {
	ClusterName string `json:"cluster"`
}

type ChangeClusterReqProps struct {
	Fields []*apistructs.EdgeFormModalField `json:"fields"`
}

func (c ComponentFormModal) Render(ctx context.Context, component *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)

	if err := c.SetBundle(bdl); err != nil {
		return err
	}

	if err := c.SetComponent(component); err != nil {
		return err
	}

	orgID, err := strconv.ParseInt(c.ctxBundle.Identity.OrgID, 10, 64)
	if err != nil {
		return fmt.Errorf("component %s parse org id error: %v", component.Name, err)
	}

	identity := c.ctxBundle.Identity

	if c.component.State == nil {
		c.component.State = make(map[string]interface{})
	}

	switch event.Operation {
	case apistructs.EdgeOperationSubmit:
		if err = c.OperateSubmit(orgID, identity); err != nil {
			return fmt.Errorf("operation submit error: %v", err)
		}
		break
	case apistructs.EdgeOperationCgCluster:
		if err = c.OperateChangeCluster(orgID, identity); err != nil {
			return fmt.Errorf("operation change cluster error: %v", err)
		}
		break
	case apistructs.RenderingOperation:
		if err = c.OperateRendering(orgID, identity); err != nil {
			return fmt.Errorf("rendering operation error: %v", err)
		}
		break
	default:
		c.component.Props = getProps(nil, nil, nil, nil, "", "")
	}

	c.component.Operations = getOperations()

	return nil
}

func getOperations() apistructs.EdgeOperations {
	return apistructs.EdgeOperations{
		"submit": apistructs.EdgeOperation{
			Key:    "submit",
			Reload: true,
		},
		"cancel": apistructs.EdgeOperation{
			Reload: false,
			Key:    "cancel",
			Command: apistructs.EdgeJumpCommand{
				Key:    "set",
				Target: "addAppDrawer",
				State: apistructs.EdgeJumpCommandState{
					Visible: false,
				},
			},
		},
	}
}

// Operation:
// - viewDetail->all disabled;
// - update: deployResource:
//           - mirror disabled: appName, deployResource, cluster;
//           - middleware disabled: appName, deployResource, middlewareType, cluster;
// Param: cluster, sites, configSets； select part.
func getProps(clusters, sites, configSets, depends []map[string]interface{}, operation, deployResource string) apistructs.EdgeFormModalPointProps {
	formModal := apistructs.EdgeFormModalPointProps{
		Fields:      []*apistructs.EdgeFormModalField{},
		FooterAlign: "right",
	}

	appNameField := apistructs.EdgeFormModalField{
		Key:       "appName",
		Label:     "应用名",
		Component: "input",
		Required:  true,
		Rules: []apistructs.EdgeFormModalFieldRule{
			{
				Pattern: AppNameMatchRegexp,
				Message: AppNameRegexpError,
			},
		},
		ComponentProps: map[string]interface{}{
			"maxLength": ApplicationNameLength,
		},
	}
	deployResourceField := apistructs.EdgeFormModalField{
		Key:          "deployResource",
		Label:        "部署源",
		Component:    "select",
		Required:     true,
		DefaultValue: MIDDLEWARE,
		ComponentProps: map[string]interface{}{
			"placeholder": "请选择部署源",
			"options": []map[string]interface{}{
				{
					"name":  "镜像",
					"value": MIRROR,
				},
				{
					"name":  "中间件",
					"value": MIDDLEWARE,
				},
			},
		},
		Operations: map[string]interface{}{
			"onChange": apistructs.EdgeOperation{
				Key:    "changeDepResource",
				Reload: true,
			},
		},
	}

	middleWareTypeField := apistructs.EdgeFormModalField{
		Key:       "middlewareType",
		Label:     "中间件类型",
		Component: "select",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"placeholder": "请选择中间件类型",
			"options": []map[string]interface{}{
				{
					"name":  MysqlAddonName,
					"value": fmt.Sprintf("%s:%s", MysqlAddonName, MysqlAddonVersion),
				},
			},
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "deployResource",
					"operator": "=",
					"value":    "MIRROR",
				},
			},
		},
	}

	clusterField := apistructs.EdgeFormModalField{
		Key:       "cluster",
		Label:     "集群",
		Component: "select",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"placeholder": "请选择集群",
			"options":     clusters,
		},
		Operations: map[string]interface{}{
			"change": apistructs.EdgeOperation{
				Key:    "clusterChange",
				Reload: true,
			},
		},
	}

	sitesField := apistructs.EdgeFormModalField{
		Key:       "sites",
		Label:     "站点",
		Component: "select",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"placeholder": "请选择对应集群下的站点",
			"mode":        "multiple",
			"options":     sites,
			"selectAll":   true,
		},
	}
	cfgSets := apistructs.EdgeFormModalField{
		Key:       "configSets",
		Label:     "配置集",
		Component: "select",
		Required:  false,
		ComponentProps: map[string]interface{}{
			"placeholder": "请选择配置集",
			"allowClear":  true,
			"options":     configSets,
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "deployResource",
					"operator": "=",
					"value":    "MIDDLEWARE",
				},
			},
		},
	}
	dependApps := apistructs.EdgeFormModalField{
		Key:       "depends",
		Label:     "依赖",
		Component: "select",
		Required:  false,
		ComponentProps: map[string]interface{}{
			"placeholder": "请选择依赖",
			"mode":        "multiple",
			"allowClear":  true,
			"options":     depends,
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "deployResource",
					"operator": "=",
					"value":    "MIDDLEWARE",
				},
			},
		},
	}

	replicaField := apistructs.EdgeFormModalField{
		Key:       "copyNum",
		Label:     "副本数",
		Component: "inputNumber",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"placeholder": "请输入副本数",
			"className":   "full-width",
			"precision":   0,
			"min":         0,
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "deployResource",
					"operator": "=",
					"value":    "MIDDLEWARE",
				},
			},
		},
	}

	storageGroup := apistructs.EdgeFormModalField{
		Label:     "",
		Key:       "storage",
		Component: "formGroup",
		Group:     "storage",
		ComponentProps: map[string]interface{}{
			"indentation": true,
			"showDivider": false,
			"title":       "CPU和内存",
			"direction":   "row",
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "deployResource",
					"operator": "=",
					"value":    "MIDDLEWARE",
				},
			},
		},
	}
	cpuRequestField := apistructs.EdgeFormModalField{
		Key:       "storage.cpuRequest",
		Group:     "storage",
		Label:     "CPU需求(核)",
		Component: "inputNumber",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"placeholder": "请输入CPU需求",
			"className":   "full-width",
			"precision":   1,
		},
	}
	cpuLimitField := apistructs.EdgeFormModalField{
		Key:       "storage.cpuLimit",
		Group:     "storage",
		Label:     "CPU限制(核)",
		Component: "inputNumber",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"placeholder": "请输入CPU限制",
			"className":   "full-width",
			"precision":   1,
		},
	}
	memRequestField := apistructs.EdgeFormModalField{
		Key:       "storage.memoryRequest",
		Label:     "内存需求(MB)",
		Component: "inputNumber",
		Required:  true,
		Group:     "storage",
		ComponentProps: map[string]interface{}{
			"placeholder": "请输入内存需求",
			"className":   "full-width",
			"precision":   0,
		},
	}
	memLimitField := apistructs.EdgeFormModalField{
		Key:       "storage.memoryLimit",
		Group:     "storage",
		Label:     "内存限制(MB)",
		Component: "inputNumber",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"placeholder": "请输入内存限制",
			"className":   "full-width",
			"precision":   0,
		},
	}

	mirrorGroup := apistructs.EdgeFormModalField{
		Label:     "",
		Key:       "mirror",
		Component: "formGroup",
		Group:     "mirror",
		ComponentProps: map[string]interface{}{
			"indentation": true,
			"title":       "镜像配置",
			"direction":   "row",
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "deployResource",
					"operator": "=",
					"value":    "MIDDLEWARE",
				},
			},
		},
	}
	mirrorAddrField := apistructs.EdgeFormModalField{
		Key:       "mirror.mirrorAddress",
		Group:     "mirror",
		Label:     "镜像地址",
		Component: "input",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"maxLength": apistructs.EdgeDefaultValueMaxLength,
		},
	}
	mirrorUser := apistructs.EdgeFormModalField{
		Key:       "mirror.registryUser",
		Group:     "mirror",
		Label:     "镜像仓库用户名",
		Component: "input",
		Required:  false,
		ComponentProps: map[string]interface{}{
			"maxLength": apistructs.EdgeDefaultNameMaxLength,
		},
	}
	mirrorPassword := apistructs.EdgeFormModalField{
		Key:        "mirror.registryPassword",
		Group:      "mirror",
		Label:      "镜像仓库密码",
		Component:  "input",
		Required:   false,
		IsPassword: true,
		ComponentProps: map[string]interface{}{
			"maxLength": apistructs.EdgeDefaultNameMaxLength,
		},
	}

	healthCheckGroup := apistructs.EdgeFormModalField{
		Label:     "",
		Key:       "healthCheckConfig",
		Component: "formGroup",
		Group:     "healthCheckConfig",
		ComponentProps: map[string]interface{}{
			"indentation": true,
			"showDivider": false,
			"direction":   "row",
			"title":       "健康检查配置",
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "deployResource",
					"operator": "=",
					"value":    "MIDDLEWARE",
				},
			},
		},
	}
	healthCheckType := apistructs.EdgeFormModalField{
		Key:          "healthCheckConfig.healthCheckType",
		Group:        "healthCheckConfig",
		Label:        "检查方式",
		Component:    "select",
		Required:     true,
		DefaultValue: "HTTP",
		ComponentProps: map[string]interface{}{
			"placeholder": "请选择健康检查方式",
			"options": []map[string]interface{}{
				{
					"name":  "HTTP",
					"value": "HTTP",
				},
				{
					"name":  "COMMAND",
					"value": "COMMAND",
				},
			},
		},
	}
	healthCheckExec := apistructs.EdgeFormModalField{
		Key:       "healthCheckConfig.HealthCheckExec",
		Group:     "healthCheckConfig",
		Label:     "检查命令",
		Component: "input",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"placeholder": "请输入健康检查命令",
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "healthCheckConfig.healthCheckType",
					"operator": "=",
					"value":    "HTTP",
				},
			},
		},
	}
	healthCheckPath := apistructs.EdgeFormModalField{
		Key:       "healthCheckConfig.path",
		Group:     "healthCheckConfig",
		Label:     "检查路径",
		Component: "input",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"placeholder": "请输入健康检查路径",
			"maxLength":   apistructs.EdgeDefaultValueMaxLength,
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "healthCheckConfig.healthCheckType",
					"operator": "=",
					"value":    "COMMAND",
				},
			},
		},
	}
	healthCheckPort := apistructs.EdgeFormModalField{
		Key:       "healthCheckConfig.port",
		Group:     "healthCheckConfig",
		Label:     "端口",
		Component: "inputNumber",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"min":       1,
			"max":       65535,
			"precision": 0,
			"className": "full-width",
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "healthCheckConfig.healthCheckType",
					"operator": "=",
					"value":    "COMMAND",
				},
			},
		},
	}

	portMapGroup := apistructs.EdgeFormModalField{
		Label:     "",
		Key:       "portMap",
		Component: "formGroup",
		Group:     "portMap",
		ComponentProps: map[string]interface{}{
			"expandable":    true,
			"defaultExpand": false,
			"showDivider":   false,
			"indentation":   true,
			"title":         "端口映射",
		},
		RemoveWhen: [][]map[string]interface{}{
			{
				{
					"field":    "deployResource",
					"operator": "=",
					"value":    "MIDDLEWARE",
				},
			},
		},
	}

	if operation == apistructs.EdgeOperationAddApp {
		portMapGroup.ComponentProps["defaultExpand"] = true
	}

	portMapFieldDisabled := false

	if operation == apistructs.EdgeOperationViewDetail {
		portMapFieldDisabled = true
		formModal.ReadOnly = true
	}

	portMapDataField := apistructs.EdgeFormModalField{
		Key:       "portMap.data",
		Label:     "映射规则 (协议-容器端口-服务端口)",
		Group:     "portMap",
		Component: "arrayObj",
		LabelTip:  "请依次填写协议，容器端口和服务端口，容器端口为容器内应用程序监听的端口，服务端口建议与容器端口一致",
		Required:  true,
		ComponentProps: map[string]interface{}{
			"direction": "row",
			"disabled":  portMapFieldDisabled,
			"objItems": []map[string]interface{}{
				{
					"key":       "protocol",
					"component": "select",
					"labelTip":  "请选择协议",
					"required":  true,
					"options":   "k1:TCP",
					"componentProps": map[string]interface{}{
						"placeholder": "请选择协议",
						"disabled":    portMapFieldDisabled,
					},
				},
				{
					"key":       "container",
					"component": "inputNumber",
					"required":  true,
					"componentProps": map[string]interface{}{
						"className":   "full-width",
						"precision":   0,
						"min":         1,
						"max":         65535,
						"placeholder": "请输入容器端口",
						"disabled":    portMapFieldDisabled,
					},
				},
				{
					"key":       "service",
					"component": "inputNumber",
					"required":  true,
					"componentProps": map[string]interface{}{
						"className":   "full-width",
						"precision":   0,
						"min":         1,
						"max":         65535,
						"placeholder": "请输入服务端口",
						"disabled":    portMapFieldDisabled,
					},
				},
			},
		},
	}

	if operation == apistructs.EdgeOperationUpdate {
		appNameField.Disabled = true
		deployResourceField.Disabled = true
		clusterField.Disabled = true
		sitesField.Label = "站点 (新增的站点会部署对应的实例, 被删除的站点上应用对应的实例会被销毁)"
		cfgSets.Label = "配置集 (全站点影响)"
		dependApps.Label = "依赖 (全站点影响)"
		replicaField.Label = "副本数 (全站点影响)"
		storageGroup.Label = "CPU和内存 (全站点影响)"
		mirrorGroup.Label = "镜像配置 (全站点影响)"
		healthCheckGroup.Label = "健康检查配置 (全站点影响)"
		if deployResource == MIDDLEWARE {
			middleWareTypeField.Disabled = true
		}
	}

	formModal.Fields = append(formModal.Fields, &appNameField)
	formModal.Fields = append(formModal.Fields, &deployResourceField)

	formModal.Fields = append(formModal.Fields, &middleWareTypeField)

	formModal.Fields = append(formModal.Fields, &clusterField)
	formModal.Fields = append(formModal.Fields, &sitesField)
	formModal.Fields = append(formModal.Fields, &cfgSets)
	formModal.Fields = append(formModal.Fields, &dependApps)

	formModal.Fields = append(formModal.Fields, &replicaField)

	formModal.Fields = append(formModal.Fields, &storageGroup)
	formModal.Fields = append(formModal.Fields, &cpuRequestField)
	formModal.Fields = append(formModal.Fields, &cpuLimitField)
	formModal.Fields = append(formModal.Fields, &memRequestField)
	formModal.Fields = append(formModal.Fields, &memLimitField)

	formModal.Fields = append(formModal.Fields, &mirrorGroup)
	formModal.Fields = append(formModal.Fields, &mirrorAddrField)
	formModal.Fields = append(formModal.Fields, &mirrorUser)
	formModal.Fields = append(formModal.Fields, &mirrorPassword)

	formModal.Fields = append(formModal.Fields, &healthCheckGroup)
	formModal.Fields = append(formModal.Fields, &healthCheckType)
	formModal.Fields = append(formModal.Fields, &healthCheckExec)
	formModal.Fields = append(formModal.Fields, &healthCheckPath)
	formModal.Fields = append(formModal.Fields, &healthCheckPort)

	formModal.Fields = append(formModal.Fields, &portMapGroup)
	formModal.Fields = append(formModal.Fields, &portMapDataField)

	// All disabled.
	if operation == apistructs.EdgeOperationViewDetail {
		for _, field := range formModal.Fields {
			field.Disabled = true
		}
	} else if operation != apistructs.EdgeOperationUpdate {
		for _, field := range formModal.Fields {
			field.Disabled = false
		}
	}

	return formModal
}
