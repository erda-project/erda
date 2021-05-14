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

import (
	"github.com/erda-project/erda/modules/hepa/repository/orm"
)

type GatewayGroupInfo struct {
	GatewayGroup orm.GatewayGroup    `json:"gatewayGroup"`
	Policies     []orm.GatewayPolicy `json:"policies"`
}

func (dto *GatewayGroupInfo) AddPolicy(policy *orm.GatewayPolicy) {
	dto.Policies = append(dto.Policies, *policy)
}
