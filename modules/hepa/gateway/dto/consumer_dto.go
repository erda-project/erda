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
