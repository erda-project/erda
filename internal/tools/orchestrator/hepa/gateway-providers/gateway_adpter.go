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

package gateway_providers

import (
	. "github.com/erda-project/erda/internal/tools/orchestrator/hepa/gateway-providers/dto"
)

type GatewayAdapter interface {
	GatewayProviderExist() bool
	GetVersion() (string, error)
	CheckPluginEnabled(pluginName string) (bool, error)
	CreateConsumer(req *ConsumerReqDto) (*ConsumerRespDto, error)
	DeleteConsumer(string) error
	CreateOrUpdateRoute(req *RouteReqDto) (*RouteRespDto, error)
	DeleteRoute(string) error
	UpdateRoute(req *RouteReqDto) (*RouteRespDto, error)
	CreateOrUpdateService(req *ServiceReqDto) (*ServiceRespDto, error)
	DeleteService(string) error
	DeletePluginIfExist(req *PluginReqDto) error
	CreateOrUpdatePlugin(req *PluginReqDto) (*PluginRespDto, error)
	CreateOrUpdatePluginById(req *PluginReqDto) (*PluginRespDto, error)
	GetPlugin(req *PluginReqDto) (*PluginRespDto, error)
	AddPlugin(req *PluginReqDto) (*PluginRespDto, error)
	UpdatePlugin(req *PluginReqDto) (*PluginRespDto, error)
	PutPlugin(req *PluginReqDto) (*PluginRespDto, error)
	RemovePlugin(string) error
	CreateCredential(req *CredentialReqDto) (*CredentialDto, error)
	DeleteCredential(string, string, string) error
	GetCredentialList(string, string) (*CredentialListDto, error)
	CreateAclGroup(string, string) error
	CreateUpstream(req *UpstreamDto) (*UpstreamDto, error)
	GetUpstreamStatus(string) (*UpstreamStatusRespDto, error)
	AddUpstreamTarget(string, *TargetDto) (*TargetDto, error)
	DeleteUpstreamTarget(string, string) error
	TouchRouteOAuthMethod(string) error
	GetRoutes() ([]RouteRespDto, error)
	GetRoutesWithTag(tag string) ([]RouteRespDto, error)
}
