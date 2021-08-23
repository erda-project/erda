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

import "github.com/erda-project/erda/modules/hepa/repository/orm"

type DomainType string

const (
	ServiceDomain DomainType = "service"
	GatewayDomain DomainType = "gateway"
	OtherDomain   DomainType = "other"
)

type ManageDomainReq struct {
	Domain      string
	ClusterName string
	Type        DomainType
	ProjectID   string
	Workspace   string
}

type DomainLinkInfo struct {
	ProjectID   string `json:"projectID"`
	AppID       string `json:"appID"`
	RuntimeID   string `json:"runtimeID"`
	ServiceName string `json:"serviceName"`
	Workspace   string `json:"workspace"`
	TenantGroup string `json:"tenantGroup"`
}

type ManageDomainInfo struct {
	ID          string          `json:"id"`
	Domain      string          `json:"domain"`
	ClusterName string          `json:"clusterName"`
	Type        DomainType      `json:"type"`
	ProjectName string          `json:"projectName"`
	AppName     string          `json:"appName"`
	Workspace   string          `json:"workspace"`
	Link        *DomainLinkInfo `json:"link,omitempty"`
}

func (req ManageDomainReq) GenSelectOptions() []orm.SelectOption {
	var result []orm.SelectOption
	if req.Domain != "" {
		result = append(result, orm.SelectOption{
			Type:   orm.FuzzyMatch,
			Column: "domain",
			Value:  req.Domain,
		})
	}

	switch req.Type {
	case ServiceDomain:
		result = append(result, orm.SelectOption{
			Type:   orm.FuzzyMatch,
			Column: "type",
			Value:  "service_",
		})
	case GatewayDomain:
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "type",
			Value:  orm.DT_PACKAGE,
		})
	case OtherDomain:
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "type",
			Value:  orm.DT_COMPONENT,
		})
	}
	if req.ProjectID != "" {
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "project_id",
			Value:  req.ProjectID,
		})
	}
	if req.Workspace != "" {
		result = append(result, orm.SelectOption{
			Type:   orm.ExactMatch,
			Column: "workspace",
			Value:  req.Workspace,
		})
	}
	return result
}
