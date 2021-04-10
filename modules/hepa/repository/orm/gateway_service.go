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

package orm

type GatewayService struct {
	ServiceId   string `json:"service_id" xorm:"not null default '' comment('服务id') index VARCHAR(128)"`
	ServiceName string `json:"service_name" xorm:"comment('服务名称') VARCHAR(64)"`
	Url         string `json:"url" xorm:"comment('具体路径') VARCHAR(256)"`
	Protocol    string `json:"protocol" xorm:"comment('协议') VARCHAR(32)"`
	Host        string `json:"host" xorm:"comment('主机') VARCHAR(128)"`
	Port        string `json:"port" xorm:"comment('端口') VARCHAR(32)"`
	Path        string `json:"path" xorm:"comment('路径') VARCHAR(128)"`
	Config      string `json:"config" xorm:"comment('选填配置') VARCHAR(1024)"`
	ApiId       string `json:"api_id" xorm:"not null default '' comment('apiid') VARCHAR(32)"`
	BaseRow     `xorm:"extends"`
}
