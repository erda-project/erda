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

package assembler

import (
	providerDto "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
	gw "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway/dto"
	db "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/orm"
)

type PluginParams struct {
	PolicyId   string
	GroupId    string
	ServiceId  string
	RouteId    string
	ConsumerId string
	ApiId      string
}

type GatewayDbAssembler interface {
	Resp2GatewayService(*providerDto.ServiceRespDto, *db.GatewayService) error
	Resp2GatewayServiceByApi(*providerDto.ServiceRespDto, gw.ApiDto, string) (*db.GatewayService, error)
	Resp2GatewayRoute(*providerDto.RouteRespDto, *db.GatewayRoute) error
	Resp2GatewayRouteByAPi(*providerDto.RouteRespDto, string, string) (*db.GatewayRoute, error)
	Resp2GatewayPluginInstance(*providerDto.PluginRespDto, PluginParams) (*db.GatewayPluginInstance, error)
	BuildGatewayApi(gw.ApiDto, string, []db.GatewayPolicy, string, ...string) (*db.GatewayApi, error)
	BuildConsumerInfo(*db.GatewayConsumer) (*gw.ConsumerInfoDto, error)
	BuildConsumerApiInfo(*db.GatewayConsumerApi, *db.GatewayApi) (*gw.ConsumerApiInfoDto, error)
	BuildConsumerApiPolicyInfo(*db.GatewayPolicy) (*gw.ConsumerApiPolicyInfoDto, error)
}

type GatewayKongAssembler interface {
	BuildKongServiceReq(string, *gw.ApiDto) (*providerDto.ServiceReqDto, error)
	BuildKongRouteReq(string, *gw.ApiDto, string, bool) (*providerDto.RouteReqDto, error)
	BuildKongPluginReqDto(string, *db.GatewayPolicy, string, string, string) (*providerDto.PluginReqDto, error)
}

type GatewayGroupAssembler interface {
	GroupInfo2Dto([]gw.GatewayGroupInfo) ([]gw.GwApiGroupDto, error)
}
