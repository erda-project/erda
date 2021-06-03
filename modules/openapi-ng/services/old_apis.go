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

package services

import (
	"strings"

	"github.com/erda-project/erda/modules/openapi-ng/api"
	openapis "github.com/erda-project/erda/modules/openapi/api"
)

// RegisterAPIs register all old apis
func RegisterOldAPIs(add func(spec *api.Spec)) {
	for _, a := range openapis.API {
		var attrs map[string]interface{}
		if a.CheckLogin || a.TryCheckLogin || a.CheckToken || a.CheckBasicAuth {
			var authers []string
			if a.CheckLogin || a.TryCheckLogin {
				authers = append(authers, "session")
			}
			if a.CheckBasicAuth || a.TryCheckLogin {
				authers = append(authers, "basic")
			}
			if a.CheckToken || a.TryCheckLogin {
				authers = append(authers, "token")
			}
			attrs = map[string]interface{}{
				"authers": authers,
			}
			if a.TryCheckLogin {
				attrs["try_auth"] = true
			}
		}
		add(&api.Spec{
			Method:      a.Method,
			Path:        replacePath(a.Path.String()),
			BackendPath: replacePath(a.BackendPath.String()),
			Service:     getServiceName(a.Host),
			Handler:     a.Custom,
			Attributes:  attrs,
		})
	}
}

func replacePath(path string) string {
	path = strings.Replace(path, "<", "{", -1)
	path = strings.Replace(path, ">", "}", -1)
	path = strings.Replace(path, "{*}", "{_}", -1)
	path = strings.Replace(path, " ", "_", -1)
	return path
}

func getServiceName(host string) string {
	// gittar-adaptor.marathon.l4lb.thisdcos.directory:1086
	idx := strings.Index(host, ".")
	if idx <= 0 {
		return ""
	}
	return host[:idx]
}

// type Spec struct {
// 	Path        *Path
// 	BackendPath *Path

// 	Scheme      Scheme
// 	Method      string
// 	Custom      func(rw http.ResponseWriter, req *http.Request)

// 	Host        string

// 	Audit func(*AuditContext) error

// 	NeedDesensitize bool // 是否需要对返回的 userinfo 进行脱敏处理，id也会被脱敏

// 	CheckLogin     bool
// 	TryCheckLogin  bool // 和CheckLogin区别为如果不登录也会通过,只是没有user-id
// 	CheckToken     bool
// 	CheckBasicAuth bool

// 	ChunkAPI bool

// 	// CustomResponse  func(*http.Response) error
// 	// `Host` 是API原始配置
// 	// 分别转化为 marathon 和 k8s 的host
// 	MarathonHost string
// 	K8SHost      string
// 	Port         int
// }
