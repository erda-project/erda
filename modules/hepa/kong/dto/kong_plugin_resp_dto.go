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

type KongPluginRespDto struct {
	Id         string                 `json:"id"`
	ServiceId  string                 `json:"service_id"`
	RouteId    string                 `json:"route_id"`
	ConsumerId string                 `json:"consumer_id"`
	Route      *KongObj               `json:"route"`
	Service    *KongObj               `json:"service"`
	Consumer   *KongObj               `json:"consumer"`
	Name       string                 `json:"name"`
	Config     map[string]interface{} `json:"config"`
	Enabled    bool                   `json:"enabled"`
	CreatedAt  int64                  `json:"created_at"`
	PolicyId   string                 `json:"-"`
}

type KongPluginsDto struct {
	Total int64               `json:"total"`
	Data  []KongPluginRespDto `json:"data"`
}

func (dto *KongPluginRespDto) Compatiable() {
	if dto == nil {
		return
	}
	if dto.Route != nil {
		dto.RouteId = dto.Route.Id
	}
	if dto.Service != nil {
		dto.ServiceId = dto.Service.Id
	}
	if dto.Consumer != nil {
		dto.ConsumerId = dto.Consumer.Id
	}
}
