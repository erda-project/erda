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

// table names
var (
	TableSystemView = "sp_dashboard_block_system"
	TableCustomView = "sp_dashboard_block"
)

// systemBlock .
type SystemView struct {
	ID         string    `gorm:"column:id" json:"id"`
	Name       string    `gorm:"column:name" json:"name"`
	Version    string    `gorm:"column:version" json:"version"`
	Desc       string    `gorm:"column:desc" json:"desc"`
	Scope      string    `gorm:"column:scope" json:"scope"`
	ScopeID    string    `gorm:"column:scope_id" json:"scopeId"`
	ViewConfig string    `gorm:"column:view_config" json:"viewConfig"`
	DataConfig string    `gorm:"column:data_config" json:"dataConfig"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName .
func (SystemView) TableName() string { return TableSystemView }

var systemViewFieldColumns = gormutil.GetFieldToColumnMap(reflect.TypeOf(SystemView{}))

// userBlock .
type CustomView struct {
	ID         string    `gorm:"column:id" json:"id"`
	Name       string    `gorm:"column:name" json:"name"`
	Version    string    `gorm:"column:version" json:"version"`
	Desc       string    `gorm:"column:desc" json:"desc"`
	Scope      string    `gorm:"column:scope" json:"scope"`
	ScopeID    string    `gorm:"column:scope_id" json:"scopeId"`
	ViewConfig string    `gorm:"column:view_config" json:"viewConfig"`
	DataConfig string    `gorm:"column:data_config" json:"dataConfig"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

// TableName .
func (CustomView) TableName() string { return TableCustomView }

var customViewFieldColumns = gormutil.GetFieldToColumnMap(reflect.TypeOf(CustomView{}))
