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

package apis

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

// 转换成 openapi.api.Spec，方便用户写的类型
type ApiSpec struct {
	Path        string
	BackendPath string
	Method      string
	Host        string
	// 正常情况下，使用 internal/pkg/innerdomain 能解析转换出 `Host` 对应的 marathonHost 和 k8sHost,
	// 但是，当 `Host` 中的地址是老版的 marathon 内部地址，那么就无法确定 k8s地址会是什么，需要用 `K8SHost` 显式指定
	// 比如以下地址就无法转换
	// "hepa-gateway-1.hepagateway.addon-hepa-gateway.v1.runtimes.marathon.l4lb.thisdcos.directory"
	K8SHost         string
	Scheme          string
	Custom          func(rw http.ResponseWriter, req *http.Request)
	CustomResponse  func(*http.Response) error // 如果是 websocket，没意义，在 generator 里检查
	Audit           func(ctx *spec.AuditContext) error
	NeedDesensitize bool // 是否需要对返回的 userinfo 进行脱敏处理
	CheckLogin      bool
	TryCheckLogin   bool
	CheckToken      bool
	CheckBasicAuth  bool
	ChunkAPI        bool
	Doc             string
	// API 请求 & 应答 类型, 定义在 apistructs
	RequestType  interface{}
	ResponseType interface{}
	// 是否为真正的openapi，会生成2份 swagger doc， 一份是只有openapi的，另一份有所有注册的API
	IsOpenAPI bool
	// API 分类， 默认为Path的第二部分 /a/b/c -> b
	Group string
}

// Convert2AccessibleApi 直接从 openapi 定义生成 openapi oauth2 token 可访问的 api 格式
func (spec ApiSpec) Convert2AccessibleApi() apistructs.AccessibleAPI {
	return apistructs.AccessibleAPI{
		Path:   spec.Path,
		Method: spec.Method,
		Schema: spec.Scheme,
	}
}
