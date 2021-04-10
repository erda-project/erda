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

package dto

var INNER_HOSTS = []string{
	"dev-api-gateway.kube-system.svc.cluster.local",
	"test-api-gateway.kube-system.svc.cluster.local",
	"staging-api-gateway.kube-system.svc.cluster.local",
	"api-gateway.kube-system.svc.cluster.local",
}

type ApiDto struct {
	// 入口来源
	From string `json:"from"`
	// 路径
	Path string `json:"path"`
	// 方法
	Method string `json:"method"`
	// 转发方式
	RedirectType string `json:"redirectType"`
	// 转发地址
	RedirectAddr string `json:"redirectAddr"`
	// 转发地址
	RedirectPath string `json:"redirectPath"`
	// 域名
	Hosts []string `json:"hosts"`
	//项目名称
	ProjectId string `json:"projectId"`
	// 描述
	Description string `json:"description"`
	// 应用名
	DiceApp string `json:"diceApp"`
	// 服务名
	DiceService string `json:"diceService"`
	// 是否开启公网访问
	OuterNetEnable bool `json:"outerNetEnable"`
	// 分支id
	RuntimeId string `json:"runtimeId"`
	//环境
	Env              string      `json:"env"`
	RegisterType     string      `json:"-"`
	NeedAuth         int         `json:"-"`
	UpstreamApiId    string      `json:"-"`
	RuntimeServiceId string      `json:"-"`
	IsInner          int         `json:"-"`
	Swagger          interface{} `json:"-"`
	DaoId            string      `json:"-"`
}

type ApiReqOptionDto struct {
	// 插件信息
	Policies []string `json:"policies"`
	// consumerId
	ConsumerId string `json:"consumerId"`
}

type ApiReqDto struct {
	*ApiDto
	*ApiReqOptionDto
}

func (dto *ApiReqDto) IsEmpty() bool {
	if (len(dto.Path) == 0 && len(dto.Method) == 0) || len(dto.Env) == 0 || len(dto.ProjectId) == 0 {
		return true
	}
	if dto.RedirectPath == "" {
		dto.RedirectPath = dto.Path
	}
	switch dto.RedirectType {
	case RT_URL:
		return dto.RedirectAddr == ""
	case RT_SERVICE:
		return false
	default:
		return true
	}
}

func (dto *ApiReqDto) AddHost(host string) {
	dto.Hosts = append(dto.Hosts, host)
}
