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

type RuntimeDomain struct {
	AppName      string `json:"appName"`
	Domain       string `json:"domain"`
	DomainType   string `json:"domainType"`
	CustomDomain string `json:"customDomain"`
	RootDomain   string `json:"rootDomain"`
	UseHttps     bool   `json:"useHttps"`
	PackageId    string `json:"packageId,omitempty"`
	TenantGroup  string `json:"tenantGroup,omitempty"`
}

type SortByTypeList []RuntimeDomain

func (list SortByTypeList) Len() int      { return len(list) }
func (list SortByTypeList) Swap(i, j int) { list[i], list[j] = list[j], list[i] }
func (list SortByTypeList) Less(i, j int) bool {
	if list[i].DomainType == EDT_DEFAULT {
		return true
	}
	if list[j].DomainType == EDT_DEFAULT {
		return false
	}
	return true
}

type RuntimeDomainsDto map[string][]RuntimeDomain
