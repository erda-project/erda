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

package db

import (
	"reflect"
	"time"

	"github.com/erda-project/erda/pkg/database/gormutil"
)

// tables name
const (
	TableInstance       = "tb_tmc_instance"
	TableInstanceTenant = "tb_tmc_instance_tenant"
	TableTmc            = "tb_tmc"
)

// InstanceTenant .
type InstanceTenant struct {
	ID          string    `gorm:"column:id;primary_key"`
	InstanceID  string    `gorm:"column:instance_id"`
	Config      string    `gorm:"column:config"`
	Options     string    `gorm:"column:options"`
	TenantGroup string    `gorm:"column:tenant_group"`
	Engine      string    `gorm:"column:engine"`
	Az          string    `gorm:"column:az"`
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
	IsDeleted   string    `gorm:"column:is_deleted"`
}

// TableName .
func (InstanceTenant) TableName() string { return TableInstanceTenant }

var instanceTenantFieldColumns = gormutil.GetFieldToColumnMap(reflect.TypeOf(Instance{}))

// Instance .
type Instance struct {
	ID         string    `gorm:"column:id;primary_key"`
	Engine     string    `gorm:"column:engine"`
	Version    string    `gorm:"column:version"`
	ReleaseID  string    `gorm:"column:release_id"`
	Status     string    `gorm:"column:status"`
	Az         string    `gorm:"column:az"`
	Config     string    `gorm:"column:config"`
	Options    string    `gorm:"column:options"`
	IsCustom   string    `gorm:"column:is_custom"`
	IsDeleted  string    `gorm:"column:is_deleted"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

// TableName .
func (Instance) TableName() string { return TableInstance }

var instanceFieldColumns = gormutil.GetFieldToColumnMap(reflect.TypeOf(Instance{}))

// Tmc .
type Tmc struct {
	ID          int       `gorm:"column:id;primary_key"`
	Name        string    `gorm:"column:name"`
	Engine      string    `gorm:"column:engine"`
	ServiceType string    `gorm:"column:service_type"`
	DeployMode  string    `gorm:"column:deploy_mode"`
	IsDeleted   string    `gorm:"column:is_deleted"`
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
}

// TableName .
func (Tmc) TableName() string { return TableTmc }

var tmcFieldColumns = gormutil.GetFieldToColumnMap(reflect.TypeOf(Instance{}))
