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

package dto

type Service struct {
	// 绑定服务id，必填
	Id string `json:"id"`
}

type KongObj struct {
	Id string `json:"id"`
}

type KongRouteReqDto struct {
	// 协议列表，默认["http", "https"]
	Protocols []string `json:"protocols,omitempty"`
	// 以下三个参数至少需要填一个
	// 1、方法列表
	Methods []string `json:"methods,omitempty"`
	// 2、主机列表
	Hosts []string `json:"hosts,omitempty"`
	// 3、路径列表
	Paths []string `json:"paths,omitempty"`
	// 选填，当通过路径之一匹配路由时，从上游请求URL中去除匹配的前缀。
	StripPath *bool `json:"strip_path,omitempty"`
	// 选填，当通过主机域名中的一个匹配路由时，在上游请求报头中使用请求主机头。
	// 默认情况下设置为false，上游主机头将设置为服务的主机。
	PreserveHost *bool `json:"preserve_host,omitempty"`
	// 绑定服务，必填
	Service *Service `json:"service,omitempty"`
	// 正则匹配优先级,当前使用路径中/的个数
	RegexPriority int `json:"regex_priority,omitempty"`
	// 真正的路由id，更新时使用
	RouteId string `json:"-"`
}

// IsEmpty
func (dto KongRouteReqDto) IsEmpty() bool {
	return dto.Service == nil || len(dto.Service.Id) == 0 ||
		(len(dto.Methods) == 0 && len(dto.Hosts) == 0 && len(dto.Paths) == 0)
}
