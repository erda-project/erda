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

type KongServiceReqDto struct {
	// 服务名称,选填
	Name string `json:"name,omitempty"`
	// 具体路径，如果设置了这个可以不设下面4个
	Url string `json:"url,omitempty"`
	// 1、协议，默认http
	Protocol string `json:"protocol,omitempty"`
	// 2、主机，没设url则必传
	Host string `json:"host,omitempty"`
	// 3、端口，默认80
	Port int `json:"port,omitempty"`
	// 4、路径，默认null
	Path string `json:"path,omitempty"`
	// 选填，重试次数
	Retries *int `json:"retries,omitempty"`
	// 选填，连接超时时间
	ConnectTimeout int `json:"connect_timeout,omitempty"`
	// 选填，写超时时间
	WriteTimeout int `json:"write_timeout,omitempty"`
	// 选填，读超时时间
	ReadTimeout int `json:"read_timeout,omitempty"`
	// 真正的服务id，更新时使用
	ServiceId string `json:"-"`
}

// IsEmpty
func (dto KongServiceReqDto) IsEmpty() bool {
	return len(dto.Url) == 0 && len(dto.Host) == 0
}
