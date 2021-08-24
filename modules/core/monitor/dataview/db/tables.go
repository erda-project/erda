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
