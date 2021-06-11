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

package dbclient

import (
	"github.com/erda-project/erda/pkg/database/dbengine"
)

// EdgeSite edge site model
type EdgeSite struct {
	dbengine.BaseModel
	OrgID       int64
	Name        string
	DisplayName string
	Description string
	Logo        string
	ClusterID   int64
	Status      int64
}

func (EdgeSite) TableName() string {
	return "edge_sites"
}

// EdgeConfigSet edge config set,  union key: clusterName and name
type EdgeConfigSet struct {
	dbengine.BaseModel
	OrgID       int64
	ClusterID   int64
	Name        string
	DisplayName string
	Description string
}

func (EdgeConfigSet) TableName() string {
	return "edge_configsets"
}

// EdgeConfigSetItem edge config data model, union key: clusterID, siteID, configKey
type EdgeConfigSetItem struct {
	dbengine.BaseModel
	ConfigsetID int64
	Scope       string
	SiteID      int64
	ItemKey     string
	ItemValue   string
}

func (EdgeConfigSetItem) TableName() string {
	return "edge_configsets_item"
}

// EdgeApp edge app model
type EdgeApp struct {
	dbengine.BaseModel
	OrgID               int64
	Name                string
	ClusterID           int64
	Type                string
	Image               string
	ProductID           int64
	AddonName           string
	AddonVersion        string
	RegistryAddr        string
	RegistryUser        string
	RegistryPassword    string
	HealthCheckType     string
	HealthCheckHttpPort int
	HealthCheckHttpPath string
	HealthCheckExec     string
	ConfigSetName       string
	Replicas            int32
	Description         string
	EdgeSites           string
	DependApp           string
	LimitCpu            float64
	RequestCpu          float64
	LimitMem            float64
	RequestMem          float64
	PortMaps            string
	ExtraData           string
}

func (EdgeApp) TableName() string {
	return "edge_apps"
}
