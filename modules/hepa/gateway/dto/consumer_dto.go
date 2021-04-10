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

type EndPoint struct {
	OuterAddr string `json:"outerAddr"`
	InnerAddr string `json:"innerAddr"`
	InnerTips string `json:"innerTips"`
}

type ConsumerDto struct {
	ConsumerCredentialsDto
	Endpoint        EndPoint `json:"endpoint"`
	GatewayInstance string   `json:"gatewayInstance"`
	ClusterName     string   `json:"clusterName"`
}

type ConsumerCredentialsDto struct {
	ConsumerName string                  `json:"consumerName"`
	ConsumerId   string                  `json:"consumerId"`
	AuthConfig   *orm.ConsumerAuthConfig `json:"authConfig"`
}
