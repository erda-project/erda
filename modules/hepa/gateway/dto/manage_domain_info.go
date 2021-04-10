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
