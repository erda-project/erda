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

package assembler

import (
	gw "github.com/erda-project/erda/modules/hepa/gateway/dto"
	kong "github.com/erda-project/erda/modules/hepa/kong/dto"
	db "github.com/erda-project/erda/modules/hepa/repository/orm"
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
	Resp2GatewayService(*kong.KongServiceRespDto, *db.GatewayService) error
	Resp2GatewayServiceByApi(*kong.KongServiceRespDto, gw.ApiDto, string) (*db.GatewayService, error)
	Resp2GatewayRoute(*kong.KongRouteRespDto, *db.GatewayRoute) error
	Resp2GatewayRouteByAPi(*kong.KongRouteRespDto, string, string) (*db.GatewayRoute, error)
	Resp2GatewayPluginInstance(*kong.KongPluginRespDto, PluginParams) (*db.GatewayPluginInstance, error)
	BuildGatewayApi(gw.ApiDto, string, []db.GatewayPolicy, string, ...string) (*db.GatewayApi, error)
	BuildConsumerInfo(*db.GatewayConsumer) (*gw.ConsumerInfoDto, error)
	BuildConsumerApiInfo(*db.GatewayConsumerApi, *db.GatewayApi) (*gw.ConsumerApiInfoDto, error)
	BuildConsumerApiPolicyInfo(*db.GatewayPolicy) (*gw.ConsumerApiPolicyInfoDto, error)
}

type GatewayKongAssembler interface {
	BuildKongServiceReq(string, *gw.ApiDto) (*kong.KongServiceReqDto, error)
	BuildKongRouteReq(string, *gw.ApiDto, string, bool) (*kong.KongRouteReqDto, error)
	BuildKongPluginReqDto(string, *db.GatewayPolicy, string, string, string) (*kong.KongPluginReqDto, error)
}

type GatewayGroupAssembler interface {
	GroupInfo2Dto([]gw.GatewayGroupInfo) ([]gw.GwApiGroupDto, error)
}
