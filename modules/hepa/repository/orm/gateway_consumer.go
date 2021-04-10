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

const (
	APIM_CLIENT_CONSUMER string = "apim_client"
)

type GatewayConsumer struct {
	ClientId      string `json:"client_id" xorm:"not null default '' comment('对应的客户端id') unique VARCHAR(32)"`
	ConsumerId    string `json:"consumer_id" xorm:"not null default '' comment('消费者id') unique VARCHAR(128)"`
	ConsumerName  string `json:"consumer_name" xorm:"not null default '' comment('消费者名称') VARCHAR(128)"`
	Config        string `json:"config" xorm:"comment('配置信息，存放key等') VARCHAR(1024)"`
	Endpoint      string `json:"endpoint" xorm:"not null default '' comment('终端') VARCHAR(256)"`
	OrgId         string `json:"org_id" xorm:"not null VARCHAR(32)"`
	ProjectId     string `json:"project_id" xorm:"not null VARCHAR(32)"`
	Env           string `json:"env" xorm:"not null VARCHAR(32)"`
	Az            string `json:"az" xorm:"not null VARCHAR(32)"`
	Description   string `json:"description" xorm:"comment('备注') VARCHAR(256)"`
	Type          string `json:"type" xorm:"not null default 'project' comment('调用方类型') VARCHAR(16)"`
	CloudapiAppId string `json:"cloudapi_app_id" xorm:"not null default '' comment('阿里云API网关的app id') VARCHAR(128)"`
	BaseRow       `xorm:"extends"`
}
