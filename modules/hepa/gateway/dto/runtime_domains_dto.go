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
	"github.com/erda-project/erda/pkg/strutil"
)

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
	if list[i].DomainType == list[j].DomainType {
		return list.lesByDomain(i, j)
	}
	if list[i].DomainType == EDT_DEFAULT {
		return true
	}
	if list[j].DomainType == EDT_DEFAULT {
		return false
	}
	if list[i].DomainType == EDT_CUSTOM {
		return true
	}
	if list[j].DomainType == EDT_CUSTOM {
		return false
	}
	return true
}

func (list SortByTypeList) lesByDomain(i, j int) bool {
	domainI := list[i].Domain
	domainJ := list[j].Domain
	strutil.ReverseSlice(domainI)
	strutil.ReverseSlice(domainJ)
	return domainI < domainJ
}

type RuntimeDomainsDto map[string][]RuntimeDomain
