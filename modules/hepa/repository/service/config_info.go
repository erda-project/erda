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

package service

type ConfigItem struct {
	GatewayEndpoint string `json:"GATEWAY_ENDPOINT"`
	VipKongHost     string `json:"VIP_KONG_HOST"`
	ProxyKongPort   string `json:"PROXY_KONG_PORT"`
	AdminEndpoint   string `json:"ADMIN_ENDPOINT"`
	DiceAddon       string `json:"GATEWAY_INSTANCE_ID"`
}

type ConfigInfo struct {
	ConfigItem
	IsolationConfig ConfigItem `json:"ISOLATION_CONFIG"`
	IsolationEnvs   []string   `json:"ISOLATION_ENVS"`
	ProjectName     string     `json:"PROJECT_NAME"`
}
