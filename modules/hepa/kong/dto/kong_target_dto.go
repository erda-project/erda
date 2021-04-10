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

type KongUpstreamStatusRespDto struct {
	Data []KongTargetDto `json:"data"`
}

type KongTargetDto struct {
	Id     string `json:"id,omitempty"`
	Target string `json:"target"`
	// 默认是100
	Weight     int64  `json:"weight,omitempty"`
	UpstreamId string `json:"upstream_id,omitempty"`
	CreatedAt  int64  `json:"created_at,omitempty"`
	Health     string `json:"health,omitempty"`
}
