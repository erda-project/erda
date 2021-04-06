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

package nexus

type RepositoryCreator interface {
	GetName() string
	GetOnline() bool
	GetStorage() HostedRepositoryStorageConfig
	GetCleanup() *RepositoryCleanupConfig
	GetFormat() RepositoryFormat // maven2 / npm / ...
	GetType() RepositoryType     // hosted / proxy / group
}
type RepositoryProxyCreator interface {
	RepositoryCreator
	GetProxy() RepositoryProxyConfig
	GetNegativeCache() RepositoryNegativeCacheConfig
	GetHttpClient() RepositoryHttpClientConfig
	GetRoutingRule() string
}
type RepositoryGroupCreator interface {
	RepositoryCreator
	GetGroup() RepositoryGroupConfig
}

type BaseRepositoryCreateRequest struct {
	// A unique identifier for this repository
	// pattern: ^[a-zA-Z0-9\-]{1}[a-zA-Z0-9_\-\.]*$
	// example: internal
	Name string `json:"name"`
	// Whether this repository accepts incoming requests
	Online  bool                          `json:"online"`
	Storage HostedRepositoryStorageConfig `json:"storage"`
	Cleanup *RepositoryCleanupConfig      `json:"cleanup"`
}

func (r BaseRepositoryCreateRequest) GetName() string                           { return r.Name }
func (r BaseRepositoryCreateRequest) GetOnline() bool                           { return r.Online }
func (r BaseRepositoryCreateRequest) GetStorage() HostedRepositoryStorageConfig { return r.Storage }
func (r BaseRepositoryCreateRequest) GetCleanup() *RepositoryCleanupConfig      { return r.Cleanup }

type HostedRepositoryCreateRequest struct{ BaseRepositoryCreateRequest }

func (r HostedRepositoryCreateRequest) GetType() RepositoryType { return RepositoryTypeHosted }

type ProxyRepositoryCreateRequest struct {
	BaseRepositoryCreateRequest
	Proxy         RepositoryProxyConfig         `json:"proxy"`
	NegativeCache RepositoryNegativeCacheConfig `json:"negativeCache"`
	HttpClient    RepositoryHttpClientConfig    `json:"httpClient"`
	RoutingRule   string                        `json:"routingRule"`
}

func (r ProxyRepositoryCreateRequest) GetProxy() RepositoryProxyConfig { return r.Proxy }
func (r ProxyRepositoryCreateRequest) GetNegativeCache() RepositoryNegativeCacheConfig {
	return r.NegativeCache
}
func (r ProxyRepositoryCreateRequest) GetHttpClient() RepositoryHttpClientConfig { return r.HttpClient }
func (r ProxyRepositoryCreateRequest) GetRoutingRule() string                    { return r.RoutingRule }
func (r ProxyRepositoryCreateRequest) GetType() RepositoryType                   { return RepositoryTypeProxy }

type GroupRepositoryCreateRequest struct {
	BaseRepositoryCreateRequest
	Group RepositoryGroupConfig `json:"group"`
}

func (r GroupRepositoryCreateRequest) GetGroup() RepositoryGroupConfig { return r.Group }
func (r GroupRepositoryCreateRequest) GetType() RepositoryType         { return RepositoryTypeGroup }

//////////

type RepositoryListRequest struct{}
type RepositoryGetRequest struct{ RepositoryName string }
type RepositoryDeleteRequest struct{ RepositoryName string }
type RepositoryInvalidateCacheRequest struct{ RepositoryName string }
type RepositoryRebuildIndexRequest struct{ RepositoryName string }
