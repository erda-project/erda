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

type maven struct{}

func (m maven) GetFormat() RepositoryFormat { return RepositoryFormatMaven }

// hosted
type MavenHostedRepositoryCreateRequest struct {
	maven
	HostedRepositoryCreateRequest
	Maven RepositoryMavenConfig `json:"maven,omitempty"`
}
type MavenHostedRepositoryUpdateRequest MavenHostedRepositoryCreateRequest

// proxy
type MavenProxyRepositoryCreateRequest struct {
	maven
	ProxyRepositoryCreateRequest
	Maven RepositoryMavenConfig `json:"maven,omitempty"`
}
type MavenProxyRepositoryUpdateRequest MavenProxyRepositoryCreateRequest

// group
type MavenGroupRepositoryCreateRequest struct {
	maven
	GroupRepositoryCreateRequest
}
type MavenGroupRepositoryUpdateRequest MavenGroupRepositoryCreateRequest
