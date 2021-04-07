// Package apierrors 定义了错误列表
package apierrors

import (
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

var (
	ErrCreateFoobar = err("ErrCreateFoobar", "创建失败例子")
)

var (
	ErrGetAddonConfig    = err("ErrGetAddonConfig", "获取addon配置失败")
	ErrGetAddonStatus    = err("ErrGetAddonStatus", "获取addon状态失败")
	ErrUpdateAddonConfig = err("ErrUpdateAddonConfig", "更新addon配置失败")
	ErrCreateCluster     = err("ErrCreateCluster", "创建集群失败")
)

var (
	ErrListEdgeSite         = err("ErrListEdgeSite", "查询边缘站点失败")
	ErrGetEdgeSite          = err("ErrGetEdgeSite", "获取边缘站点失败")
	ErrCreateEdgeSite       = err("ErrCreateEdgeSite", "创建边缘站点失败")
	ErrUpdateEdgeSite       = err("ErrUpdateEdgeSite", "更新边缘站点失败")
	ErrDeleteEdgeSite       = err("ErrDeleteEdgeSite", "删除边缘站点失败")
	ErrGetEdgeSiteInit      = err("ErrGetEdgeSiteInit", "获取边缘站点初始化脚本失败")
	ErrOfflineEdgeSite      = err("ErrOfflineEdgeSite", "下线边缘站点失败")
	ErrListEdgeConfigSet    = err("ErrListEdgeConfigSet", "查询边缘配置集失败")
	ErrCreateEdgeConfigSet  = err("ErrCreateEdgeConfigSet", "创建边缘配置集失败")
	ErrUpdateEdgeConfigSet  = err("ErrUpdateEdgeConfigSet", "更新边缘配置集失败")
	ErrDeleteEdgeConfigSet  = err("ErrDeleteEdgeConfigSet", "删除边缘配置集失败")
	ErrListEdgeCfgSetItem   = err("ErrListEdgeCfgSetItem", "查询边缘配置集失败")
	ErrCreateEdgeCfgSetItem = err("ErrCreateEdgeCfgSetItem", "创建边缘配置项失败")
	ErrUpdateEdgeCfgSetItem = err("ErrUpdateEdgeCfgSetItem", "更新边缘配置项失败")
	ErrDeleteEdgeCfgSetItem = err("ErrDeleteEdgeCfgSetItem", "删除边缘配置项失败")
	ErrListEdgeApp          = err("ErrListEdgeSite", "查询边缘应用失败")
	ErrCreateEdgeApp        = err("ErrCreateEdgeApp", "创建边缘应用失败")
	ErrUpdateEdgeApp        = err("ErrUpdateEdgeSite", "更新边缘应用失败")
	ErrDeleteEdgeApp        = err("ErrDeleteEdgeSite", "删除边缘应用失败")
	ErrRestartEdgeApp       = err("ErrRestartEdgeApp", "重启边缘应用失败")
	ErrOfflineEdgeAppSite   = err("ErrOfflineEdgeAppSite", "下线应用失败")
	AccessDeny              = err("ErrAccessDeny", "权限拒绝")
)

func err(template, defaultValue string) *errorresp.APIError {
	return errorresp.New(errorresp.WithTemplateMessage(template, defaultValue))
}
