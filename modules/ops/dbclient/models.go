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
	"time"

	"github.com/erda-project/erda/pkg/dbengine"
)

type RecordType string

const (
	RecordTypeAddNodes               RecordType = "addNodes"
	RecordTypeAddEssNodes            RecordType = "addEssNodes"
	RecordTypeAddAliNodes            RecordType = "addAliNodes"
	RecordTypeRmNodes                RecordType = "rmNodes"
	RecordTypeDeleteNodes            RecordType = "deleteNodes"
	RecordTypeDeleteEssNodes         RecordType = "deleteEssNodes"
	RecordTypeDeleteEssNodesCronJob  RecordType = "deleteEssNodesCronJob"
	RecordTypeSetLabels              RecordType = "setLabels"
	RecordTypeAddAliECSECluster      RecordType = "addAliECSEdgeCluster"
	RecordTypeAddAliACKECluster      RecordType = "addAliACKEdgeCluster" // TODO remove
	RecordTypeAddAliCSECluster       RecordType = "addAliCSEdgeCluster"
	RecordTypeAddAliCSManagedCluster RecordType = "addAliCSManagedEdgeCluster"
	RecordTypeUpgradeEdgeCluster     RecordType = "upgradeEdgeCluster"
	RecordTypeOfflineEdgeCluster     RecordType = "offlineEdgeCluster"
	RecordTypeCreateAliCloudMysql    RecordType = "createAliCloudMysql"
	RecordTypeCreateAliCloudMysqlDB  RecordType = "createAliCloudMysqlDB"
	RecordTypeCreateAliCloudRedis    RecordType = "createAliCloudRedis"
	RecordTypeCreateAliCloudOss      RecordType = "createAliCloudOss"
	RecordTypeCreateAliCloudOns      RecordType = "createAliCloudOns"
	RecordTypeCreateAliCloudOnsTopic RecordType = "createAliCloudOnsTopic"
	RecordTypeCreateAliCloudGateway  RecordType = "createAliCloudGateway"
)

func (r RecordType) String() string {
	return string(r)
}

type StatusType string

const (
	StatusTypeSuccess    StatusType = "success"
	StatusTypeFailed     StatusType = "failed"
	StatusTypeProcessing StatusType = "processing"
	StatusTypeUnknown    StatusType = "unknown"
)

type Record struct {
	dbengine.BaseModel
	RecordType  RecordType `gorm:"type:varchar(64)"`
	UserID      string     `gorm:"type:varchar(64)"`
	OrgID       string     `gorm:"type:varchar(64);index"`
	ClusterName string     `gorm:"type:varchar(64);index"`
	Status      StatusType `gorm:"type:varchar(64)"`
	Detail      string     `gorm:"type:text"`

	PipelineID uint64
}

func (Record) TableName() string {
	return "ops_record"
}

type VendorType string

const (
	VendorTypeAliyun VendorType = "aliyun"
)

type OrgAK struct {
	dbengine.BaseModel
	OrgID       string     `gorm:"type:varchar(64);index"`
	Vendor      VendorType `gorm:"type:varchar(64);index"`
	AccessKey   string     `gorm:"type:text"`
	SecretKey   string     `gorm:"type:text"`
	Description string     `gorm:"type:text"`
}

func (OrgAK) TableName() string {
	return "ops_orgak"
}

type ResourceType string

const (
	ResourceTypeMysql           ResourceType = "MYSQL"
	ResourceTypeMysqlDB         ResourceType = "MYSQL_DB"
	ResourceTypeGateway         ResourceType = "GATEWAY"
	ResourceTypeGatewayVpcGrant ResourceType = "GATEWAY_VPC_GRANT"
	ResourceTypeOns             ResourceType = "ONS"
	ResourceTypeOnsTopic        ResourceType = "ONS_TOPIC"
	ResourceTypeRedis           ResourceType = "REDIS"
	ResourceTypeOss             ResourceType = "OSS"
)

func (r ResourceType) String() string {
	return string(r)
}

type RoutingStatus string

const (
	ResourceStatusCreated  RoutingStatus = "CREATED"
	ResourceStatusDeleted  RoutingStatus = "DELETED"
	ResourceStatusAttached RoutingStatus = "ATTACHED"
	ResourceStatusDetached RoutingStatus = "DETACHED"
)

func (r RoutingStatus) String() string {
	return string(r)
}

type ResourceRouting struct {
	dbengine.BaseModel
	ResourceID string `gorm:"type:varchar(128); index"`
	// e.g mysql instance/db name
	ResourceName string `gorm:"type:varchar(64)"`
	// e.g mysql/mysql db
	ResourceType ResourceType  `gorm:"type:varchar(32)"`
	Vendor       string        `gorm:"type:varchar(32)"`
	OrgID        string        `gorm:"type:varchar(64)"`
	ClusterName  string        `gorm:"type:varchar(64)"`
	ProjectID    string        `gorm:"type:varchar(64); index"`
	AddonID      string        `gorm:"type:varchar(64)"`
	Status       RoutingStatus `gorm:"type:varchar(32)"`
	RecordID     uint64
	Detail       string `gorm:"type:text"`
}

func (ResourceRouting) TableName() string {
	return "cloud_resource_routing"
}

// addon management
type AddonManagement struct {
	ID          uint64 `gorm:"primary_key"`
	AddonID     string `gorm:"type:varchar(64)"` // Primary key
	Name        string `gorm:"type:varchar(64)"`
	ProjectID   string
	OrgID       string
	AddonConfig string `gorm:"type:text"`
	CPU         float64
	Mem         uint64
	Nodes       int
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
}

func (AddonManagement) TableName() string {
	return "tb_addon_management"
}

// edge site model
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

// edge config set,  union key: clusterName and name
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

// edge config data model, union key: clusterID, siteID, configKey
// TODO: ugly name
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

// edge app model
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
