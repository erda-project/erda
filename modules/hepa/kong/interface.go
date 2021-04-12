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

package kong

import (
	. "github.com/erda-project/erda/modules/hepa/kong/dto"
)

type KongAdapter interface {
	KongExist() bool
	GetVersion() (string, error)
	CheckPluginEnabled(pluginName string) (bool, error)
	CreateConsumer(req *KongConsumerReqDto) (*KongConsumerRespDto, error)
	DeleteConsumer(string) error
	CreateOrUpdateRoute(req *KongRouteReqDto) (*KongRouteRespDto, error)
	DeleteRoute(string) error
	UpdateRoute(req *KongRouteReqDto) (*KongRouteRespDto, error)
	CreateOrUpdateService(req *KongServiceReqDto) (*KongServiceRespDto, error)
	DeleteService(string) error
	DeletePluginIfExist(req *KongPluginReqDto) error
	CreateOrUpdatePlugin(req *KongPluginReqDto) (*KongPluginRespDto, error)
	CreateOrUpdatePluginById(req *KongPluginReqDto) (*KongPluginRespDto, error)
	GetPlugin(req *KongPluginReqDto) (*KongPluginRespDto, error)
	AddPlugin(req *KongPluginReqDto) (*KongPluginRespDto, error)
	UpdatePlugin(req *KongPluginReqDto) (*KongPluginRespDto, error)
	PutPlugin(req *KongPluginReqDto) (*KongPluginRespDto, error)
	RemovePlugin(string) error
	CreateCredential(req *KongCredentialReqDto) (*KongCredentialDto, error)
	DeleteCredential(string, string, string) error
	GetCredentialList(string, string) (*KongCredentialListDto, error)
	CreateAclGroup(string, string) error
	CreateUpstream(req *KongUpstreamDto) (*KongUpstreamDto, error)
	GetUpstreamStatus(string) (*KongUpstreamStatusRespDto, error)
	AddUpstreamTarget(string, *KongTargetDto) (*KongTargetDto, error)
	DeleteUpstreamTarget(string, string) error
	TouchRouteOAuthMethod(string) error
	GetRoutes() ([]KongRouteRespDto, error)
}
